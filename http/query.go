package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	jsoniter "github.com/json-iterator/go"
)

// objToQueryValues 将对象转换为 url.Values
// 支持 map[string]string、map[string]any 和任意 struct
func objToQueryValues(obj interface{}) (url.Values, error) {
	if obj == nil {
		return url.Values{}, nil
	}

	values := url.Values{}

	// 处理 map[string]string
	if m, ok := obj.(map[string]string); ok {
		for k, v := range m {
			values.Add(k, v)
		}
		return values, nil
	}

	// 处理 map[string]interface{}
	if m, ok := obj.(map[string]interface{}); ok {
		for k, v := range m {
			values.Add(k, fmt.Sprint(v))
		}
		return values, nil
	}

	// 其他类型通过 JSON marshal/unmarshal 转换
	// 使用 jsoniter 序列化，使用标准库 decoder + UseNumber() 反序列化
	// 避免大整数被 float64 吞精度（如 12345678 → "1.2345678e+07"）
	data, err := jsoniter.Marshal(obj)
	if err != nil {
		return nil, fmt.Errorf("序列化对象失败: %w", err)
	}

	var m map[string]interface{}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber() // 确保数字保留为 json.Number，而非 float64
	if err := decoder.Decode(&m); err != nil {
		return nil, fmt.Errorf("反序列化对象失败: %w", err)
	}

	for k, v := range m {
		values.Add(k, fmt.Sprint(v))
	}

	return values, nil
}

// isGetOrHead 判断是否为 GET 或 HEAD 方法
func isGetOrHead(method string) bool {
	m := strings.ToUpper(method)
	return m == "GET" || m == "HEAD"
}
