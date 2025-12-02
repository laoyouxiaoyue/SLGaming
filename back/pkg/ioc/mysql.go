package ioc

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// InitMySQL 根据配置初始化 MySQL 连接
func InitMySQL(cfg MySQLConfig) (*gorm.DB, error) {
	dsn := cfg.GetDSN()
	if dsn == "" {
		return nil, fmt.Errorf("mysql dsn is empty")
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("open mysql: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql db: %w", err)
	}

	// 设置连接池参数
	if maxIdleConns := cfg.GetMaxIdleConns(); maxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(maxIdleConns)
	}

	if maxOpenConns := cfg.GetMaxOpenConns(); maxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(maxOpenConns)
	}

	// 设置连接最大生命周期
	if connMaxLifetime := cfg.GetConnMaxLifetime(); connMaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(connMaxLifetime)
	}

	// 设置连接最大空闲时间
	if connMaxIdleTime := cfg.GetConnMaxIdleTime(); connMaxIdleTime > 0 {
		sqlDB.SetConnMaxIdleTime(connMaxIdleTime)
	}

	return db, nil
}
