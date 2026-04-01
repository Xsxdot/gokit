package redis

import (
	"context"

	"github.com/redis/go-redis/v9"
	"github.com/xsxdot/gokit/mq"
)

// Component Redis MQ 组件
type Component struct {
	rdb      **redis.Client
	producer *mq.Producer // 可选：回填 Producer
	consumer *mq.Consumer // 可选：回填 Consumer
}

// NewComponent 创建 Redis MQ 组件
func NewComponent(rdb **redis.Client) *Component {
	return &Component{rdb: rdb}
}

// WithProducer 设置 Producer 回填目标（可选）
// 使用指针的指针来实现回填
func (c *Component) WithProducer(entity *mq.Producer) *Component {
	c.producer = entity
	return c
}

// WithConsumer 设置 Consumer 回填目标（可选）
func (c *Component) WithConsumer(entity *mq.Consumer) *Component {
	c.consumer = entity
	return c
}

func (c *Component) Name() string      { return "redis-mq" }
func (c *Component) ConfigKey() string { return "" }
func (c *Component) ConfigPtr() any    { return nil }
func (c *Component) EntityPtr() any    { return nil } // 多实体场景不使用单指针

func (c *Component) Start(ctx context.Context, cfg any) error {
	if c.producer != nil {
		p, err := NewProducer(*c.rdb)
		if err != nil {
			return err
		}
		*c.producer = p
	}
	if c.consumer != nil {
		cons, err := NewConsumer(*c.rdb)
		if err != nil {
			return err
		}
		*c.consumer = cons
	}
	return nil
}

func (c *Component) Stop() error {
	if c.producer != nil && *c.producer != nil {
		(*c.producer).Close()
	}
	if c.consumer != nil && *c.consumer != nil {
		(*c.consumer).Close()
	}
	return nil
}
