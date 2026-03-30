package utils

import (
	"strconv"
)

// ParseInt64 将字符串转换为int64类型，如果转换失败则返回默认值
func ParseInt64(s string, defaultVal int64) int64 {
	if s == "" {
		return defaultVal
	}

	val, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return defaultVal
	}

	return val
}
