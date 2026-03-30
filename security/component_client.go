package security

import (
	"context"

	"github.com/xsxdot/gokit/config"
)

type ClientAuthComponent struct {
	key    string
	config config.JwtConfig
	entity *ClientAuth
}

// NewClientAuthComponent 创建 ClientAuth 组件
func NewClientAuthComponent(key string) *ClientAuthComponent {
	return &ClientAuthComponent{key: key}
}

func (c *ClientAuthComponent) Name() string      { return "client-auth" }
func (c *ClientAuthComponent) ConfigKey() string { return c.key }
func (c *ClientAuthComponent) ConfigPtr() any    { return &c.config }
func (c *ClientAuthComponent) EntityPtr() any    { return c.entity }
func (c *ClientAuthComponent) Start(ctx context.Context, cfg any) error {
	conf := cfg.(*config.JwtConfig)
	c.entity = NewClientAuth([]byte(conf.ClientSecret))
	return nil
}
func (c *ClientAuthComponent) Stop() error { return nil }
