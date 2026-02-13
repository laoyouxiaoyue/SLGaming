package svc

import (
	"context"
	"log"
	"time"

	"SLGaming/back/pkg/ioc"
	"SLGaming/back/pkg/lock"
	"SLGaming/back/services/order/internal/config"
	orderioc "SLGaming/back/services/order/internal/ioc"
	orderMQ "SLGaming/back/services/order/internal/mq"

	"SLGaming/back/services/user/userclient"

	rocketmq "github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config config.Config

	DB      *gorm.DB
	Redis   *redis.Redis
	UserRPC userclient.User

	// 分布式锁（用于订单创建、取消等并发控制）
	DistributedLock *lock.DistributedLock

	// RocketMQ 生产者（用于发送订单领域事件）
	OrderEventProducer rocketmq.Producer

	// RocketMQ 事务生产者（用于 CreateOrder 等需要事务消息的场景）
	OrderEventTxProducer rocketmq.TransactionProducer
}

func NewServiceContext(c config.Config) *ServiceContext {
	// 初始化 MySQL
	db, err := orderioc.InitMysql(c.Mysql)
	if err != nil {
		log.Fatalf("failed to init mysql: %v", err)
	}

	// 初始化 Redis（如果配置了）
	var redisClient *redis.Redis
	var distributedLock *lock.DistributedLock
	if c.Redis.Host != "" {
		redisClient = redis.MustNewRedis(c.Redis.RedisConf)
		// 使用 Redis 配置创建分布式锁（基于 redsync，类似 Redisson）
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

	// 初始化 RocketMQ Producer（如果配置了）
	if len(c.RocketMQ.NameServers) > 0 {
		mqCfg := &ioc.RocketMQConfigAdapter{
			NameServers: c.RocketMQ.NameServers,
			Namespace:   c.RocketMQ.Namespace,
			AccessKey:   c.RocketMQ.AccessKey,
			SecretKey:   c.RocketMQ.SecretKey,
		}
		// 普通 Producer（用于非事务消息场景，当前可能暂未使用）
		if producer, err := ioc.InitRocketMQProducer(mqCfg, "order-event-producer"); err != nil {
			logx.Errorf("init rocketmq producer failed: %v", err)
		} else {
			ctx.OrderEventProducer = producer
			logx.Infof("init rocketmq producer success, nameservers=%v", c.RocketMQ.NameServers)
		}

		// 事务 Producer（用于订单相关的事务消息：CreateOrder、CancelOrder、CompleteOrder）
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

	// 初始化 User RPC 客户端（通过 Consul 服务发现）
	if c.Upstream.UserService != "" {
		rpcTimeout := c.Upstream.RPCTimeout
		if rpcTimeout <= 0 {
			rpcTimeout = 10 * time.Second
		}
		if cli, err := newRPCClient(c.Consul, c.Upstream.UserService, rpcTimeout); err != nil {
			logx.Errorf("init user rpc client failed: service=%s, err=%v", c.Upstream.UserService, err)
		} else {
			ctx.UserRPC = userclient.NewUser(cli)
			logx.Infof("init user rpc client success: service=%s (timeout=%v)", c.Upstream.UserService, rpcTimeout)
		}
	}

	return ctx
}

func newRPCClient(consulConf config.ConsulConf, serviceName string, timeout time.Duration) (zrpc.Client, error) {
	endpoints, err := orderioc.ResolveServiceEndpoints(consulConf, serviceName)
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
