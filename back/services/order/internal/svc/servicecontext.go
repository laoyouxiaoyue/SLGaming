package svc

import (
	"context"
	"log"

	"SLGaming/back/pkg/ioc"
	"SLGaming/back/pkg/lock"
	"SLGaming/back/pkg/rpc"
	"SLGaming/back/services/order/internal/config"
	orderioc "SLGaming/back/services/order/internal/ioc"
	orderMQ "SLGaming/back/services/order/internal/mq"

	"SLGaming/back/services/user/userclient"

	rocketmq "github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config config.Config

	DB      *gorm.DB
	Redis   *redis.Redis
	UserRPC userclient.User

	DistributedLock *lock.DistributedLock

	OrderEventProducer   rocketmq.Producer
	OrderEventTxProducer rocketmq.TransactionProducer
	dynamicRPCClients    []*rpc.DynamicRPCClient
}

func NewServiceContext(c config.Config) *ServiceContext {
	db, err := orderioc.InitMysql(c.Mysql)
	if err != nil {
		log.Fatalf("failed to init mysql: %v", err)
	}

	var redisClient *redis.Redis
	var distributedLock *lock.DistributedLock
	if c.Redis.Host != "" {
		redisClient = redis.MustNewRedis(c.Redis.RedisConf)
		redisConfig := &ioc.RedisConfigAdapter{
			Host: c.Redis.Host,
			Type: c.Redis.Type,
			Pass: c.Redis.Pass,
			Tls:  c.Redis.Tls,
		}
		distributedLock = lock.NewDistributedLockFromConfig(redisConfig)
		logx.Infof("分布式锁已初始化（基于 redsync）")
	} else {
		logx.Infof("Redis 未配置，分布式锁功能不可用")
	}

	ctx := &ServiceContext{
		Config:          c,
		DB:              db,
		Redis:           redisClient,
		DistributedLock: distributedLock,
	}

	if len(c.RocketMQ.NameServers) > 0 {
		mqCfg := &ioc.RocketMQConfigAdapter{
			NameServers: c.RocketMQ.NameServers,
			Namespace:   c.RocketMQ.Namespace,
			AccessKey:   c.RocketMQ.AccessKey,
			SecretKey:   c.RocketMQ.SecretKey,
		}
		if producer, err := ioc.InitRocketMQProducer(mqCfg, "order-event-producer"); err != nil {
			logx.Errorf("init rocketmq producer failed: %v", err)
		} else {
			ctx.OrderEventProducer = producer
			logx.Infof("init rocketmq producer success, nameservers=%v", c.RocketMQ.NameServers)
		}

		txProducer, err := ioc.InitRocketMQTransactionProducer(
			mqCfg,
			"order-transaction-producer",
			func(cctx context.Context, msg *primitive.Message) primitive.LocalTransactionState {
				return orderMQ.ExecuteOrderTx(cctx, db, msg)
			},
			func(cctx context.Context, msg *primitive.Message) primitive.LocalTransactionState {
				return orderMQ.CheckOrderTx(cctx, db, msg)
			},
		)
		if err != nil {
			logx.Errorf("init rocketmq transaction producer failed: %v", err)
		} else {
			ctx.OrderEventTxProducer = txProducer
			logx.Infof("init rocketmq transaction producer success, nameservers=%v", c.RocketMQ.NameServers)
		}
	}

	if c.Upstream.UserService != "" {
		retryOpts := c.Upstream.Retry
		if retryOpts.MaxRetries == 0 {
			retryOpts = rpc.DefaultRetryOptions()
		}

		cli, err := rpc.NewDynamicRPCClientOrFallback(&ioc.ConsulConfigAdapter{
			Address: c.Consul.Address,
			Token:   c.Consul.Token,
		}, rpc.DynamicClientOptions{
			ServiceName: c.Upstream.UserService,
			Timeout:     c.Upstream.RPCTimeout,
			Retry:       retryOpts,
		})
		if err != nil {
			logx.Errorf("init user rpc client failed: service=%s, err=%v", c.Upstream.UserService, err)
		} else if cli != nil {
			ctx.UserRPC = userclient.NewUser(cli)
			logx.Infof("init user rpc client success: service=%s (动态客户端+自动重试)", c.Upstream.UserService)
		}
	}

	return ctx
}

func (s *ServiceContext) Close() {
	for _, client := range s.dynamicRPCClients {
		if client != nil {
			client.Stop()
		}
	}
}
