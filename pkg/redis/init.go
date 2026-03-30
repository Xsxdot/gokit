package redis

import (
	"context"
	"net"
	"strings"
	"time"

	"github.com/xsxdot/gokit/config"

	"github.com/xsxdot/gokit/pkg/proxy"

	"github.com/go-redis/cache/v9"
	"github.com/redis/go-redis/v9"
)

// InitRDB 初始化 Redis 客户端
func InitRDB(redisCfg config.RedisConfig, proxyCfg config.ProxyConfig) *redis.Client {
	if redisCfg.Mode == "single" {
		opts := &redis.Options{
			Addr:     redisCfg.Host,
			Password: redisCfg.Password,
			DB:       redisCfg.DB,
		}

		// 如果启用代理，配置自定义dialer
		if proxyCfg.Enabled {
			dialer := proxy.NewDialer(proxyCfg)
			opts.Dialer = func(ctx context.Context, network, addr string) (net.Conn, error) {
				return dialer.Dial(network, addr)
			}
		}

		return redis.NewClient(opts)
	}

	failoverOpts := &redis.FailoverOptions{
		MasterName:       "mymaster",
		SentinelAddrs:    strings.Split(redisCfg.Host, ","),
		Password:         redisCfg.Password,
		SentinelPassword: redisCfg.Password,
		DB:               redisCfg.DB,
	}

	// 如果启用代理，配置自定义dialer
	if proxyCfg.Enabled {
		dialer := proxy.NewDialer(proxyCfg)
		failoverOpts.Dialer = func(ctx context.Context, network, addr string) (net.Conn, error) {
			return dialer.Dial(network, addr)
		}
	}

	return redis.NewFailoverClient(failoverOpts)
}

// InitCache 初始化 Redis 缓存客户端
func InitCache(rdb *redis.Client) *cache.Cache {
	return cache.New(&cache.Options{
		Redis:      rdb,
		LocalCache: cache.NewTinyLFU(5, time.Minute),
	})
}