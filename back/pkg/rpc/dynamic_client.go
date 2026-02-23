package rpc

import (
	"sync"
	"time"

	"SLGaming/back/pkg/ioc"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

// DynamicRPCClient 支持动态更新的RPC客户端
//
// 该客户端封装了以下功能：
//   - 服务发现：通过Consul自动发现服务端点
//   - 动态更新：监听服务变化，自动更新连接
//   - 自动重试：内置重试拦截器，对临时故障自动重试
//   - 线程安全：使用读写锁保护并发访问
//
// 工作原理:
//
//	┌─────────────────────────────────────────────────────────┐
//	│                   DynamicRPCClient                      │
//	│  ┌─────────────┐    ┌─────────────┐    ┌────────────┐  │
//	│  │   client    │───►│   watcher   │───►│   Consul   │  │
//	│  │ (zrpc.Client)    │(ConsulWatcher)   │            │  │
//	│  └─────────────┘    └─────────────┘    └────────────┘  │
//	│         │                  │                           │
//	│         │                  │ 服务变化回调               │
//	│         │                  ▼                           │
//	│         │           updateClient()                     │
//	│         │                  │                           │
//	│         ▼                  ▼                           │
//	│  ┌─────────────────────────────────────────────────┐   │
//	│  │           gRPC Connection Pool                   │   │
//	│  │   endpoint1:8080, endpoint2:8080, ...           │   │
//	│  └─────────────────────────────────────────────────┘   │
//	└─────────────────────────────────────────────────────────┘
type DynamicRPCClient struct {
	// mu 保护并发访问 client 和 endpoints
	mu sync.RWMutex

	// client 底层的zrpc客户端
	// 当服务端点变化时会被替换
	client zrpc.Client

	// watcher Consul服务监听器
	// 每5秒检查一次服务端点变化
	watcher *ioc.ConsulWatcher

	// serviceName 服务名称（用于日志和调试）
	serviceName string

	// endpoints 当前服务端点列表
	endpoints []string

	// timeout RPC调用超时时间
	timeout time.Duration

	// retryOpts 重试配置
	retryOpts RetryOptions
}

// DynamicClientOptions 动态RPC客户端配置选项
type DynamicClientOptions struct {
	// ServiceName 服务名称（在Consul中注册的名称）
	// 例如: "user-rpc", "order-rpc"
	ServiceName string

	// Timeout 单次RPC调用的超时时间
	// 默认10秒
	Timeout time.Duration

	// Retry 重试配置
	// 为空时使用默认配置
	Retry RetryOptions
}

// NewDynamicRPCClient 创建支持动态更新的RPC客户端
//
// 创建流程:
//  1. 验证配置参数
//  2. 从Consul解析服务端点（允许失败，服务可能未启动）
//  3. 创建带重试拦截器的zrpc客户端
//  4. 启动Consul监听器，监听服务变化
//
// 使用示例:
//
//	client, err := rpc.NewDynamicRPCClient(consulConfig, rpc.DynamicClientOptions{
//	    ServiceName: "user-rpc",
//	    Timeout:     10 * time.Second,
//	    Retry:       rpc.DefaultRetryOptions(),
//	})
func NewDynamicRPCClient(consulConf ioc.ConsulConfig, opts DynamicClientOptions) (*DynamicRPCClient, error) {
	// 验证必要参数
	if consulConf.GetAddress() == "" || opts.ServiceName == "" {
		return nil, nil
	}

	// 设置默认值
	if opts.Timeout <= 0 {
		opts.Timeout = 10 * time.Second
	}

	if opts.Retry.MaxRetries == 0 {
		opts.Retry = DefaultRetryOptions()
	}

	// 初始解析服务端点（允许失败，服务可能还没启动）
	endpoints, err := ioc.ResolveServiceEndpoints(consulConf, opts.ServiceName)
	if err != nil {
		logx.Infof("[dynamic_rpc] 初始解析服务端点失败（服务可能未启动）: service=%s, error=%v，将等待服务注册后自动发现", opts.ServiceName, err)
		endpoints = []string{}
	}

	// 创建zrpc客户端
	var client zrpc.Client
	if len(endpoints) > 0 {
		client = createClientWithRetry(endpoints, opts.Timeout, opts.Retry)
		logx.Infof("[dynamic_rpc] 创建动态客户端: service=%s, initial_endpoints=%v", opts.ServiceName, endpoints)
	} else {
		// 服务未注册时，创建一个占位客户端
		// 等待服务注册后通过watcher更新
		client = createClientWithRetry([]string{"127.0.0.1:0"}, opts.Timeout, opts.Retry)
		logx.Infof("[dynamic_rpc] 创建动态客户端（等待服务注册）: service=%s", opts.ServiceName)
	}

	// 创建动态客户端实例
	dynamicClient := &DynamicRPCClient{
		client:      client,
		serviceName: opts.ServiceName,
		endpoints:   endpoints,
		timeout:     opts.Timeout,
		retryOpts:   opts.Retry,
	}

	// 创建Consul监听器，监听服务变化
	watcher, err := ioc.NewConsulWatcher(consulConf, opts.ServiceName, func(newEndpoints []string) {
		if len(newEndpoints) > 0 {
			dynamicClient.updateClient(newEndpoints)
		} else {
			logx.Infof("[dynamic_rpc] 服务暂未注册: service=%s，继续等待...", opts.ServiceName)
		}
	})
	if err != nil {
		// watcher创建失败不影响客户端使用，只是无法动态更新
		logx.Errorf("[dynamic_rpc] 创建 Consul watcher 失败: service=%s, error=%v", opts.ServiceName, err)
		return dynamicClient, nil
	}

	dynamicClient.watcher = watcher
	logx.Infof("[dynamic_rpc] 动态客户端创建成功: service=%s (支持动态更新、负载均衡、自动重试)", opts.ServiceName)

	return dynamicClient, nil
}

// createClientWithRetry 创建带重试拦截器的zrpc客户端
//
// 内部函数，用于创建底层的zrpc.Client实例。
// 自动注入重试拦截器，所有RPC调用都会自动重试。
func createClientWithRetry(endpoints []string, timeout time.Duration, retryOpts RetryOptions) zrpc.Client {
	return zrpc.MustNewClient(zrpc.RpcClientConf{
		Endpoints: endpoints,
		NonBlock:  true, // 非阻塞模式，不等待连接建立
		Timeout:   int64(timeout / time.Millisecond),
	}, zrpc.WithUnaryClientInterceptor(UnaryRetryInterceptor(retryOpts)))
}

// GetClient 获取当前的zrpc客户端
// 线程安全
func (d *DynamicRPCClient) GetClient() zrpc.Client {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.client
}

// GetEndpoints 获取当前的服务端点列表
// 线程安全
func (d *DynamicRPCClient) GetEndpoints() []string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.endpoints
}

