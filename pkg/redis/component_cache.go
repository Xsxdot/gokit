package redis

import (
	"context"

	"github.com/go-redis/cache/v9"
	"github.com/redis/go-redis/v9"
)

type CacheComponent struct {
	rdb    *redis.Client
	entity *cache.Cache
}

// NewCacheComponent 创建 Cache 组件
// rdb: 依赖的 Redis 客户端（需要先启动 RDBComponent）
func NewCacheComponent(rdb *redis.Client) *CacheComponent {
	return &CacheComponent{rdb: rdb}
}

func (c *CacheComponent) Name() string           { return "cache" }
func (c *CacheComponent) ConfigKey() string      { return "" }
func (c *CacheComponent) ConfigPtr() any         { return nil }
func (c *CacheComponent) EntityPtr() any         { return c.entity }
func (c *CacheComponent) Start(ctx context.Context, cfg any) error {
	c.entity = InitCache(c.rdb)
	return nil
}
func (c *CacheComponent) Stop() error { return nil }