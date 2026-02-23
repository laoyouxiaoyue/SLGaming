package svc

import (
	"time"

	"SLGaming/back/pkg/ioc"
	"SLGaming/back/pkg/rpc"
	"SLGaming/back/services/agent/agentclient"
	"SLGaming/back/services/code/codeclient"
	"SLGaming/back/services/gateway/internal/config"
	"SLGaming/back/services/gateway/internal/jwt"
	"SLGaming/back/services/order/orderclient"
	"SLGaming/back/services/user/userclient"

	rocketmq "github.com/apache/rocketmq-client-go/v2"
	"github.com/smartwalle/alipay/v3"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

type ServiceContext struct {
	Config config.Config

	CodeRPC       codeclient.Code
	UserRPC       userclient.User
	OrderRPC      orderclient.Order
	AgentRPC      agentclient.Agent
	Alipay        *alipay.Client
	JWT           *jwt.JWTManager
	TokenStore    jwt.TokenStore
	CacheRedis    *redis.Redis
	EventProducer rocketmq.Producer

	dynamicClients []*rpc.DynamicRPCClient
}

func NewServiceContext(c config.Config) *ServiceContext {
	ctx := &ServiceContext{
		Config: c,
	}

	rpcTimeout := c.Upstream.RPCTimeout
	if rpcTimeout <= 0 {
		rpcTimeout = 10 * time.Second
	}

	retryOpts := c.Upstream.Retry
	if retryOpts.MaxRetries == 0 {
		retryOpts = rpc.DefaultRetryOptions()
	}

	consulAdapter := &ioc.ConsulConfigAdapter{
		Address: c.Consul.Address,
		Token:   c.Consul.Token,
	}

	if c.Upstream.CodeService != "" {
		if cli, err := rpc.NewDynamicRPCClientOrFallback(consulAdapter, rpc.DynamicClientOptions{
			ServiceName: c.Upstream.CodeService,
			Timeout:     rpcTimeout,
			Retry:       retryOpts,
		}); err != nil {
			logx.Errorf("初始化 Code RPC 客户端失败: service=%s, error=%v", c.Upstream.CodeService, err)
		} else if cli != nil {
			ctx.CodeRPC = codeclient.NewCode(cli)
			logx.Infof("成功初始化 Code RPC 客户端: service=%s (动态客户端+自动重试)", c.Upstream.CodeService)
		}
	}

	if c.Upstream.UserService != "" {
		if cli, err := rpc.NewDynamicRPCClientOrFallback(consulAdapter, rpc.DynamicClientOptions{
			ServiceName: c.Upstream.UserService,
			Timeout:     rpcTimeout,
			Retry:       retryOpts,
		}); err != nil {
			logx.Errorf("初始化 User RPC 客户端失败: service=%s, error=%v", c.Upstream.UserService, err)
		} else if cli != nil {
			ctx.UserRPC = userclient.NewUser(cli)
			logx.Infof("成功初始化 User RPC 客户端: service=%s (动态客户端+自动重试)", c.Upstream.UserService)
		}
	}

	if c.Upstream.OrderService != "" {
		if cli, err := rpc.NewDynamicRPCClientOrFallback(consulAdapter, rpc.DynamicClientOptions{
			ServiceName: c.Upstream.OrderService,
			Timeout:     rpcTimeout,
			Retry:       retryOpts,
		}); err != nil {
			logx.Errorf("初始化 Order RPC 客户端失败: service=%s, error=%v", c.Upstream.OrderService, err)
		} else if cli != nil {
			ctx.OrderRPC = orderclient.NewOrder(cli)
			logx.Infof("成功初始化 Order RPC 客户端: service=%s (动态客户端+自动重试)", c.Upstream.OrderService)
		}
	}

	if c.Upstream.AgentService != "" {
		if cli, err := rpc.NewDynamicRPCClientOrFallback(consulAdapter, rpc.DynamicClientOptions{
			ServiceName: c.Upstream.AgentService,
			Timeout:     rpcTimeout,
			Retry:       retryOpts,
		}); err != nil {
			logx.Errorf("初始化 Agent RPC 客户端失败: service=%s, error=%v", c.Upstream.AgentService, err)
		} else if cli != nil {
			ctx.AgentRPC = agentclient.NewAgent(cli)
			logx.Infof("成功初始化 Agent RPC 客户端: service=%s (动态客户端+自动重试)", c.Upstream.AgentService)
		}
	}

	secretKey := c.JWT.SecretKey
	if secretKey == "" {
		secretKey = "default-secret-key-change-in-production"
		logx.Infof("JWT secret key not configured, using default key")
	}
	accessTokenDuration := c.JWT.AccessTokenDuration
	if accessTokenDuration <= 0 {
		accessTokenDuration = 10 * time.Minute
	}
	refreshTokenDuration := c.JWT.RefreshTokenDuration
	if refreshTokenDuration <= 0 {
		refreshTokenDuration = 14 * 24 * time.Hour
	}
	ctx.JWT = jwt.NewJWTManager(secretKey, accessTokenDuration, refreshTokenDuration)

	if c.Redis.Host == "" {
		logx.Errorf("Redis 配置不能为空，Refresh Token 存储需要 Redis")
		panic("Redis 配置不能为空，Refresh Token 存储需要 Redis")
	}
	redisClient := redis.MustNewRedis(c.Redis.RedisConf)
	ctx.TokenStore = jwt.NewRedisTokenStore(redisClient)
	ctx.CacheRedis = redisClient
	logx.Infof("使用 Redis 存储 Refresh Token")

	if c.Alipay.AppID != "" && c.Alipay.PrivateKey != "" {
		client, err := alipay.New(c.Alipay.AppID, c.Alipay.PrivateKey, c.Alipay.IsProduction)
		if err != nil {
			logx.Errorf("初始化支付宝客户端失败: %v", err)
		} else {
			if c.Alipay.AlipayPublicKey != "" {
				client.LoadAliPayPublicKey(c.Alipay.AlipayPublicKey)
			}
			ctx.Alipay = client
			logx.Infof("成功初始化支付宝客户端")
		}
	}

	if len(c.RocketMQ.NameServers) > 0 {
		mqCfg := &ioc.RocketMQConfigAdapter{
			NameServers: c.RocketMQ.NameServers,
			Namespace:   c.RocketMQ.Namespace,
			AccessKey:   c.RocketMQ.AccessKey,
			SecretKey:   c.RocketMQ.SecretKey,
		}
		if producer, err := ioc.InitRocketMQProducer(mqCfg, "gateway-recharge-producer"); err != nil {
			logx.Errorf("init rocketmq producer failed: %v", err)
		} else {
			ctx.EventProducer = producer
			logx.Infof("init rocketmq producer success, nameservers=%v", c.RocketMQ.NameServers)
		}
	}

	return ctx
}

func (s *ServiceContext) Close() {
	for _, client := range s.dynamicClients {
		if client != nil {
			client.Stop()
		}
	}
}
