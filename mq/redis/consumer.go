package redis

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/xsxdot/gokit/mq"
)

const (
	defaultBlockTimeout         = 5 * time.Second
	defaultPendingRetryInterval = 30 * time.Second
)

type subscription struct {
	topic   string
	group   string
	mode    mq.ConsumeMode
	handler mq.Handler
}

// consumer Redis Stream 消费者实现
type consumer struct {
	client               *redis.Client
	blockTimeout         time.Duration
	pendingRetryInterval time.Duration
	subs                 []*subscription
	started              bool
	closed               bool
	mu                   sync.Mutex
	stopCh               chan struct{}
	wg                   sync.WaitGroup
}

// 编译时检查接口实现
var _ mq.Consumer = (*consumer)(nil)

// NewConsumer 创建 Redis 消费者
func NewConsumer(rdb *redis.Client) (mq.Consumer, error) {
	blockTimeout := defaultBlockTimeout
	pendingRetry := defaultPendingRetryInterval
	return &consumer{
		client:               rdb,
		blockTimeout:         blockTimeout,
		pendingRetryInterval: pendingRetry,
		stopCh:               make(chan struct{}),
	}, nil
}

// Subscribe 注册主题订阅，需在 Start 前调用
func (c *consumer) Subscribe(topic string, group string, mode mq.ConsumeMode, handler mq.Handler) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.started {
		return mq.ErrAlreadyStarted
	}
	c.subs = append(c.subs, &subscription{
		topic:   topic,
		group:   group,
		mode:    mode,
		handler: handler,
	})
	return nil
}

// Start 启动消费者
func (c *consumer) Start() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.started {
		return mq.ErrAlreadyStarted
	}
	if len(c.subs) == 0 {
		return mq.ErrNoSubscription
	}
	c.started = true

	// 为每个订阅创建对应的消费协程
	ctx := context.Background()
	for _, sub := range c.subs {
		if sub.mode == mq.ConsumeModeBroadcast {
			// 广播模式：不使用 consumer group，每个消费者独立消费
			c.wg.Add(1)
			go c.broadcastConsumeLoop(sub)
		} else {
			// 集群模式：使用 consumer group
			c.client.XGroupCreateMkStream(ctx, sub.topic, sub.group, "0").Err()
			c.wg.Add(2)
			go c.consumeLoop(sub)
			go c.reclaimPendingLoop(sub)
		}
	}

	return nil
}

// Close 关闭消费者
func (c *consumer) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return mq.ErrAlreadyClosed
	}
	c.closed = true
	close(c.stopCh)
	c.wg.Wait()
	// 不关闭 client，由用户自行管理
	return nil
}

// consumeLoop 消费循环：XREADGROUP 拉取新消息
func (c *consumer) consumeLoop(sub *subscription) {
	defer c.wg.Done()

	consumerName := fmt.Sprintf("consumer-%d", time.Now().UnixNano())

	for {
		select {
		case <-c.stopCh:
			return
		default:
		}

		streams, err := c.client.XReadGroup(context.Background(), &redis.XReadGroupArgs{
			Group:    sub.group,
			Consumer: consumerName,
			Streams:  []string{sub.topic, ">"},
			Count:    10,
			Block:    c.blockTimeout,
		}).Result()
		if err != nil {
			if err == redis.Nil {
				continue
			}
			// 短暂等待后重试
			select {
			case <-c.stopCh:
				return
			case <-time.After(time.Second):
				continue
			}
		}

		for _, stream := range streams {
			for _, xMsg := range stream.Messages {
				msg := parseRedisMessage(sub.topic, xMsg)
				result := sub.handler(context.Background(), msg)
				if result == mq.ConsumeSuccess {
					c.client.XAck(context.Background(), sub.topic, sub.group, xMsg.ID)
				}
				// ConsumeRetry: 不 ACK，消息将在 pending 列表中等待重试
			}
		}
	}
}

// reclaimPendingLoop 定期处理 pending 中的超时消息
func (c *consumer) reclaimPendingLoop(sub *subscription) {
	defer c.wg.Done()

	ticker := time.NewTicker(c.pendingRetryInterval)
	defer ticker.Stop()

	for {
		select {
		case <-c.stopCh:
			return
		case <-ticker.C:
			c.processPending(sub)
		}
	}
}

// processPending 处理超时的 pending 消息
func (c *consumer) processPending(sub *subscription) {
	ctx := context.Background()

	// 获取 pending 列表中的消息
	pending, err := c.client.XPendingExt(ctx, &redis.XPendingExtArgs{
		Stream: sub.topic,
		Group:  sub.group,
		Start:  "-",
		End:    "+",
		Count:  50,
		Idle:   c.pendingRetryInterval,
	}).Result()
	if err != nil || len(pending) == 0 {
		return
	}

	consumerName := fmt.Sprintf("reclaimer-%d", time.Now().UnixNano())

	// 收集需要 claim 的消息 ID
	ids := make([]string, 0, len(pending))
	for _, p := range pending {
		ids = append(ids, p.ID)
	}

	// XCLAIM 这些消息
	messages, err := c.client.XClaim(ctx, &redis.XClaimArgs{
		Stream:   sub.topic,
		Group:    sub.group,
		Consumer: consumerName,
		MinIdle:  c.pendingRetryInterval,
		Messages: ids,
	}).Result()
	if err != nil {
		return
	}

	for _, xMsg := range messages {
		msg := parseRedisMessage(sub.topic, xMsg)
		result := sub.handler(ctx, msg)
		if result == mq.ConsumeSuccess {
			c.client.XAck(ctx, sub.topic, sub.group, xMsg.ID)
		}
	}
}

// broadcastConsumeLoop 广播消费循环：XREAD 拉取消息，每个消费者独立维护消费位置
func (c *consumer) broadcastConsumeLoop(sub *subscription) {
	defer c.wg.Done()

	ctx := context.Background()
	// 消费位置存储在 Redis 中，key 格式: {topic}:broadcast:{group}:offset
	offsetKey := fmt.Sprintf("%s:broadcast:%s:offset", sub.topic, sub.group)

	// 获取上次消费位置，默认从最新消息开始
	lastID, err := c.client.Get(ctx, offsetKey).Result()
	if err != nil {
		lastID = "$" // $ 表示从最新消息开始，0 表示从头开始
	}

	for {
		select {
		case <-c.stopCh:
			return
		default:
		}

		// XREAD 读取新消息
		streams, err := c.client.XRead(ctx, &redis.XReadArgs{
			Streams: []string{sub.topic, lastID},
			Count:   10,
			Block:   c.blockTimeout,
		}).Result()
		if err != nil {
			if err == redis.Nil {
				continue
			}
			select {
			case <-c.stopCh:
				return
			case <-time.After(time.Second):
				continue
			}
		}

		for _, stream := range streams {
			for _, xMsg := range stream.Messages {
				msg := parseRedisMessage(sub.topic, xMsg)
				// 广播模式下不重试，消费失败直接跳过
				_ = sub.handler(ctx, msg)
				// 更新消费位置
				lastID = xMsg.ID
				c.client.Set(ctx, offsetKey, lastID, 0)
			}
		}
	}
}
