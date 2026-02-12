// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package svc

import (
	"sync"
	"time"

	"SLGaming/back/pkg/ioc"
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
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

type ServiceContext struct {
	Config config.Config

	CodeRPC    codeclient.Code
	UserRPC    userclient.User
	OrderRPC   orderclient.Order
	AgentRPC   agentclient.Agent
	Alipay     *alipay.Client
	JWT        *jwt.JWTManager
	TokenStore jwt.TokenStore
	CacheRedis *redis.Redis
	// RocketMQ 事件生产者（用于发送充值回调事件）
	EventProducer rocketmq.Producer
}

func NewServiceContext(c config.Config) *ServiceContext {
	ctx := &ServiceContext{
		Config: c,
	}

	rpcTimeout := c.Upstream.RPCTimeout
	if rpcTimeout <= 0 {
		rpcTimeout = 10 * time.Second
	}

	if c.Upstream.CodeService != "" {
		if cli, err := newRPCClient(c.Consul, c.Upstream.CodeService, rpcTimeout); err != nil {
			logx.Errorf("初始化 Code RPC 客户端失败: service=%s, error=%v", c.Upstream.CodeService, err)
		} else {
			ctx.CodeRPC = codeclient.NewCode(cli)
			logx.Infof("成功初始化 Code RPC 客户端: service=%s (支持动态服务发现)", c.Upstream.CodeService)
		}
	}

	if c.Upstream.UserService != "" {
		if cli, err := newRPCClient(c.Consul, c.Upstream.UserService, rpcTimeout); err != nil {
			logx.Errorf("初始化 User RPC 客户端失败: service=%s, error=%v", c.Upstream.UserService, err)
		} else {
			ctx.UserRPC = userclient.NewUser(cli)
			logx.Infof("成功初始化 User RPC 客户端: service=%s (支持动态服务发现)", c.Upstream.UserService)
		}
	}

	if c.Upstream.OrderService != "" {
		if cli, err := newRPCClient(c.Consul, c.Upstream.OrderService, rpcTimeout); err != nil {
			logx.Errorf("初始化 Order RPC 客户端失败: service=%s, error=%v", c.Upstream.OrderService, err)
		} else {
			ctx.OrderRPC = orderclient.NewOrder(cli)
			logx.Infof("成功初始化 Order RPC 客户端: service=%s (支持动态服务发现)", c.Upstream.OrderService)
		}
	}

	if c.Upstream.AgentService != "" {
		if cli, err := newRPCClient(c.Consul, c.Upstream.AgentService, rpcTimeout); err != nil {
			logx.Errorf("初始化 Agent RPC 客户端失败: service=%s, error=%v", c.Upstream.AgentService, err)
		} else {
			ctx.AgentRPC = agentclient.NewAgent(cli)
			logx.Infof("成功初始化 Agent RPC 客户端: service=%s (支持动态服务发现)", c.Upstream.AgentService)
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
		accessTokenDuration = 10 * time.Minute // 默认 10 分钟
	}
	refreshTokenDuration := c.JWT.RefreshTokenDuration
	if refreshTokenDuration <= 0 {
		refreshTokenDuration = 14 * 24 * time.Hour // 默认 14 天
	}
	ctx.JWT = jwt.NewJWTManager(secretKey, accessTokenDuration, refreshTokenDuration)

	// 初始化 Token 存储（必须使用 Redis）
	if c.Redis.Host == "" {
		logx.Errorf("Redis 配置不能为空，Refresh Token 存储需要 Redis")
		panic("Redis 配置不能为空，Refresh Token 存储需要 Redis")
	}
	redisClient := redis.MustNewRedis(c.Redis.RedisConf)
	ctx.TokenStore = jwt.NewRedisTokenStore(redisClient)
	ctx.CacheRedis = redisClient
	logx.Infof("使用 Redis 存储 Refresh Token")

	// 初始化支付宝客户端
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

	// 初始化 RocketMQ Producer（如果配置了）
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

// DynamicRPCClient 支持动态更新的 RPC 客户端包装器
type DynamicRPCClient struct {
	mu          sync.RWMutex
	client      zrpc.Client
	watcher     *ioc.ConsulWatcher
	serviceName string
	endpoints   []string
	timeout     time.Duration
}

// GetClient 获取当前的 RPC 客户端
func (d *DynamicRPCClient) GetClient() zrpc.Client {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.client
}

// GetEndpoints 获取当前的 endpoints（用于日志）
func (d *DynamicRPCClient) GetEndpoints() []string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.endpoints
}

// updateClient 更新 RPC 客户端（当服务端点变化时）
func (d *DynamicRPCClient) updateClient(endpoints []string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	oldEndpoints := d.endpoints

	d.client = zrpc.MustNewClient(zrpc.RpcClientConf{
		Endpoints: endpoints,
		NonBlock:  true,
		Timeout:   int64(d.timeout / time.Millisecond),
	})

	d.endpoints = endpoints

	logx.Infof("[dynamic_rpc] 更新客户端: service=%s, old_endpoints=%v, new_endpoints=%v",
		d.serviceName, oldEndpoints, endpoints)
}

// Stop 停止监听
func (d *DynamicRPCClient) Stop() {
	if d.watcher != nil {
		d.watcher.Stop()
	}
	// go-zero 的 zrpc.Client 会在服务停止时自动关闭连接
}

func newRPCClient(consulConf config.ConsulConf, serviceName string, timeout time.Duration) (zrpc.Client, error) {
	if consulConf.Address != "" {
		return newConsulDynamicRPCClient(consulConf, serviceName, timeout)
	}

	return newConsulStaticRPCClient(consulConf, serviceName, timeout)
}

// newConsulDynamicRPCClient 使用 Consul Watch 创建动态更新的 RPC 客户端
func newConsulDynamicRPCClient(consulConf config.ConsulConf, serviceName string, timeout time.Duration) (zrpc.Client, error) {
	adapter := &ioc.ConsulConfigAdapter{
		Address: consulConf.Address,
		Token:   consulConf.Token,
	}

	endpoints, err := ioc.ResolveServiceEndpoints(adapter, serviceName)
	if err != nil {
		logx.Infof("初始解析服务端点失败（服务可能未启动）: service=%s, error=%v，将等待服务注册后自动发现", serviceName, err)
		endpoints = []string{}
	}

	var client zrpc.Client
	if len(endpoints) > 0 {
		client = zrpc.MustNewClient(zrpc.RpcClientConf{
			Endpoints: endpoints,
			NonBlock:  true,
			Timeout:   int64(timeout / time.Millisecond),
		})
		logx.Infof("成功创建动态 Consul RPC 客户端: service=%s, initial_endpoints=%v", serviceName, endpoints)
	} else {
		client = zrpc.MustNewClient(zrpc.RpcClientConf{
			Endpoints: []string{"127.0.0.1:0"},
			NonBlock:  true,
			Timeout:   int64(timeout / time.Millisecond),
		})
		logx.Infof("创建动态 Consul RPC 客户端（等待服务注册）: service=%s, 将自动发现服务端点", serviceName)
	}

	dynamicClient := &DynamicRPCClient{
		client:      client,
		serviceName: serviceName,
		endpoints:   endpoints,
		timeout:     timeout,
	}

	watcher, err := ioc.NewConsulWatcher(adapter, serviceName, func(newEndpoints []string) {
		if len(newEndpoints) > 0 {
			dynamicClient.updateClient(newEndpoints)
		} else {
			logx.Infof("[dynamic_rpc] 服务暂未注册: service=%s，继续等待...", serviceName)
		}
	})
	if err != nil {
		logx.Errorf("创建 Consul watcher 失败: service=%s, error=%v，将使用静态端点", serviceName, err)
		return &dynamicRPCClientWrapper{client: dynamicClient}, nil
	}

	dynamicClient.watcher = watcher
	logx.Infof("成功创建动态 Consul RPC 客户端: service=%s (支持动态更新和负载均衡)", serviceName)

	return &dynamicRPCClientWrapper{client: dynamicClient}, nil
}

// dynamicRPCClientWrapper 包装 DynamicRPCClient 实现 zrpc.Client 接口
type dynamicRPCClientWrapper struct {
	client *DynamicRPCClient
}

func (w *dynamicRPCClientWrapper) Conn() *grpc.ClientConn {
	// 每次调用都获取最新的客户端，确保使用最新的 endpoints
	// GetClient() 会加锁并返回最新的 d.client
	// 这样每次 RPC 调用都会使用最新的连接和 endpoints
	return w.client.GetClient().Conn()
}

func (w *dynamicRPCClientWrapper) Close() error {
	w.client.Stop()
	return nil
}

// newConsulStaticRPCClient 使用 Consul 静态解析创建 RPC 客户端（不支持动态更新）
func newConsulStaticRPCClient(consulConf config.ConsulConf, serviceName string, timeout time.Duration) (zrpc.Client, error) {
	adapter := &ioc.ConsulConfigAdapter{
		Address: consulConf.Address,
		Token:   consulConf.Token,
	}

	endpoints, err := ioc.ResolveServiceEndpoints(adapter, serviceName)
	if err != nil {
		logx.Errorf("解析服务端点失败: service=%s, error=%v", serviceName, err)
		return nil, err
	}

	client := zrpc.MustNewClient(zrpc.RpcClientConf{
		Endpoints: endpoints,
		NonBlock:  true,
		Timeout:   int64(timeout / time.Millisecond),
	})

	logx.Infof("成功创建静态 Consul RPC 客户端: service=%s, endpoints=%v", serviceName, endpoints)
	return client, nil
}
