package mysql

import (
	"context"
	"fmt"
	"net"
	"time"

	config "github.com/xsxdot/gokit/config"
	"github.com/xsxdot/gokit/pkg/proxy"

	mysqldriver "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// InitMysql 初始化 MySQL 数据库连接
func InitMysql(database config.Database, proxyCfg config.ProxyConfig) (*gorm.DB, error) {
	var dialector gorm.Dialector

	if proxyCfg.Enabled {
		// 注册自定义dialer到MySQL驱动
		dialerName := fmt.Sprintf("proxy_%d", time.Now().UnixNano())
		dialer := proxy.NewDialer(proxyCfg)

		mysqldriver.RegisterDialContext(dialerName, func(ctx context.Context, addr string) (net.Conn, error) {
			return dialer.Dial("tcp", addr)
		})

		// 构建DSN，使用自定义dialer
		dsn := fmt.Sprintf("%s:%s@%s(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			database.User, database.Password, dialerName, database.Host, database.Port, database.DbName)
		dialector = mysql.Open(dsn)
	} else {
		// 不使用代理的常规连接
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			database.User, database.Password, database.Host, database.Port, database.DbName)
		dialector = mysql.Open(dsn)
	}

	return gorm.Open(dialector, &gorm.Config{})
}

type DBComponent struct {
	key      string
	config   config.Database
	proxyCfg *config.ProxyConfig
	entity   **gorm.DB
}

// NewDBComponent 创建数据库组件
func NewDBComponent(key string, entity **gorm.DB) *DBComponent {
	return &DBComponent{key: key, entity: entity}
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

	entity, err := InitMysql(*dbCfg, proxyCfg)
	*c.entity = entity
	return err
}
func (c *DBComponent) Stop() error {
	if c.entity != nil {
		sqlDB, _ := (*c.entity).DB()
		return sqlDB.Close()
	}
	return nil
}
