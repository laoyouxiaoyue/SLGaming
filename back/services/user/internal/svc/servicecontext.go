package svc

import (
	"sync"

	pkgIoc "SLGaming/back/pkg/ioc"
	"SLGaming/back/services/user/internal/config"
	"SLGaming/back/services/user/internal/ioc"

	rocketmq "github.com/apache/rocketmq-client-go/v2"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"gorm.io/gorm"
)

type ServiceContext struct {
	mu     sync.RWMutex
	config config.Config
	db     *gorm.DB

	// Redis 客户端（用于排名 ZSet）
	Redis *redis.Redis

	// RocketMQ 生产者（用于发送订单相关事件，如 ORDER_REFUND_SUCCEEDED）
	EventProducer rocketmq.Producer
}

// NewServiceContext 根据配置初始化所有依赖。
func NewServiceContext(c config.Config) *ServiceContext {
	db, err := ioc.InitMysql(c.Mysql)
	if err != nil {
		panic(err)
	}

	ctx := &ServiceContext{
		config: c,
		db:     db,
	}

	// 初始化 Redis（使用 zrpc.RpcServerConf 中的 Redis 配置）
	if c.Redis.Host != "" {
		redisAdapter := &pkgIoc.RedisConfigAdapter{
			Host: c.Redis.Host,
			Type: c.Redis.Type,
			Pass: c.Redis.Pass,
			Tls:  c.Redis.Tls,
		}
		redisClient, err := pkgIoc.InitRedis(redisAdapter)
		if err != nil {
			logx.Errorf("init redis failed: %v", err)
		} else {
			ctx.Redis = redisClient
			logx.Infof("Redis 已初始化")
		}
	} else {
		logx.Infof("Redis 未配置，排名功能不可用")
	}

	// 初始化 RocketMQ Producer（如果配置了）
	if len(c.RocketMQ.NameServers) > 0 {
		mqCfg := &pkgIoc.RocketMQConfigAdapter{
			NameServers: c.RocketMQ.NameServers,
			Namespace:   c.RocketMQ.Namespace,
			AccessKey:   c.RocketMQ.AccessKey,
			SecretKey:   c.RocketMQ.SecretKey,
		}
		producer, err := pkgIoc.InitRocketMQProducer(mqCfg, "user-event-producer")
		if err != nil {
			logx.Errorf("init rocketmq producer failed: %v", err)
		} else {
			ctx.EventProducer = producer
			logx.Infof("init rocketmq producer success, nameservers=%v", c.RocketMQ.NameServers)
		}
	}

	return ctx
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
