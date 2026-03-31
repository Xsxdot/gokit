package redis

import (
	"context"

	"github.com/xsxdot/gokit/config"

	"github.com/redis/go-redis/v9"
)

type RDBComponent struct {
	key      string
	config   config.RedisConfig
	proxyCfg *config.ProxyConfig
	entity   **redis.Client
}

// NewRDBComponent 创建 Redis 组件
func NewRDBComponent(key string, entity **redis.Client) *RDBComponent {
	return &RDBComponent{key: key, entity: entity}
}

// WithProxy 设置代理配置（可选）
func (c *RDBComponent) WithProxy(proxyCfg *config.ProxyConfig) *RDBComponent {
	c.proxyCfg = proxyCfg
	return c
}

func (c *RDBComponent) Name() string      { return "redis" }
func (c *RDBComponent) ConfigKey() string { return c.key }
func (c *RDBComponent) ConfigPtr() any    { return &c.config }
func (c *RDBComponent) EntityPtr() any    { return c.entity }
func (c *RDBComponent) Start(ctx context.Context, cfg any) error {
	redisCfg := cfg.(*config.RedisConfig)
	proxyCfg := config.ProxyConfig{}
	if c.proxyCfg != nil {
		proxyCfg = *c.proxyCfg
	}
	entity := InitRDB(*redisCfg, proxyCfg)
	*c.entity = entity
	return nil
}
func (c *RDBComponent) Stop() error {
	if c.entity != nil {
		return (*c.entity).Close()
	}
	return nil
}
