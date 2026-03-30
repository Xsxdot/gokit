package utils

import (
	"fmt"
	"time"
)

// GenerateOutTradeNo 生成商户订单号
// 格式：时间戳(14位) + Snowflake ID 后8位
// 使用 Snowflake 算法保证分布式环境下唯一性
func GenerateOutTradeNo() string {
	timestamp := time.Now().Format("20060102150405")
	id := getSnowflakeIDSuffix(8)
	return fmt.Sprintf("%s%s", timestamp, id)
}

// GenerateOutRefundNo 生成商户退款单号
// 格式：R + 时间戳(14位) + Snowflake ID 后8位
func GenerateOutRefundNo() string {
	timestamp := time.Now().Format("20060102150405")
	id := getSnowflakeIDSuffix(8)
	return fmt.Sprintf("R%s%s", timestamp, id)
}

// GenerateOrderNo 生成订单号
// 格式：ORDER + 日期(8位) + Snowflake ID 后10位
func GenerateOrderNo() string {
	date := time.Now().Format("20060102")
	id := getSnowflakeIDSuffix(10)
	return fmt.Sprintf("ORDER%s%s", date, id)
}

// getSnowflakeIDSuffix 获取 Snowflake ID 的后 n 位
func getSnowflakeIDSuffix(length int) string {
	id, err := NextIDString()
	if err != nil {
		// 降级方案：使用时间戳纳秒作为后备
		id = fmt.Sprintf("%d", time.Now().UnixNano())
	}
	if len(id) > length {
		return id[len(id)-length:]
	}
	return id
}
