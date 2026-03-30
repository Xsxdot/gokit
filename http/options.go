package http

import (
	"time"
)

// Options 可复用的HTTP请求配置对象
type Options struct {
	Timeout time.Duration     // 请求超时时间
	Headers map[string]string // 默认请求头
	Cookies map[string]string // 默认Cookie
	Proxy   *ProxyConfig      // 代理配置
	TLS     *TLSConfig        // TLS配置
}

// ProxyConfig 代理配置
// 支持格式：
// - http://host:port
// - http://user:pass@host:port
// - https://user:pass@host:port
// - socks5://user:pass@host:port
type ProxyConfig struct {
	URL      string // 代理URL
	Username string // 用户名（可选，也可包含在URL中）
	Password string // 密码（可选，也可包含在URL中）
}

// TLSConfig TLS配置
type TLSConfig struct {
	InsecureSkipVerify bool // 是否跳过证书验证
}

// NewOptions 创建默认配置
func NewOptions() *Options {
	return &Options{
		Timeout: 10 * time.Second,
		Headers: make(map[string]string),
		Cookies: make(map[string]string),
	}
}

// WithTimeout 设置超时时间
func (o *Options) WithTimeout(timeout time.Duration) *Options {
	o.Timeout = timeout
	return o
}

// WithHeader 添加请求头
func (o *Options) WithHeader(key, value string) *Options {
	if o.Headers == nil {
		o.Headers = make(map[string]string)
	}
	o.Headers[key] = value
	return o
}

// WithHeaders 批量设置请求头
func (o *Options) WithHeaders(headers map[string]string) *Options {
	if o.Headers == nil {
		o.Headers = make(map[string]string)
	}
	for k, v := range headers {
		o.Headers[k] = v
	}
	return o
}

// WithCookie 添加Cookie
func (o *Options) WithCookie(key, value string) *Options {
	if o.Cookies == nil {
		o.Cookies = make(map[string]string)
	}
	o.Cookies[key] = value
	return o
}

// WithProxy 设置代理
func (o *Options) WithProxy(proxyURL string) *Options {
	o.Proxy = &ProxyConfig{URL: proxyURL}
	return o
}

// WithProxyAuth 设置带认证的代理
func (o *Options) WithProxyAuth(proxyURL, username, password string) *Options {
	o.Proxy = &ProxyConfig{
		URL:      proxyURL,
		Username: username,
		Password: password,
	}
	return o
}

// WithInsecureSkipVerify 设置是否跳过TLS证书验证
func (o *Options) WithInsecureSkipVerify(skip bool) *Options {
	if o.TLS == nil {
		o.TLS = &TLSConfig{}
	}
	o.TLS.InsecureSkipVerify = skip
	return o
}

// Clone 克隆配置对象
func (o *Options) Clone() *Options {
	if o == nil {
		return NewOptions()
	}

	clone := &Options{
		Timeout: o.Timeout,
		Headers: make(map[string]string),
		Cookies: make(map[string]string),
	}

	for k, v := range o.Headers {
		clone.Headers[k] = v
	}

	for k, v := range o.Cookies {
		clone.Cookies[k] = v
	}

	if o.Proxy != nil {
		clone.Proxy = &ProxyConfig{
			URL:      o.Proxy.URL,
			Username: o.Proxy.Username,
			Password: o.Proxy.Password,
		}
	}

	if o.TLS != nil {
		clone.TLS = &TLSConfig{
			InsecureSkipVerify: o.TLS.InsecureSkipVerify,
		}
	}

	return clone
}
