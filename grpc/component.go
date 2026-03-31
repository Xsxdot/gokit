package grpc

import (
	"context"

	config "github.com/xsxdot/gokit/config"

	"go.uber.org/zap"
)

type GRPCServerComponent struct {
	key          string
	config       config.GRPCConfig
	logger       *zap.Logger
	entity       **Server
	services     []ServiceRegistrar  // 预注册的服务列表
	authProvider AuthProvider        // 预设置的鉴权提供者
	skipMethods  []string            // 跳过鉴权的方法列表
}

// NewGRPCServerComponent 创建 gRPC Server 组件
func NewGRPCServerComponent(key string, entity **Server) *GRPCServerComponent {
	return &GRPCServerComponent{key: key, entity: entity}
}

// WithLogger 设置 Logger（可选）
func (c *GRPCServerComponent) WithLogger(logger *zap.Logger) *GRPCServerComponent {
	c.logger = logger
	return c
}

// RegisterService 预注册服务（必须在 Start 前调用）
func (c *GRPCServerComponent) RegisterService(service ServiceRegistrar) *GRPCServerComponent {
	c.services = append(c.services, service)
	return c
}

// SetAuthProvider 预设置鉴权提供者（必须在 Start 前调用）
func (c *GRPCServerComponent) SetAuthProvider(provider AuthProvider) *GRPCServerComponent {
	c.authProvider = provider
	return c
}

// SetSkipMethods 设置跳过鉴权的方法列表（必须在 Start 前调用）
func (c *GRPCServerComponent) SetSkipMethods(methods []string) *GRPCServerComponent {
	c.skipMethods = methods
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

	// 如果启用鉴权或有 authProvider，创建 AuthConfig
	if conf.EnableAuth || c.authProvider != nil {
		grpcCfg.Auth = DefaultAuthConfig()
		if len(c.skipMethods) > 0 {
			grpcCfg.Auth.SkipMethods = append(grpcCfg.Auth.SkipMethods, c.skipMethods...)
		}
	}

	// 创建 Server
	server := NewServer(grpcCfg, logger)

	// 设置鉴权提供者（如果有）
	if c.authProvider != nil {
		server.SetAuthProvider(c.authProvider)
	}

	// 注册所有预注册的服务
	for _, service := range c.services {
		if err := server.RegisterService(service); err != nil {
			return err
		}
	}

	// 赋值给 entity
	*c.entity = server

	// 启动服务器
	return server.Start()
}
func (c *GRPCServerComponent) Stop() error {
	if c.entity != nil {
		(*c.entity).Stop()
	}
	return nil
}
