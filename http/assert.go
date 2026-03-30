package http

import (
	"fmt"
	"strings"

	errorc "github.com/xsxdot/gokit/err"
)

var errBuilder = errorc.NewErrorBuilder("HttpClient")

// buildAssertError 统一构造断言失败错误
// keys 非空时：从响应 JSON 中提取指定 keys 的值作为错误上下文
//   - 如果至少有一个 key 存在，则只显示提取的字段
//   - 如果所有 keys 都不存在或 body 为空，则 fallback 到完整响应
//
// keys 为空时：将完整响应（status + headers + body）作为错误上下文
func (r *Response) buildAssertError(baseMsg string, keys []string) error {
	var contextMsg strings.Builder
	contextMsg.WriteString(baseMsg)

	// 附加请求信息，方便定位是哪个接口触发了断言失败
	if r.raw != nil && r.raw.Request != nil {
		contextMsg.WriteString(fmt.Sprintf(" [%s %s]", r.raw.Request.Method, r.raw.Request.URL))
	}

	shouldShowFullResponse := len(keys) == 0

	if len(keys) > 0 && len(r.body) > 0 {
		result := r.Gson()
		foundAnyKey := false
		var keyInfo strings.Builder

		for _, key := range keys {
			value := result.Get(key)
			if value.Exists() {
				foundAnyKey = true
				keyInfo.WriteString(fmt.Sprintf(", %s=%v", key, value.Value()))
			} else {
				keyInfo.WriteString(fmt.Sprintf(", %s=<not_found>", key))
			}
		}

		if foundAnyKey {
			contextMsg.WriteString(fmt.Sprintf(" [StatusCode=%d%s]", r.StatusCode(), keyInfo.String()))
		} else {
			shouldShowFullResponse = true
		}
	} else if len(keys) > 0 && len(r.body) == 0 {
		shouldShowFullResponse = true
	}

	if shouldShowFullResponse {
		contextMsg.WriteString(fmt.Sprintf("\n--- Response Context ---\nStatusCode: %d\n", r.StatusCode()))

		headers := r.HeadersFlat()
		if len(headers) > 0 {
			contextMsg.WriteString("Headers:\n")
			for k, v := range headers {
				contextMsg.WriteString(fmt.Sprintf("  %s: %s\n", k, v))
			}
		}

		contextMsg.WriteString(fmt.Sprintf("Body (length=%d):\n%s\n", len(r.body), string(r.body)))
		contextMsg.WriteString("--- End Response ---")
	}

	return errBuilder.New(contextMsg.String(), nil).WithTraceID(r.ctx)
}

// EnsureStatusCode 确保HTTP状态码等于指定值
func (r *Response) EnsureStatusCode(code int, keys ...string) *Response {
	if r.err != nil {
		return r
	}
	if r.StatusCode() != code {
		r.err = r.buildAssertError(
			fmt.Sprintf("期望状态码 %d，实际得到 %d", code, r.StatusCode()),
			keys,
		)
	}
	return r
}

// EnsureStatus2xx 确保HTTP状态码为2xx
func (r *Response) EnsureStatus2xx(keys ...string) *Response {
	if r.err != nil {
		return r
	}
	code := r.StatusCode()
	if code < 200 || code >= 300 {
		r.err = r.buildAssertError(
			fmt.Sprintf("期望2xx状态码，实际得到 %d", code),
			keys,
		)
	}
	return r
}

// EnsureContains 确保响应体包含指定字符串
func (r *Response) EnsureContains(substr string, keys ...string) *Response {
	if r.err != nil {
		return r
	}
	if !strings.Contains(string(r.body), substr) {
		r.err = r.buildAssertError(
			fmt.Sprintf("响应体不包含字符串: %s", substr),
			keys,
		)
	}
	return r
}

// EnsureNotContains 确保响应体不包含指定字符串
func (r *Response) EnsureNotContains(substr string, keys ...string) *Response {
	if r.err != nil {
		return r
	}
	if strings.Contains(string(r.body), substr) {
		r.err = r.buildAssertError(
			fmt.Sprintf("响应体不应包含字符串: %s", substr),
			keys,
		)
	}
	return r
}

// EnsureJsonStringEq 确保JSON中某个key的string值等于期望值
func (r *Response) EnsureJsonStringEq(key, expected string, keys ...string) *Response {
	if r.err != nil {
		return r
	}
	actual := r.Gson().Get(key).String()
	if actual != expected {
		r.err = r.buildAssertError(
			fmt.Sprintf("JSON路径 %s 期望值为 %s，实际得到 %s", key, expected, actual),
			keys,
		)
	}
	return r
}

// EnsureJsonStringNe 确保JSON中某个key的string值不等于指定值
func (r *Response) EnsureJsonStringNe(key, notExpected string, keys ...string) *Response {
	if r.err != nil {
		return r
	}
	actual := r.Gson().Get(key).String()
	if actual == notExpected {
		r.err = r.buildAssertError(
			fmt.Sprintf("JSON路径 %s 的值不应为 %s", key, notExpected),
			keys,
		)
	}
	return r
}

// EnsureJsonExists 确保JSON中某个key存在
func (r *Response) EnsureJsonExists(key string, keys ...string) *Response {
	if r.err != nil {
		return r
	}
	if !r.Gson().Get(key).Exists() {
		r.err = r.buildAssertError(
			fmt.Sprintf("JSON路径 %s 不存在", key),
			keys,
		)
	}
	return r
}

// EnsureJsonIntEq 确保JSON中某个key的int值等于期望值
func (r *Response) EnsureJsonIntEq(key string, expected int64, keys ...string) *Response {
	if r.err != nil {
		return r
	}
	actual := r.Gson().Get(key).Int()
	if actual != expected {
		r.err = r.buildAssertError(
			fmt.Sprintf("JSON路径 %s 期望值为 %d，实际得到 %d", key, expected, actual),
			keys,
		)
	}
	return r
}

// EnsureJsonBoolEq 确保JSON中某个key的bool值等于期望值
func (r *Response) EnsureJsonBoolEq(key string, expected bool, keys ...string) *Response {
	if r.err != nil {
		return r
	}
	actual := r.Gson().Get(key).Bool()
	if actual != expected {
		r.err = r.buildAssertError(
			fmt.Sprintf("JSON路径 %s 期望值为 %v，实际得到 %v", key, expected, actual),
			keys,
		)
	}
	return r
}
