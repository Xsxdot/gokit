package http

import (
	"context"
	"io"
	stdhttp "net/http"

	"github.com/go-resty/resty/v2"
	json "github.com/json-iterator/go"
	"github.com/tidwall/gjson"
)

// Response 标准HTTP响应对象
type Response struct {
	err       error           // 请求错误或断言错误
	raw       *resty.Response // 底层resty响应
	body      []byte          // 响应体（缓存）
	ctx       context.Context // 上下文（用于traceId等）
	gsonCache *gjson.Result   // gjson 解析缓存，避免批量断言时重复解析
}

// newResponse 创建响应对象
func newResponse(ctx context.Context, raw *resty.Response, err error) *Response {
	r := &Response{
		err: err,
		raw: raw,
		ctx: ctx,
	}
	if raw != nil {
		r.body = raw.Body()
	}
	return r
}

// Err 返回最终错误（请求错误或断言错误）
func (r *Response) Err() error {
	return r.err
}

// StatusCode 获取HTTP状态码
func (r *Response) StatusCode() int {
	if r.raw == nil {
		return 0
	}
	return r.raw.StatusCode()
}

// RawBodyStream 返回原始 HTTP 响应体流；仅在请求侧调用了 SetDoNotParseResponse(true) 时有效。
// 调用方必须在读取结束后关闭返回的 ReadCloser，否则会造成连接泄漏。
func (r *Response) RawBodyStream() io.ReadCloser {
	if r.raw == nil || r.raw.RawResponse == nil || r.raw.RawResponse.Body == nil {
		return nil
	}
	return r.raw.RawResponse.Body
}

// Header 获取响应头
func (r *Response) Header(key string) string {
	if r.raw == nil {
		return ""
	}
	return r.raw.Header().Get(key)
}

// Headers 获取所有响应头，保留多值（如多个 Set-Cookie）。
// 返回类型为 http.Header（即 map[string][]string）。
func (r *Response) Headers() stdhttp.Header {
	if r.raw == nil {
		return nil
	}
	return r.raw.Header()
}

// HeadersFlat 获取所有响应头的扁平化视图（每个 key 只保留第一个值）。
// 适用于不需要多值的简单场景。
func (r *Response) HeadersFlat() map[string]string {
	if r.raw == nil {
		return nil
	}
	result := make(map[string]string)
	for k, v := range r.raw.Header() {
		if len(v) > 0 {
			result[k] = v[0]
		}
	}
	return result
}

// Bytes 返回响应体字节数组
func (r *Response) Bytes() ([]byte, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.body, nil
}

// String 返回响应体字符串
func (r *Response) String() (string, error) {
	if r.err != nil {
		return "", r.err
	}
	return string(r.body), nil
}

// Unwrap 终结并提取原始 Body
func (r *Response) Unwrap() ([]byte, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.body, nil
}

// Gson 返回gjson.Result用于快速JSON查询。
// 解析结果会被缓存，批量断言时不会重复解析 body。
func (r *Response) Gson() gjson.Result {
	if r.err != nil || len(r.body) == 0 {
		return gjson.Result{}
	}
	if r.gsonCache == nil {
		res := gjson.ParseBytes(r.body)
		r.gsonCache = &res
	}
	return *r.gsonCache
}

// Bind 将响应体反序列化到结构体
func (r *Response) Bind(v interface{}) error {
	if r.err != nil {
		return r.err
	}
	if len(r.body) == 0 {
		return nil
	}
	err := json.Unmarshal(r.body, v)
	if err != nil {
		r.err = err
	}
	return r.err
}

// IsOK 检查是否请求成功（无错误）
func (r *Response) IsOK() bool {
	return r.err == nil
}
