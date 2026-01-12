package ioc

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/zeromicro/go-zero/core/logx"
)

// ConsulWatcher Consul 服务监听器
type ConsulWatcher struct {
	client      *api.Client
	serviceName string
	endpoints   []string
	mu          sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
	onChange    func([]string)
}

// NewConsulWatcher 创建 Consul 服务监听器
func NewConsulWatcher(cfg ConsulConfig, serviceName string, onChange func([]string)) (*ConsulWatcher, error) {
	if cfg.GetAddress() == "" || serviceName == "" {
		return nil, fmt.Errorf("consul address or service name is empty")
	}

	client, err := api.NewClient(&api.Config{
		Address: cfg.GetAddress(),
		Token:   cfg.GetToken(),
	})
	if err != nil {
		return nil, fmt.Errorf("create consul client: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	watcher := &ConsulWatcher{
		client:      client,
		serviceName: serviceName,
		ctx:         ctx,
		cancel:      cancel,
		onChange:    onChange,
	}

	// 初始解析一次
	endpoints, err := watcher.resolveEndpoints()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("initial resolve endpoints failed: %w", err)
	}
	watcher.setEndpoints(endpoints)

	// 启动监听
	go watcher.watch()

	return watcher, nil
}

// GetEndpoints 获取当前服务端点列表
func (w *ConsulWatcher) GetEndpoints() []string {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.endpoints
}

// Stop 停止监听
func (w *ConsulWatcher) Stop() {
	w.cancel()
}

// setEndpoints 设置服务端点列表
func (w *ConsulWatcher) setEndpoints(endpoints []string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.endpoints = endpoints
}

// resolveEndpoints 解析服务端点
func (w *ConsulWatcher) resolveEndpoints() ([]string, error) {
	// 先尝试通过健康检查查询
	services, _, err := w.client.Health().Service(w.serviceName, "", false, nil)
	if err != nil {
		logx.Errorf("[consul_watch] 失败: 从Consul查询服务失败, service=%s, error=%v", w.serviceName, err)
		return nil, fmt.Errorf("query consul service: %w", err)
	}

	var endpoints []string
	if len(services) > 0 {
		for _, service := range services {
			address := service.Service.Address
			port := service.Service.Port
			if address == "" {
				address = "127.0.0.1"
			}
			endpoint := fmt.Sprintf("%s:%d", address, port)
			endpoints = append(endpoints, endpoint)
		}
		logx.Infof("[consul_watch] 成功: 从健康检查解析服务端点, service=%s, endpoints=%v", w.serviceName, endpoints)
		return endpoints, nil
	}

	// 如果健康检查找不到，从 Agent 查询
	logx.Infof("[consul_watch] 信息: 通过健康检查未找到服务，尝试从Agent查询, service=%s", w.serviceName)
	agentServices, err := w.client.Agent().Services()
	if err != nil {
		logx.Errorf("[consul_watch] 失败: 查询Agent服务失败, service=%s, error=%v", w.serviceName, err)
		return nil, fmt.Errorf("query consul agent services: %w", err)
	}

	for _, svc := range agentServices {
		if svc.Service == w.serviceName {
			address := svc.Address
			if address == "" {
				address = "127.0.0.1"
			}
			endpoint := fmt.Sprintf("%s:%d", address, svc.Port)
			endpoints = append(endpoints, endpoint)
		}
	}

	if len(endpoints) > 0 {
		logx.Infof("[consul_watch] 成功: 从Agent解析服务端点, service=%s, endpoints=%v", w.serviceName, endpoints)
		return endpoints, nil
	}

	logx.Errorf("[consul_watch] 失败: 在Consul中未找到服务, service=%s", w.serviceName)
	return nil, fmt.Errorf("service %s not found in consul", w.serviceName)
}

// watch 监听服务变化
func (w *ConsulWatcher) watch() {
	ticker := time.NewTicker(5 * time.Second) // 每5秒检查一次
	defer ticker.Stop()

	for {
		select {
		case <-w.ctx.Done():
			logx.Infof("[consul_watch] 停止监听: service=%s", w.serviceName)
			return
		case <-ticker.C:
			endpoints, err := w.resolveEndpoints()
			if err != nil {
				logx.Errorf("[consul_watch] 失败: 解析服务端点失败, service=%s, error=%v", w.serviceName, err)
				continue
			}

			// 检查是否有变化
			currentEndpoints := w.GetEndpoints()
			if !endpointsEqual(currentEndpoints, endpoints) {
				logx.Infof("[consul_watch] 检测到服务变化: service=%s, old=%v, new=%v", w.serviceName, currentEndpoints, endpoints)
				w.setEndpoints(endpoints)
				if w.onChange != nil {
					w.onChange(endpoints)
				}
			}
		}
	}
}

// endpointsEqual 比较两个端点列表是否相等
func endpointsEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	amap := make(map[string]bool)
	for _, e := range a {
		amap[e] = true
	}
	for _, e := range b {
		if !amap[e] {
			return false
		}
	}
	return true
}
