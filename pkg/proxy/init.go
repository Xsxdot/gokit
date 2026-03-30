package proxy

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	config "github.com/xsxdot/gokit/config"

	"golang.org/x/net/proxy"
)

// NewDialer 创建 SOCKS5 dialer
// 如果代理未启用，返回默认的 net.Dialer
func NewDialer(cfg config.ProxyConfig) proxy.Dialer {
	if !cfg.Enabled {
		return &net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}
	}

	address := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	var auth *proxy.Auth
	if cfg.Username != "" && cfg.Password != "" {
		auth = &proxy.Auth{
			User:     cfg.Username,
			Password: cfg.Password,
		}
	}

	dialer, err := proxy.SOCKS5("tcp", address, auth, proxy.Direct)
	if err != nil {
		// 如果创建代理失败，返回默认dialer
		return &net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}
	}

	return dialer
}

// NewContextDialer 创建支持 context 的 dialer 函数
func NewContextDialer(cfg config.ProxyConfig) func(ctx context.Context, network, address string) (net.Conn, error) {
	dialer := NewDialer(cfg)

	return func(ctx context.Context, network, address string) (net.Conn, error) {
		return dialer.Dial(network, address)
	}
}

// NewHTTPTransport 创建配置了代理的 HTTP Transport
func NewHTTPTransport(cfg config.ProxyConfig) *http.Transport {
	if !cfg.Enabled {
		return &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		}
	}

	dialer := NewDialer(cfg)

	return &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return dialer.Dial(network, addr)
		},
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
}

// NewHTTPClient 创建配置了代理的 HTTP 客户端
func NewHTTPClient(cfg config.ProxyConfig) *http.Client {
	return &http.Client{
		Transport: NewHTTPTransport(cfg),
		Timeout:   30 * time.Second,
	}
}