package grpc

import (
	"context"

	"gokit/config"

	"go.uber.org/zap"
)

type GRPCServerComponent struct {
	key    string
	config config.GRPCConfig
	logger *zap.Logger
	entity *Server
}

// NewGRPCServerComponent 创建 gRPC Server 组件
func NewGRPCServerComponent(key string) *GRPCServerComponent {
	return &GRPCServerComponent{key: key}
}

// WithLogger 设置 Logger（可选）
func (c *GRPCServerComponent) WithLogger(logger *zap.Logger) *GRPCServerComponent {
	c.logger = logger
	return c
}

func (c *GRPCServerComponent) Name() string      { return "grpc" }
func (c *GRPCServerComponent) ConfigKey() string { return c.key }
func (c *GRPCServerComponent) ConfigPtr() any    { return &c.config }
func (c *GRPCServerComponent) EntityPtr() any    { return c.entity }
func (c *GRPCServerComponent) Start(ctx context.Context, cfg any) error {
	conf := cfg.(*config.GRPCConfig)
	logger := c.logger
	if logger == nil {
		logger = zap.NewNop()
	}
	// 转换 config.GRPCConfig 为 grpc.Config
	grpcCfg := &Config{
		Address:           conf.Address,
		EnableReflection:  conf.EnableReflection,
		EnableRecovery:    conf.EnableRecovery,
		EnableValidation:  conf.EnableValidation,
		EnableAuth:        conf.EnableAuth,
		EnablePermission:  conf.EnablePermission,
		LogLevel:          conf.LogLevel,
		MaxRecvMsgSize:    conf.MaxRecvMsgSize,
		MaxSendMsgSize:    conf.MaxSendMsgSize,
		ConnectionTimeout: conf.ConnectionTimeout,
	}
	c.entity = NewServer(grpcCfg, logger)
	return c.entity.Start()
}
func (c *GRPCServerComponent) Stop() error {
	if c.entity != nil {
		c.entity.Stop()
	}
	return nil
}
