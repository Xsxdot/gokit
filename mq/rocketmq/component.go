package rocketmq

import (
	"context"
	"errors"

	"github.com/xsxdot/gokit/mq"
)

// Component RocketMQ 组件
type Component struct {
	key      string
	config   Config
	producer *mq.Producer // 可选：回填 Producer
	consumer *mq.Consumer // 可选：回填 Consumer
}

// NewComponent 创建 RocketMQ 组件
func NewComponent(key string) *Component {
	return &Component{key: key}
}

// WithProducer 设置 Producer 回填目标（可选）
func (c *Component) WithProducer(entity *mq.Producer) *Component {
	c.producer = entity
	return c
}

// WithConsumer 设置 Consumer 回填目标（可选）
func (c *Component) WithConsumer(entity *mq.Consumer) *Component {
	c.consumer = entity
	return c
}

func (c *Component) Name() string      { return "rocketmq" }
func (c *Component) ConfigKey() string { return c.key }
func (c *Component) ConfigPtr() any    { return &c.config }
func (c *Component) EntityPtr() any    { return nil } // 多实体场景不使用单指针

func (c *Component) Start(ctx context.Context, cfg any) error {
	conf := cfg.(*Config)
	if c.producer != nil {
		p, err := NewProducer(conf)
		if err != nil {
			return err
		}
		*c.producer = p
	}
	if c.consumer != nil {
		cons, err := NewConsumer(conf)
		if err != nil {
			return err
		}
		*c.consumer = cons
	}
	return nil
}

func (c *Component) Stop() error {
	var errs []error
	if c.producer != nil && *c.producer != nil {
		if err := (*c.producer).Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if c.consumer != nil && *c.consumer != nil {
		if err := (*c.consumer).Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}