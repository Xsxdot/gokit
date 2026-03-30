package security

import (
	"context"
	"time"

	"gokit/config"
)

type AdminAuthComponent struct {
	key    string
	config config.JwtConfig
	entity *AdminAuth
}

// NewAdminAuthComponent 创建 AdminAuth 组件
func NewAdminAuthComponent(key string) *AdminAuthComponent {
	return &AdminAuthComponent{key: key}
}

func (c *AdminAuthComponent) Name() string      { return "admin-auth" }
func (c *AdminAuthComponent) ConfigKey() string { return c.key }
func (c *AdminAuthComponent) ConfigPtr() any    { return &c.config }
func (c *AdminAuthComponent) EntityPtr() any    { return c.entity }
func (c *AdminAuthComponent) Start(ctx context.Context, cfg any) error {
	conf := cfg.(*config.JwtConfig)
	c.entity = NewAdminAuth([]byte(conf.AdminSecret), time.Duration(conf.ExpireTime)*time.Second)
	return nil
}
func (c *AdminAuthComponent) Stop() error { return nil }
