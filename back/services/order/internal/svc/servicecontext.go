package svc

import (
	"log"

	"SLGaming/back/services/order/internal/config"
	orderioc "SLGaming/back/services/order/internal/ioc"

	"SLGaming/back/services/user/userclient"

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
}

func NewServiceContext(c config.Config) *ServiceContext {
	// 初始化 MySQL
	db, err := orderioc.InitMysql(c.Mysql)
	if err != nil {
		log.Fatalf("failed to init mysql: %v", err)
	}

	// 初始化 Redis（如果配置了）
	var redisClient *redis.Redis
	if c.Redis.Host != "" {
		redisClient = redis.MustNewRedis(c.Redis.RedisConf)
	}

	ctx := &ServiceContext{
		Config: c,
		DB:     db,
		Redis:  redisClient,
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