// GetServiceName 获取服务名称
func (d *DynamicRPCClient) GetServiceName() string {
	return d.serviceName
}

// Conn 获取底层的gRPC连接
// 用于需要直接访问grpc.ClientConn的场景
func (d *DynamicRPCClient) Conn() *grpc.ClientConn {
	return d.GetClient().Conn()
}

// updateClient 更新RPC客户端（当服务端点变化时）
//
// 由ConsulWatcher的回调函数调用，当检测到服务端点变化时：
//  1. 创建新的zrpc客户端（使用新的端点列表）
//  2. 替换旧的客户端
//  3. 旧的连接会被垃圾回收
//
// 注意：这个操作是原子性的，不会影响正在进行的RPC调用
func (d *DynamicRPCClient) updateClient(endpoints []string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	oldEndpoints := d.endpoints

	// 创建新客户端
	d.client = createClientWithRetry(endpoints, d.timeout, d.retryOpts)
	d.endpoints = endpoints

	logx.Infof("[dynamic_rpc] 客户端已更新: service=%s, old_endpoints=%v, new_endpoints=%v",
		d.serviceName, oldEndpoints, endpoints)
}

// Stop 停止动态客户端
// 会停止Consul监听器
func (d *DynamicRPCClient) Stop() {
	if d.watcher != nil {
		d.watcher.Stop()
	}
}

// Close 关闭动态客户端（实现io.Closer接口）
func (d *DynamicRPCClient) Close() error {
	d.Stop()
	return nil
}

// DynamicRPCClientWrapper 包装DynamicRPCClient实现zrpc.Client接口
//
// 这个包装器使得DynamicRPCClient可以无缝替换普通的zrpc.Client。
// 所有生成的RPC客户端代码都接受zrpc.Client接口。
type DynamicRPCClientWrapper struct {
	client *DynamicRPCClient
}

// NewDynamicRPCClientWrapper 创建包装器
func NewDynamicRPCClientWrapper(client *DynamicRPCClient) *DynamicRPCClientWrapper {
	return &DynamicRPCClientWrapper{client: client}
}

// Conn 返回gRPC连接（实现zrpc.Client接口）
func (w *DynamicRPCClientWrapper) Conn() *grpc.ClientConn {
	return w.client.Conn()
}

// Close 关闭客户端（实现zrpc.Client接口）
func (w *DynamicRPCClientWrapper) Close() error {
	return w.client.Close()
}

// GetClient 获取内部的DynamicRPCClient
// 用于需要访问动态客户端特性的场景
func (w *DynamicRPCClientWrapper) GetClient() *DynamicRPCClient {
	return w.client
}

// NewDynamicRPCClientOrFallback 创建动态RPC客户端或返回nil
//
// 这是一个便捷函数，返回zrpc.Client接口类型。
// 如果Consul未配置或服务名为空，返回nil。
//
// 使用示例:
//
//	cli, err := rpc.NewDynamicRPCClientOrFallback(consulConfig, rpc.DynamicClientOptions{
//	    ServiceName: "user-rpc",
//	    Timeout:     10 * time.Second,
//	})
//	if err != nil {
//	    return err
//	}
//	if cli == nil {
//	    // Consul未配置，使用静态端点
//	}
//	userClient := userclient.NewUser(cli)
func NewDynamicRPCClientOrFallback(consulConf ioc.ConsulConfig, opts DynamicClientOptions) (zrpc.Client, error) {
	client, err := NewDynamicRPCClient(consulConf, opts)
	if err != nil {
		return nil, err
	}
	if client == nil {
		return nil, nil
	}
	return NewDynamicRPCClientWrapper(client), nil
}
