// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package svc

import (
	"time"

	"SLGaming/back/services/code/codeclient"
	"SLGaming/back/services/gateway/internal/config"
	"SLGaming/back/services/gateway/internal/ioc"
	"SLGaming/back/services/gateway/internal/jwt"
	"SLGaming/back/services/order/orderclient"
	"SLGaming/back/services/user/userclient"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config config.Config

	CodeRPC    codeclient.Code
	UserRPC    userclient.User
	OrderRPC   orderclient.Order
	JWT        *jwt.JWTManager
	TokenStore jwt.TokenStore
}

func NewServiceContext(c config.Config) *ServiceContext {
	ctx := &ServiceContext{
		Config: c,
	}

	// 初始化 Code RPC 客户端
	if c.Upstream.CodeService != "" {
		if cli, err := newRPCClient(c.Consul, c.Upstream.CodeService); err != nil {
			logx.Errorf("初始化 Code RPC 客户端失败: service=%s, error=%v", c.Upstream.CodeService, err)
		} else {
			ctx.CodeRPC = codeclient.NewCode(cli)
			logx.Infof("成功初始化 Code RPC 客户端: service=%s", c.Upstream.CodeService)
		}
	}

	// 初始化 User RPC 客户端
	if c.Upstream.UserService != "" {
		if cli, err := newRPCClient(c.Consul, c.Upstream.UserService); err != nil {
			logx.Errorf("初始化 User RPC 客户端失败: service=%s, error=%v", c.Upstream.UserService, err)
		} else {
			ctx.UserRPC = userclient.NewUser(cli)
			logx.Infof("成功初始化 User RPC 客户端: service=%s", c.Upstream.UserService)
		}
	}

	// 初始化 Order RPC 客户端
	if c.Upstream.OrderService != "" {
		if cli, err := newRPCClient(c.Consul, c.Upstream.OrderService); err != nil {
			logx.Errorf("初始化 Order RPC 客户端失败: service=%s, error=%v", c.Upstream.OrderService, err)
		} else {
			ctx.OrderRPC = orderclient.NewOrder(cli)
			logx.Infof("成功初始化 Order RPC 客户端: service=%s", c.Upstream.OrderService)
		}
	}

	// 初始化 JWT 管理器
	secretKey := c.JWT.SecretKey
	if secretKey == "" {
		secretKey = "default-secret-key-change-in-production" // 默认密钥，生产环境需要修改
		logx.Infof("JWT secret key not configured, using default key")
	}
	accessTokenDuration := c.JWT.AccessTokenDuration
	if accessTokenDuration <= 0 {
		accessTokenDuration = 15 * time.Minute // 默认 15 分钟
	}
	refreshTokenDuration := c.JWT.RefreshTokenDuration
	if refreshTokenDuration <= 0 {
		refreshTokenDuration = 7 * 24 * time.Hour // 默认 7 天
	}
	ctx.JWT = jwt.NewJWTManager(secretKey, accessTokenDuration, refreshTokenDuration)

	// 初始化 Token 存储（必须使用 Redis）
	if c.Redis.Host == "" {
		logx.Errorf("Redis 配置不能为空，Refresh Token 存储需要 Redis")
		panic("Redis 配置不能为空，Refresh Token 存储需要 Redis")
	}
	redisClient := redis.MustNewRedis(c.Redis.RedisConf)
	ctx.TokenStore = jwt.NewRedisTokenStore(redisClient)
	logx.Infof("使用 Redis 存储 Refresh Token")

	return ctx
}

func newRPCClient(consulConf config.ConsulConf, serviceName string) (zrpc.Client, error) {
	endpoints, err := ioc.ResolveServiceEndpoints(consulConf, serviceName)
	if err != nil {
		logx.Errorf("解析服务端点失败: service=%s, error=%v", serviceName, err)
		return nil, err
	}

	client := zrpc.MustNewClient(zrpc.RpcClientConf{
		Endpoints: endpoints,
		NonBlock:  true,
	})

	logx.Infof("成功创建 RPC 客户端: service=%s, endpoints=%v", serviceName, endpoints)
	return client, nil
}
