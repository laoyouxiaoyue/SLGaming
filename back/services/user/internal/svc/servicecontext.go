package svc

import (
	"sync"

	"SLGaming/back/services/user/internal/config"
	"SLGaming/back/services/user/internal/ioc"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type ServiceContext struct {
	mu     sync.RWMutex
	config config.Config
	db     *gorm.DB
}

// NewServiceContext 根据配置初始化所有依赖。
func NewServiceContext(c config.Config) *ServiceContext {
	db, err := ioc.InitMysql(c.Mysql)
	if err != nil {
		panic(err)
	}

	return &ServiceContext{
		config: c,
		db:     db,
	}
}

// Config returns the latest configuration snapshot.
// Config 返回最新的配置快照（复制值，避免外部修改底层状态）。
func (s *ServiceContext) Config() config.Config {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config
}

// DB returns the current database connection.
// DB 返回当前可用的数据库连接。
func (s *ServiceContext) DB() *gorm.DB {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.db
}

// UpdateConfig 更新配置，并在必要时重建依赖资源。
func (s *ServiceContext) UpdateConfig(newCfg config.Config) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if needReconnect(s.config.Mysql, newCfg.Mysql) {
		newDB, err := ioc.InitMysql(newCfg.Mysql)
		if err != nil {
			return err
		}
		s.replaceDB(newDB)
	}

	s.config = newCfg
	return nil
}

// replaceDB 将旧 DB 关闭并替换为新的连接。
func (s *ServiceContext) replaceDB(newDB *gorm.DB) {
	if s.db != nil {
		sqlDB, err := s.db.DB()
		if err == nil {
			if err = sqlDB.Close(); err != nil {
				logx.Errorf("close old db connection: %v", err)
			}
		}
	}
	s.db = newDB
}

// needReconnect 判断 MySQL 配置是否发生变化，决定是否需要重建连接。
func needReconnect(oldCfg, newCfg config.MysqlConf) bool {
	return oldCfg.DSN != newCfg.DSN ||
		oldCfg.MaxIdleConns != newCfg.MaxIdleConns ||
		oldCfg.MaxOpenConns != newCfg.MaxOpenConns ||
		oldCfg.ConnMaxLifetime != newCfg.ConnMaxLifetime ||
		oldCfg.ConnMaxIdleTime != newCfg.ConnMaxIdleTime
}
