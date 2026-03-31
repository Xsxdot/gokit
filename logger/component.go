package logger

import (
	"context"

	config "github.com/xsxdot/gokit/config"
)

type LoggerComponent struct {
	key    string
	config config.LogConfig
	entity **Log
}

// NewLoggerComponent 创建 Logger 组件
func NewLoggerComponent(key string, entity **Log) *LoggerComponent {
	return &LoggerComponent{key: key, entity: entity}
}

func (c *LoggerComponent) Name() string      { return "logger" }
func (c *LoggerComponent) ConfigKey() string { return c.key }
func (c *LoggerComponent) ConfigPtr() any    { return &c.config }
func (c *LoggerComponent) EntityPtr() any    { return c.entity }
func (c *LoggerComponent) Start(ctx context.Context, cfg any) error {
	conf := cfg.(*config.LogConfig)
	*c.entity = InitLogger(conf.Level)
	return nil
}
func (c *LoggerComponent) Stop() error { return nil }
