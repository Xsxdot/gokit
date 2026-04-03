package http

import (
	"context"
	"net/url"

	json "github.com/json-iterator/go"
	"github.com/go-resty/resty/v2"
)

// Request HTTP请求构建器
type Request struct {
	restyReq   *resty.Request
	method     string
	url        string
	options    *Options
	ctx        context.Context
	queryObj   interface{} // GET/HEAD 时用于转 query 参数的对象
	marshalErr error       // JSON 序列化错误
}

// newRequest 创建请求构建器（内部使用）
func newRequest(method, rawURL string, opts *Options, client *resty.Client) *Request {
	if opts == nil {
		opts = NewOptions()
	}

	req := client.R()
	req.SetContext(context.Background())

	// 应用 options 中的 headers 和 cookies
	if opts.Headers != nil {
		req.SetHeaders(opts.Headers)
	}
	if opts.Cookies != nil {
		req.SetCookies(cookieMapToSlice(opts.Cookies))
	}

	return &Request{
		restyReq: req,
		method:   method,
		url:      rawURL,
		options:  opts,
		ctx:      context.Background(),
	}
}

// WithContext 设置上下文
func (r *Request) WithContext(ctx context.Context) *Request {
	r.ctx = ctx
	r.restyReq.SetContext(ctx)
	return r
}

// Header 设置单个请求头
func (r *Request) Header(key, value string) *Request {
	r.restyReq.SetHeader(key, value)
	return r
}

// Headers 批量设置请求头
func (r *Request) Headers(headers map[string]string) *Request {
	r.restyReq.SetHeaders(headers)
	return r
}

// SetDoNotParseResponse 为 true 时不将响应体读入内存，需通过 Response.RawBodyStream 读取并关闭。
func (r *Request) SetDoNotParseResponse(v bool) *Request {
	r.restyReq.SetDoNotParseResponse(v)
	return r
}

// Cookie 设置单个Cookie
func (r *Request) Cookie(key, value string) *Request {
	r.restyReq.SetCookie(&http_cookie{Name: key, Value: value})
	return r
}

// Cookies 批量设置Cookie
func (r *Request) Cookies(cookies map[string]string) *Request {
	r.restyReq.SetCookies(cookieMapToSlice(cookies))
	return r
}

// Body 设置请求体（字节数组）
func (r *Request) Body(body []byte) *Request {
	if isGetOrHead(r.method) {
		r.restyReq.SetQueryString(string(body))
	} else {
		r.restyReq.SetBody(body)
	}
	return r
}

// BodyString 设置请求体（字符串）
func (r *Request) BodyString(body string) *Request {
	if isGetOrHead(r.method) {
		r.restyReq.SetQueryString(body)
	} else {
		r.restyReq.SetBody(body)
	}
	return r
}

// JSON 设置请求体为JSON（自动序列化对象）
// 对于 GET/HEAD 请求，将对象转换为 query 参数追加到 URL
// 对于其他请求，序列化为 JSON 并设置为请求体
func (r *Request) JSON(obj interface{}) *Request {
	if isGetOrHead(r.method) {
		r.queryObj = obj
	} else {
		data, err := json.Marshal(obj)
		if err != nil {
			r.marshalErr = err
			return r
		}
		r.restyReq.SetBody(data).SetHeader("Content-Type", "application/json")
	}
	return r
}

// QueryParam 设置单个 query 参数
func (r *Request) QueryParam(key, value string) *Request {
	r.restyReq.SetQueryParam(key, value)
	return r
}

// QueryParams 批量设置 query 参数
func (r *Request) QueryParams(params map[string]string) *Request {
	r.restyReq.SetQueryParams(params)
	return r
}

// QueryParamsStruct 将 struct 序列化为 URL query 参数。
// 语义上比 JSON() 更清晰：明确表示「把这个对象拍平成 query string」。
// 对 GET/HEAD 等无 body 的请求推荐使用此方法。
func (r *Request) QueryParamsStruct(obj interface{}) *Request {
	r.queryObj = obj
	return r
}

// Do 执行HTTP请求
func (r *Request) Do() *Response {
	if r.marshalErr != nil {
		return newResponse(r.ctx, nil, errBuilder.New("JSON序列化失败", r.marshalErr).WithTraceID(r.ctx))
	}

	// 处理 GET/HEAD query 对象
	if isGetOrHead(r.method) && r.queryObj != nil {
		values, err := objToQueryValues(r.queryObj)
		if err != nil {
			return newResponse(r.ctx, nil, errBuilder.New("合并query参数失败", err).WithTraceID(r.ctx))
		}
		r.restyReq.SetQueryParamsFromValues(url.Values(values))
	}

	// 使用 resty.Execute() 统一分发，无需对每个 HTTP 方法单独处理
	raw, err := r.restyReq.Execute(r.method, r.url)

	if err != nil {
		return newResponse(r.ctx, nil, errBuilder.New("HTTP请求失败", err).WithTraceID(r.ctx))
	}

	resp := newResponse(r.ctx, raw, nil)

	return resp
}
