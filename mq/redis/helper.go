package redis

import (
	"encoding/json"

	"github.com/redis/go-redis/v9"
	"github.com/xsxdot/gokit/mq"
)

// buildRedisValues 将 Message 转换为 Redis Stream 的 key-value 对
func buildRedisValues(msg *mq.Message) map[string]interface{} {
	values := map[string]interface{}{
		"body": string(msg.Body),
	}
	if msg.Key != "" {
		values["key"] = msg.Key
	}
	if len(msg.Properties) > 0 {
		if data, err := json.Marshal(msg.Properties); err == nil {
			values["properties"] = string(data)
		}
	}
	return values
}

// parseRedisMessage 将 Redis Stream 消息解析为 Message
func parseRedisMessage(topic string, xMsg redis.XMessage) *mq.Message {
	msg := &mq.Message{
		Topic: topic,
		ID:    xMsg.ID,
	}

	if body, ok := xMsg.Values["body"]; ok {
		if s, ok := body.(string); ok {
			msg.Body = []byte(s)
		}
	}
	if key, ok := xMsg.Values["key"]; ok {
		if s, ok := key.(string); ok {
			msg.Key = s
		}
	}
	if props, ok := xMsg.Values["properties"]; ok {
		if s, ok := props.(string); ok {
			var properties map[string]string
			if err := json.Unmarshal([]byte(s), &properties); err == nil {
				msg.Properties = properties
			}
		}
	}

	return msg
}