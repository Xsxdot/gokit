package executor

import (
	"context"
)

type ExecutorComponent struct {
	entity **Executor
}

// NewExecutorComponent 创建 Executor 组件
func NewExecutorComponent(entity **Executor) *ExecutorComponent {
	return &ExecutorComponent{entity: entity}
}

func (c *ExecutorComponent) Name() string      { return "executor" }
func (c *ExecutorComponent) ConfigKey() string { return "" }
func (c *ExecutorComponent) ConfigPtr() any    { return nil }
func (c *ExecutorComponent) EntityPtr() any    { return c.entity }
func (c *ExecutorComponent) Start(ctx context.Context, cfg any) error {
	defaultCfg := DefaultConfig()
	*c.entity = NewExecutor(defaultCfg)
	return (*c.entity).Start()
}
func (c *ExecutorComponent) Stop() error {
	if c.entity != nil {
		return (*c.entity).Stop()
	}
	return nil
}
