package rocketmq

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/xsxdot/gokit/mq"
)

type rocketSubscription struct {
	topic   string
	group   string
	mode    mq.ConsumeMode
	handler mq.Handler
}

// rocketConsumer RocketMQ 消费者实现
type rocketConsumer struct {
	cfg       *Config
	subs      []*rocketSubscription
	consumers []rocketmq.PushConsumer
	started   bool
	closed    bool
	mu        sync.Mutex
}

// 编译时检查接口实现
var _ mq.Consumer = (*rocketConsumer)(nil)

// NewConsumer 创建 RocketMQ 消费者
func NewConsumer(cfg *Config) (mq.Consumer, error) {
	return &rocketConsumer{
		cfg: cfg,
	}, nil
}

// Subscribe 注册主题订阅，需在 Start 前调用
func (c *rocketConsumer) Subscribe(topic string, group string, mode mq.ConsumeMode, handler mq.Handler) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.started {
		return mq.ErrAlreadyStarted
	}
	c.subs = append(c.subs, &rocketSubscription{
		topic:   topic,
		group:   group,
		mode:    mode,
		handler: handler,
	})
	return nil
}

// Start 启动所有 push consumer
func (c *rocketConsumer) Start() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.started {
		return mq.ErrAlreadyStarted
	}
	if len(c.subs) == 0 {
		return mq.ErrNoSubscription
	}
	c.started = true

	retryTimes := c.cfg.RetryTimes
	if retryTimes <= 0 {
		retryTimes = 3
	}

	for _, sub := range c.subs {
		// 根据消费模式选择 MessageModel
		var model consumer.MessageModel
		if sub.mode == mq.ConsumeModeBroadcast {
			model = consumer.BroadCasting
		} else {
			model = consumer.Clustering
		}

		opts := []consumer.Option{
			consumer.WithNameServer(strings.Split(c.cfg.NameServer, ",")),
			consumer.WithGroupName(sub.group),
			consumer.WithConsumerModel(model),
			consumer.WithRetry(retryTimes),
		}

		if c.cfg.AccessKey != "" && c.cfg.SecretKey != "" {
			opts = append(opts, consumer.WithCredentials(primitive.Credentials{
				AccessKey: c.cfg.AccessKey,
				SecretKey: c.cfg.SecretKey,
			}))
		}
		if c.cfg.Namespace != "" {
			opts = append(opts, consumer.WithNamespace(c.cfg.Namespace))
		}

		pc, err := rocketmq.NewPushConsumer(opts...)
		if err != nil {
			c.shutdownAll()
			return fmt.Errorf("mq: create rocketmq consumer for topic %s failed: %w", sub.topic, err)
		}

		handler := sub.handler
		err = pc.Subscribe(sub.topic, consumer.MessageSelector{
			Type:       consumer.TAG,
			Expression: "*",
		}, func(ctx context.Context, msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
			for _, ext := range msgs {
				msg := parseRocketMQMessage(ext)
				result := handler(ctx, msg)
				if result == mq.ConsumeRetry {
					return consumer.ConsumeRetryLater, nil
				}
			}
			return consumer.ConsumeSuccess, nil
		})
		if err != nil {
			c.shutdownAll()
			return fmt.Errorf("mq: subscribe topic %s failed: %w", sub.topic, err)
		}

		if err = pc.Start(); err != nil {
			c.shutdownAll()
			return fmt.Errorf("mq: start rocketmq consumer for topic %s failed: %w", sub.topic, err)
		}

		c.consumers = append(c.consumers, pc)
	}

	return nil
}

// Close 关闭所有消费者
func (c *rocketConsumer) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return mq.ErrAlreadyClosed
	}
	c.closed = true
	return c.shutdownAll()
}

func (c *rocketConsumer) shutdownAll() error {
	var lastErr error
	for _, pc := range c.consumers {
		if err := pc.Shutdown(); err != nil {
			lastErr = err
		}
	}
	c.consumers = nil
	return lastErr
}