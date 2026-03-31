package rocketmq

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	rocketmqproducer "github.com/apache/rocketmq-client-go/v2/producer"
	"github.com/xsxdot/gokit/mq"
)

// delayLevelDurations RocketMQ 延时级别对应的时长
// 1s 5s 10s 30s 1m 2m 3m 4m 5m 6m 7m 8m 9m 10m 20m 30m 1h 2h
var delayLevelDurations = []time.Duration{
	1 * time.Second,
	5 * time.Second,
	10 * time.Second,
	30 * time.Second,
	1 * time.Minute,
	2 * time.Minute,
	3 * time.Minute,
	4 * time.Minute,
	5 * time.Minute,
	6 * time.Minute,
	7 * time.Minute,
	8 * time.Minute,
	9 * time.Minute,
	10 * time.Minute,
	20 * time.Minute,
	30 * time.Minute,
	1 * time.Hour,
	2 * time.Hour,
}

// rocketProducer RocketMQ 生产者实现
type rocketProducer struct {
	producer rocketmq.Producer
	closed   bool
	mu       sync.Mutex
}

// 编译时检查接口实现
var _ mq.Producer = (*rocketProducer)(nil)

// NewProducer 创建 RocketMQ 生产者
func NewProducer(cfg *Config) (mq.Producer, error) {
	retryTimes := cfg.RetryTimes
	if retryTimes <= 0 {
		retryTimes = 3
	}

	groupName := cfg.GroupName
	if groupName == "" {
		groupName = "DEFAULT_PRODUCER"
	}

	opts := []rocketmqproducer.Option{
		rocketmqproducer.WithNameServer(strings.Split(cfg.NameServer, ",")),
		rocketmqproducer.WithGroupName(groupName),
		rocketmqproducer.WithRetry(retryTimes),
		rocketmqproducer.WithQueueSelector(rocketmqproducer.NewHashQueueSelector()),
	}

	if cfg.AccessKey != "" && cfg.SecretKey != "" {
		opts = append(opts, rocketmqproducer.WithCredentials(primitive.Credentials{
			AccessKey: cfg.AccessKey,
			SecretKey: cfg.SecretKey,
		}))
	}
	if cfg.Namespace != "" {
		opts = append(opts, rocketmqproducer.WithNamespace(cfg.Namespace))
	}

	p, err := rocketmq.NewProducer(opts...)
	if err != nil {
		return nil, fmt.Errorf("mq: create rocketmq producer failed: %w", err)
	}

	if err = p.Start(); err != nil {
		return nil, fmt.Errorf("mq: start rocketmq producer failed: %w", err)
	}

	return &rocketProducer{producer: p}, nil
}

// SendMessage 普通消息
func (p *rocketProducer) SendMessage(ctx context.Context, msg *mq.Message) (*mq.SendResult, error) {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return nil, mq.ErrAlreadyClosed
	}
	p.mu.Unlock()

	rMsg := buildRocketMQMessage(msg)
	result, err := p.producer.SendSync(ctx, rMsg)
	if err != nil {
		return nil, fmt.Errorf("mq: rocketmq send failed: %w", err)
	}
	return &mq.SendResult{MessageID: result.MsgID}, nil
}

// SendDelayMessage 延时消息 — 映射到最接近的延时级别
func (p *rocketProducer) SendDelayMessage(ctx context.Context, msg *mq.Message, delay time.Duration) (*mq.SendResult, error) {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return nil, mq.ErrAlreadyClosed
	}
	p.mu.Unlock()

	level := findClosestDelayLevel(delay)
	rMsg := buildRocketMQMessage(msg)
	rMsg.WithDelayTimeLevel(level)

	result, err := p.producer.SendSync(ctx, rMsg)
	if err != nil {
		return nil, fmt.Errorf("mq: rocketmq send delay message failed: %w", err)
	}
	return &mq.SendResult{MessageID: result.MsgID}, nil
}

// SendOrderMessage 顺序消息 — 使用 shardingKey 保证相同 key 路由到同一队列
func (p *rocketProducer) SendOrderMessage(ctx context.Context, msg *mq.Message, shardingKey string) (*mq.SendResult, error) {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return nil, mq.ErrAlreadyClosed
	}
	p.mu.Unlock()

	rMsg := buildRocketMQMessage(msg)
	rMsg.WithShardingKey(shardingKey)

	result, err := p.producer.SendSync(ctx, rMsg)
	if err != nil {
		return nil, fmt.Errorf("mq: rocketmq send order message failed: %w", err)
	}
	return &mq.SendResult{MessageID: result.MsgID}, nil
}

// Close 关闭 RocketMQ 生产者
func (p *rocketProducer) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return mq.ErrAlreadyClosed
	}
	p.closed = true
	return p.producer.Shutdown()
}