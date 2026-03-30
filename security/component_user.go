package security

import (
	"context"
	"time"

	"gokit/config"
)

type UserAuthComponent struct {
	key    string
	config config.JwtConfig
	entity *UserAuth
}

// NewUserAuthComponent 创建 UserAuth 组件
func NewUserAuthComponent(key string) *UserAuthComponent {
	return &UserAuthComponent{key: key}
}

func (c *UserAuthComponent) Name() string      { return "user-auth" }
func (c *UserAuthComponent) ConfigKey() string { return c.key }
func (c *UserAuthComponent) ConfigPtr() any    { return &c.config }
func (c *UserAuthComponent) EntityPtr() any    { return c.entity }
func (c *UserAuthComponent) Start(ctx context.Context, cfg any) error {
	conf := cfg.(*config.JwtConfig)
	c.entity = NewUserAuth([]byte(conf.Secret), time.Duration(conf.ExpireTime)*time.Second)
	return nil
}
func (c *UserAuthComponent) Stop() error { return nil }
