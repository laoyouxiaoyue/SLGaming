package svc

import (
	"context"
	"fmt"
	"sync"
	"time"

	pkgIoc "SLGaming/back/pkg/ioc"
	"SLGaming/back/pkg/lock"
	"SLGaming/back/services/agent/agentclient"
	"SLGaming/back/services/user/internal/bloom"
	"SLGaming/back/services/user/internal/cache"
	"SLGaming/back/services/user/internal/config"
	"SLGaming/back/services/user/internal/ioc"
	userMQ "SLGaming/back/services/user/internal/mq"

	rocketmq "github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
	"gorm.io/gorm"
)

type ServiceContext struct {
	mu     sync.RWMutex
	config config.Config
	db     *gorm.DB

	// Redis 客户端（用于排名 ZSet）
	Redis *redis.Redis

	// 缓存管理器
	CacheManager *cache.Manager

	// 用户缓存服务
	UserCache *cache.UserCache

	// 布隆过滤器
	BloomFilter *bloom.UserBloomFilters

	// 分布式锁（用于钱包充值/扣款等并发控制）
	DistributedLock *lock.DistributedLock

	// RocketMQ 普通生产者（用于发送非事务事件）
	EventProducer rocketmq.Producer

	// RocketMQ 事务生产者（用于用户领域事件的半消息，如 ORDER_REFUND_SUCCEEDED）
	EventTxProducer rocketmq.TransactionProducer

	// Agent RPC 客户端（用于头像审核等异步任务）
	AgentRPC agentclient.Agent
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

	// 初始化 Redis（用于排行榜、缓存、布隆过滤器）
	var redisClient *redis.Redis
	if c.Redis.Host != "" {
		redisClient = redis.MustNewRedis(c.Redis.RedisConf)
		ctx.Redis = redisClient
		logx.Infof("Redis 已初始化: %s", c.Redis.Host)

		// 初始化分布式锁
		redisConfig := &pkgIoc.RedisConfigAdapter{
			Host: c.Redis.Host,
			Type: c.Redis.Type,
			Pass: c.Redis.Pass,
			Tls:  c.Redis.Tls,
		}
		ctx.DistributedLock = lock.NewDistributedLockFromConfig(redisConfig)
		logx.Infof("分布式锁已初始化")
	} else {
		logx.Infof("Redis 未配置，排名功能、缓存、布隆过滤器将不可用")
	}

	// 初始化缓存管理器和用户缓存服务
	ctx.CacheManager = cache.NewManager(redisClient)
	ctx.UserCache = cache.NewUserCache(ctx.CacheManager)
	logx.Infof("缓存服务已初始化")

	// 初始化布隆过滤器（如果为空会自动从数据库导入）
	ctx.BloomFilter = bloom.NewUserBloomFilters(redisClient, db)
	logx.Infof("布隆过滤器已初始化")

	// 初始化 RocketMQ Producer（如果配置了）
	if len(c.RocketMQ.NameServers) > 0 {
		mqCfg := &pkgIoc.RocketMQConfigAdapter{
			NameServers: c.RocketMQ.NameServers,
			Namespace:   c.RocketMQ.Namespace,
			AccessKey:   c.RocketMQ.AccessKey,
			SecretKey:   c.RocketMQ.SecretKey,
		}
		// 普通 Producer
		if producer, err := pkgIoc.InitRocketMQProducer(mqCfg, "user-event-producer"); err != nil {
			logx.Errorf("init rocketmq producer failed: %v", err)
		} else {
			ctx.EventProducer = producer
			logx.Infof("init rocketmq producer success, nameservers=%v", c.RocketMQ.NameServers)
		}

		// 事务 Producer：用于用户领域事件（目前主要是 ORDER_REFUND_SUCCEEDED）
		if txProducer, err := pkgIoc.InitRocketMQTransactionProducer(
			mqCfg,
			"user-transaction-producer",
			func(cctx context.Context, msg *primitive.Message) primitive.LocalTransactionState {
				return userMQ.ExecuteUserEventTx(cctx, db, msg)
			},
			func(cctx context.Context, msg *primitive.Message) primitive.LocalTransactionState {
				return userMQ.CheckUserEventTx(cctx, db, msg)
			},
		); err != nil {
			logx.Errorf("init rocketmq transaction producer failed: %v", err)
		} else {
			ctx.EventTxProducer = txProducer
			logx.Infof("init rocketmq transaction producer success, nameservers=%v", c.RocketMQ.NameServers)
		}
	}

	// 初始化 Agent RPC 客户端
	if c.Upstream.AgentService != "" {
		rpcTimeout := c.Upstream.RPCTimeout
		if rpcTimeout <= 0 {
			rpcTimeout = 10 * time.Second
		}
		if cli, err := newRPCClient(c.Consul, c.Upstream.AgentService, rpcTimeout); err != nil {
			logx.Errorf("init agent rpc client failed: service=%s, err=%v", c.Upstream.AgentService, err)
		} else {
			ctx.AgentRPC = agentclient.NewAgent(cli)
			logx.Infof("init agent rpc client success: service=%s (timeout=%v)", c.Upstream.AgentService, rpcTimeout)
		}
	}

	return ctx
}

func newRPCClient(consulConf config.ConsulConf, serviceName string, timeout time.Duration) (zrpc.Client, error) {
	if serviceName == "" {
		return nil, fmt.Errorf("service name is empty")
	}

	// 通过Consul服务发现获取Agent服务地址
	adapter := &pkgIoc.ConsulConfigAdapter{
		Address: consulConf.Address,
		Token:   consulConf.Token,
	}

	endpoints, err := pkgIoc.ResolveServiceEndpoints(adapter, serviceName)
	if err != nil {
		logx.Errorf("resolve service endpoints failed: service=%s, err=%v", serviceName, err)
		return nil, err
	}

	client := zrpc.MustNewClient(zrpc.RpcClientConf{
		Endpoints: endpoints,
		NonBlock:  true,
		Timeout:   int64(timeout / time.Millisecond),
	})
	logx.Infof("create rpc client success: service=%s, endpoints=%v, timeout=%v", serviceName, endpoints, timeout)

	return client, nil
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
