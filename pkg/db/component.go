package db

import (
	"context"

	config "github.com/xsxdot/gokit/config"
	"github.com/xsxdot/gokit/pkg/mysql"
	"github.com/xsxdot/gokit/pkg/pg"

	"gorm.io/gorm"
)

type DBComponent struct {
	key      string
	config   config.Database
	proxyCfg *config.ProxyConfig
	entity   *gorm.DB
}

// NewDBComponent 创建数据库组件
func NewDBComponent(key string) *DBComponent {
	return &DBComponent{key: key}
}

// WithProxy 设置代理配置（可选）
func (c *DBComponent) WithProxy(proxyCfg *config.ProxyConfig) *DBComponent {
	c.proxyCfg = proxyCfg
	return c
}

func (c *DBComponent) Name() string      { return "database" }
func (c *DBComponent) ConfigKey() string { return c.key }
func (c *DBComponent) ConfigPtr() any    { return &c.config }
func (c *DBComponent) EntityPtr() any    { return c.entity }
func (c *DBComponent) Start(ctx context.Context, cfg any) error {
	dbCfg := cfg.(*config.Database)
	var err error

	proxyCfg := config.ProxyConfig{}
	if c.proxyCfg != nil {
		proxyCfg = *c.proxyCfg
	}

	if dbCfg.Type == "postgres" {
		c.entity, err = pg.InitPg(*dbCfg)
	} else {
		c.entity, err = mysql.InitMysql(*dbCfg, proxyCfg)
	}
	return err
}
func (c *DBComponent) Stop() error {
	if c.entity != nil {
		sqlDB, _ := c.entity.DB()
		return sqlDB.Close()
	}
	return nil
}
