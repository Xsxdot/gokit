package http

import (
	"crypto/tls"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
)

// Client HTTP客户端
type Client struct {
	resty   *resty.Client
	options *Options
}

// NewClient 创建HTTP客户端
func NewClient(opts *Options) *Client {
	if opts == nil {
		opts = NewOptions()
	}

	rc := resty.New()
	applyOptionsToRestyClient(rc, opts)

	return &Client{resty: rc, options: opts}
}

// applyOptionsToRestyClient 将 Options 应用到 resty.Client
func applyOptionsToRestyClient(rc *resty.Client, opts *Options) {
	// 超时
	if opts.Timeout > 0 {
		rc.SetTimeout(opts.Timeout)
	} else {
		rc.SetTimeout(10 * time.Second)
	}

	// TLS
	if opts.TLS != nil {
		rc.SetTLSClientConfig(&tls.Config{
			InsecureSkipVerify: opts.TLS.InsecureSkipVerify,
		})
	}

	// 代理
	if opts.Proxy != nil && opts.Proxy.URL != "" {
		proxyURL := opts.Proxy.URL
		rc.SetProxy(proxyURL)
	}
}

// newRequestFromClient 使用此客户端创建一个 Request。
// 每次请求使用 Clone() 后的配置副本，防止并发请求共享同一个 map 导致数据竞争。
func (c *Client) newRequestFromClient(method, url string) *Request {
	return newRequest(method, url, c.options.Clone(), c.resty)
}

// OnAfterResponse 注册全局响应拦截器
func (c *Client) OnAfterResponse(m func(*resty.Client, *resty.Response) error) *Client {
	c.resty.OnAfterResponse(m)
	return c
}

// Get 创建GET请求
func (c *Client) Get(url string) *Request {
	return c.newRequestFromClient("GET", url)
}

// Post 创建POST请求
func (c *Client) Post(url string) *Request {
	return c.newRequestFromClient("POST", url)
}

// Put 创建PUT请求
func (c *Client) Put(url string) *Request {
	return c.newRequestFromClient("PUT", url)
}

// Delete 创建DELETE请求
func (c *Client) Delete(url string) *Request {
	return c.newRequestFromClient("DELETE", url)
}

// Patch 创建PATCH请求
func (c *Client) Patch(url string) *Request {
	return c.newRequestFromClient("PATCH", url)
}

// Head 创建HEAD请求
func (c *Client) Head(url string) *Request {
	return c.newRequestFromClient("HEAD", url)
}

// Options 创建OPTIONS请求
func (c *Client) Options(url string) *Request {
	return c.newRequestFromClient("OPTIONS", url)
}

// 全局默认客户端
var defaultClient = NewClient(nil)

// Get 快速GET请求（使用默认配置）
func Get(url string) *Request {
	return defaultClient.Get(url)
}

// Post 快速POST请求（使用默认配置）
func Post(url string) *Request {
	return defaultClient.Post(url)
}

// PostJSON 快速POST JSON请求（使用默认配置）
func PostJSON(url string, obj interface{}) *Request {
	return defaultClient.Post(url).JSON(obj)
}

// Put 快速PUT请求（使用默认配置）
func Put(url string) *Request {
	return defaultClient.Put(url)
}

// Delete 快速DELETE请求（使用默认配置）
func Delete(url string) *Request {
	return defaultClient.Delete(url)
}

// Patch 快速PATCH请求（使用默认配置）
func Patch(url string) *Request {
	return defaultClient.Patch(url)
}

// SetDefaultOptions 设置全局默认配置
func SetDefaultOptions(opts *Options) {
	if opts != nil {
		defaultClient = NewClient(opts)
	}
}

// cookieMapToSlice 将 map cookie 转为 []*http.Cookie
func cookieMapToSlice(m map[string]string) []*http.Cookie {
	cookies := make([]*http.Cookie, 0, len(m))
	for k, v := range m {
		cookies = append(cookies, &http.Cookie{Name: k, Value: v})
	}
	return cookies
}

// http_cookie is an alias to avoid naming conflicts with the http package itself
type http_cookie = http.Cookie
