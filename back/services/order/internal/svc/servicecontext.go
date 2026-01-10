package svc

import (
	"log"

	"SLGaming/back/pkg/ioc"
	"SLGaming/back/pkg/lock"
	"SLGaming/back/services/order/internal/config"
	orderioc "SLGaming/back/services/order/internal/ioc"

	"SLGaming/back/services/user/userclient"

	rocketmq "github.com/apache/rocketmq-client-go/v2"
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
		distributedLock = lock.NewDistributedLock(redisClient)
		logx.Infof("分布式锁已初始化")
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
		producer, err := ioc.InitRocketMQProducer(mqCfg, "order-event-producer")
		if err != nil {
			logx.Errorf("init rocketmq producer failed: %v", err)
		} else {
			ctx.OrderEventProducer = producer
			logx.Infof("init rocketmq producer success, nameservers=%v", c.RocketMQ.NameServers)
		}
	}

	// 初始化 User RPC 客户端（通过 Consul 服务发现）
	if c.Upstream.UserService != "" {
		if cli, err := newRPCClient(c.Consul, c.Upstream.UserService); err != nil {
			logx.Errorf("init user rpc client failed: service=%s, err=%v", c.Upstream.UserService, err)
		} else {
			ctx.UserRPC = userclient.NewUser(cli)
			logx.Infof("init user rpc client success: service=%s", c.Upstream.UserService)
		}
	}

	return ctx
}

func newRPCClient(consulConf config.ConsulConf, serviceName string) (zrpc.Client, error) {
	endpoints, err := orderioc.ResolveServiceEndpoints(consulConf, serviceName)
	if err != nil {
		logx.Errorf("resolve service endpoints failed: service=%s, err=%v", serviceName, err)
		return nil, err
	}

	client := zrpc.MustNewClient(zrpc.RpcClientConf{
		Endpoints: endpoints,
		NonBlock:  true,
	})

	logx.Infof("create rpc client success: service=%s, endpoints=%v", serviceName, endpoints)
	return client, nil
}
