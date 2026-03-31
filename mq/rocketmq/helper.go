package rocketmq

import (
	"encoding/json"
	"time"

	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/xsxdot/gokit/mq"
)

// buildRocketMQMessage 构建 RocketMQ 原生消息
func buildRocketMQMessage(msg *mq.Message) *primitive.Message {
	rMsg := primitive.NewMessage(msg.Topic, msg.Body)
	if msg.Key != "" {
		rMsg.WithKeys([]string{msg.Key})
	}
	if len(msg.Properties) > 0 {
		// 将自定义属性序列化后存入 RocketMQ 属性
		if data, err := json.Marshal(msg.Properties); err == nil {
			rMsg.WithProperty("mq_properties", string(data))
		}
	}
	return rMsg
}

// parseRocketMQMessage 将 RocketMQ 消息解析为统一 Message
func parseRocketMQMessage(ext *primitive.MessageExt) *mq.Message {
	msg := &mq.Message{
		Topic: ext.Topic,
		Body:  ext.Body,
		ID:    ext.MsgId,
	}
	if keys := ext.GetKeys(); keys != "" {
		msg.Key = keys
	}

	// 尝试解析自定义属性
	if propsStr := ext.GetProperty("mq_properties"); propsStr != "" {
		var props map[string]string
		if err := json.Unmarshal([]byte(propsStr), &props); err == nil {
			msg.Properties = props
		}
	}

	return msg
}

// findClosestDelayLevel 找到最接近指定延迟时长的延时级别（1-based）
func findClosestDelayLevel(delay time.Duration) int {
	bestLevel := 1
	bestDiff := absDuration(delayLevelDurations[0] - delay)

	for i, d := range delayLevelDurations {
		diff := absDuration(d - delay)
		if diff < bestDiff {
			bestDiff = diff
			bestLevel = i + 1 // 延时级别从 1 开始
		}
	}
	return bestLevel
}

func absDuration(d time.Duration) time.Duration {
	if d < 0 {
		return -d
	}
	return d
}