package scheduler

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type SchedulerComponent struct {
	rdb    *redis.Client
	entity *Scheduler
}

// NewSchedulerComponent 创建 Scheduler 组件
// rdb: 依赖的 Redis 客户端（可选，无则不支持分布式任务）
func NewSchedulerComponent(rdb *redis.Client) *SchedulerComponent {
	return &SchedulerComponent{rdb: rdb}
}

func (c *SchedulerComponent) Name() string           { return "scheduler" }
func (c *SchedulerComponent) ConfigKey() string      { return "" }
func (c *SchedulerComponent) ConfigPtr() any         { return nil }
func (c *SchedulerComponent) EntityPtr() any         { return c.entity }
func (c *SchedulerComponent) Start(ctx context.Context, cfg any) error {
	defaultCfg := DefaultSchedulerConfig()
	if c.rdb != nil {
		c.entity = NewSchedulerWithRedis(defaultCfg, c.rdb)
	} else {
		c.entity = NewScheduler(defaultCfg)
	}
	return c.entity.Start()
}
func (c *SchedulerComponent) Stop() error {
	if c.entity != nil {
		return c.entity.Stop()
	}
	return nil
}