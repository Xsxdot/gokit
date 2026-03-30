package oss

import (
	"context"

	"gokit/config"
)

type OSSComponent struct {
	key    string
	config config.OssConfig
	entity *AliyunService
}

// NewOSSComponent 创建 OSS 组件
func NewOSSComponent(key string) *OSSComponent {
	return &OSSComponent{key: key}
}

func (c *OSSComponent) Name() string      { return "oss" }
func (c *OSSComponent) ConfigKey() string { return c.key }
func (c *OSSComponent) ConfigPtr() any    { return &c.config }
func (c *OSSComponent) EntityPtr() any    { return c.entity }
func (c *OSSComponent) Start(ctx context.Context, cfg any) error {
	conf := cfg.(*config.OssConfig)
	service, err := NewAliyunService(conf)
	if err != nil {
		return err
	}
	c.entity = service
	return nil
}
func (c *OSSComponent) Stop() error { return nil }
