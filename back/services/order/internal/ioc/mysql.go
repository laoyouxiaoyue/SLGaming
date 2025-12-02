package ioc

import (
	"SLGaming/back/pkg/ioc"
	"SLGaming/back/services/order/internal/config"

	"gorm.io/gorm"
)

// InitMysql 根据配置初始化 MySQL 连接
func InitMysql(cfg config.MysqlConf) (*gorm.DB, error) {
	adapter := &ioc.MySQLConfigAdapter{
		DSN:             cfg.DSN,
		MaxIdleConns:    cfg.MaxIdleConns,
		MaxOpenConns:    cfg.MaxOpenConns,
		ConnMaxLifetime: cfg.ConnMaxLifetime,
		ConnMaxIdleTime: cfg.ConnMaxIdleTime,
	}
	return ioc.InitMySQL(adapter)
}


