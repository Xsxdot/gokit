package pg

import (
	"fmt"
	"time"

	config "github.com/xsxdot/gokit/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// InitPg 初始化 PostgreSQL 数据库连接
func InitPg(database config.Database) (*gorm.DB, error) {
	// 构建DSN
	dsn := fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=disable password=%s",
		database.Host, database.Port, database.User, database.DbName, database.Password)

	// 打开连接
	cfg := &gorm.Config{}
	db, err := gorm.Open(postgres.Open(dsn), cfg)
	if err != nil {
		return nil, err
	}

	// 获取底层的sql.DB并配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// PostgreSQL的驱动不直接支持自定义dialer
	// 代理功能需要在网络层面配置或使用支持代理的pgx驱动
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db, nil
}