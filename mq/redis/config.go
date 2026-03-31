package redis

import (
	"github.com/redis/go-redis/v9"
)

// Config Redis MQ 配置
// 用户需自行提供 Redis 客户端（通过 RDBComponent 或其他方式初始化）
type Config struct {
	Client *redis.Client // Redis 客户端（必填）

	// 延时消息轮询间隔（毫秒），默认 500
	DelayPollIntervalMs int `yaml:"delay-poll-interval-ms" json:"delay-poll-interval-ms"`
	// 消费者拉取超时（毫秒），默认 5000
	BlockTimeoutMs int `yaml:"block-timeout-ms" json:"block-timeout-ms"`
	// Pending 消息重试间隔（秒），默认 30
	PendingRetryIntervalSec int `yaml:"pending-retry-interval-sec" json:"pending-retry-interval-sec"`
}