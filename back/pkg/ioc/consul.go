package ioc

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/consul/api"
	"github.com/zeromicro/go-zero/core/logx"
)

// getLocalIP 获取本机的第一个非回环 IPv4 地址
func getLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", fmt.Errorf("get interface addrs: %w", err)
	}

	for _, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		if !ok {
			continue
		}

		ip := ipNet.IP
		// 跳过回环地址和 IPv6 地址
		if ip.IsLoopback() || ip.To4() == nil {
			continue
		}

		return ip.String(), nil
	}

	return "", fmt.Errorf("no non-loopback IPv4 address found")
}

// getPublicIP 获取公网 IP 地址
// 通过调用外部 API 获取，如果失败则返回错误
func getPublicIP() (string, error) {
	// 使用多个 API 作为备选，提高成功率
	apis := []string{
		"https://api.ipify.org?format=text",
		"https://ifconfig.me/ip",
		"https://api.ip.sb/ip",
		"https://icanhazip.com",
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	var lastErr error
	for _, apiURL := range apis {
		resp, err := client.Get(apiURL)
		if err != nil {
			lastErr = err
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("API returned status %d", resp.StatusCode)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			lastErr = err
			continue
		}

		ip := strings.TrimSpace(string(body))
		// 验证是否是有效的 IP 地址
		if parsedIP := net.ParseIP(ip); parsedIP != nil && parsedIP.To4() != nil {
			return ip, nil
		}
		lastErr = fmt.Errorf("invalid IP format: %s", ip)
	}

	return "", fmt.Errorf("failed to get public IP from all APIs: %w", lastErr)
}

// getPublicIPWithFallback 获取公网 IP，如果失败则回退到内网 IP
func getPublicIPWithFallback() (string, error) {
	publicIP, err := getPublicIP()
	if err == nil {
		logx.Infof("[ip_detect] 成功: 获取到公网IP, ip=%s", publicIP)
		return publicIP, nil
	}

	logx.Infof("[ip_detect] 警告: 获取公网IP失败, error=%v, 回退到内网IP", err)
	localIP, err := getLocalIP()
	if err != nil {
		return "", fmt.Errorf("failed to get both public and local IP: public=%v, local=%w", err, err)
	}

	logx.Infof("[ip_detect] 信息: 使用内网IP, ip=%s", localIP)
	return localIP, nil
}

// ConsulRegistrar Consul 服务注册器
type ConsulRegistrar struct {
	client    *api.Client
	serviceID string
}

// Deregister 注销服务
func (r *ConsulRegistrar) Deregister() {
	if r == nil {
		return
	}
	if err := r.client.Agent().ServiceDeregister(r.serviceID); err != nil {
		logx.Errorf("[consul_deregister] 失败: 服务注销失败, service_id=%s, error=%v", r.serviceID, err)
	} else {
		logx.Infof("[consul_deregister] 成功: 服务已注销, service_id=%s", r.serviceID)
	}
}

// RegisterConsul 注册服务到 Consul
// listenOn: 服务监听地址，格式如 "0.0.0.0:8080" 或 "127.0.0.1:8080"
// checkType: 健康检查类型，"http" 或 "grpc"
func RegisterConsul(cfg ConsulConfig, listenOn string, checkType string) (*ConsulRegistrar, error) {
	if cfg.GetAddress() == "" || cfg.GetServiceName() == "" {
		return nil, nil
	}

	host, portStr, err := net.SplitHostPort(listenOn)
	if err != nil {
		return nil, fmt.Errorf("split listen address: %w", err)
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("invalid listen port: %w", err)
	}

	// 确定服务注册地址
	serviceAddr := cfg.GetServiceAddress()
	if strings.TrimSpace(serviceAddr) == "" || serviceAddr == "0.0.0.0" || serviceAddr == "127.0.0.1" {
		// 如果配置为空、0.0.0.0 或 127.0.0.1，优先获取公网 IP
		if host == "" || host == "0.0.0.0" {
			// 优先尝试获取公网 IP，失败则回退到内网 IP
			ip, err := getPublicIPWithFallback()
			if err != nil {
				logx.Infof("[consul_register] 警告: 获取IP地址失败, error=%v, 使用回退IP 127.0.0.1", err)
				serviceAddr = "127.0.0.1"
			} else {
				serviceAddr = ip
			}
		} else {
			serviceAddr = host
		}
	}

	serviceID := cfg.GetServiceID()
	if serviceID == "" {
		serviceID = fmt.Sprintf("%s-%s", cfg.GetServiceName(), uuid.NewString())
	}

	client, err := api.NewClient(&api.Config{
		Address: cfg.GetAddress(),
		Token:   cfg.GetToken(),
	})
	if err != nil {
		return nil, fmt.Errorf("create consul client: %w", err)
	}

	checkInterval := cfg.GetCheckInterval()
	if checkInterval == "" {
		checkInterval = "10s"
	}
	checkTimeout := cfg.GetCheckTimeout()
	if checkTimeout == "" {
		checkTimeout = "5s"
	}

	// 根据服务类型设置健康检查
	var check *api.AgentServiceCheck
	if checkType == "http" {
		check = &api.AgentServiceCheck{
			HTTP:     fmt.Sprintf("http://%s:%d/health", serviceAddr, port),
			Interval: checkInterval,
			Timeout:  checkTimeout,
		}
	} else {
		// 默认使用 gRPC 健康检查
		check = &api.AgentServiceCheck{
			GRPC:     fmt.Sprintf("%s:%d", serviceAddr, port),
			Interval: checkInterval,
			Timeout:  checkTimeout,
		}
	}

	reg := &api.AgentServiceRegistration{
		ID:      serviceID,
		Name:    cfg.GetServiceName(),
		Address: serviceAddr,
		Port:    port,
		Tags:    cfg.GetServiceTags(),
		Check:   check,
	}

	if err := client.Agent().ServiceRegister(reg); err != nil {
		logx.Errorf("[consul_register] 失败: 服务注册失败, service=%s, address=%s:%d, error=%v", reg.Name, serviceAddr, port, err)
		return nil, fmt.Errorf("register consul service: %w", err)
	}
	logx.Infof("[consul_register] 成功: 服务已注册, service=%s, service_id=%s, address=%s:%d, check_type=%s", reg.Name, serviceID, serviceAddr, port, checkType)

	return &ConsulRegistrar{
		client:    client,
		serviceID: serviceID,
	}, nil
}

// ResolveServiceEndpoints 通过 Consul 服务发现解析服务端点
func ResolveServiceEndpoints(cfg ConsulConfig, serviceName string) ([]string, error) {
	if cfg.GetAddress() == "" || serviceName == "" {
		return nil, fmt.Errorf("consul address or service name is empty")
	}

	client, err := api.NewClient(&api.Config{
		Address: cfg.GetAddress(),
		Token:   cfg.GetToken(),
	})
	if err != nil {
		logx.Errorf("[consul_resolve] 失败: 创建Consul客户端失败, service=%s, consul_address=%s, error=%v", serviceName, cfg.GetAddress(), err)
		return nil, fmt.Errorf("create consul client: %w", err)
	}

	// 先尝试查询所有服务（包括不健康的），因为本地运行的服务可能无法通过服务器上的 Consul 健康检查
	services, _, err := client.Health().Service(serviceName, "", false, nil)
	if err != nil {
		logx.Errorf("[consul_resolve] 失败: 从Consul查询服务失败, service=%s, error=%v", serviceName, err)
		return nil, fmt.Errorf("query consul service: %w", err)
	}

	// 如果通过健康检查接口找不到服务，尝试直接从 Agent 查询所有已注册的服务
	if len(services) == 0 {
		logx.Infof("[consul_resolve] 信息: 通过健康检查未找到服务，尝试从Agent查询, service=%s", serviceName)
		agentServices, err := client.Agent().Services()
		if err != nil {
			logx.Errorf("[consul_resolve] 失败: 查询Agent服务失败, service=%s, error=%v", serviceName, err)
			return nil, fmt.Errorf("query consul agent services: %w", err)
		}

		var endpoints []string
		for _, svc := range agentServices {
			if svc.Service == serviceName {
				address := svc.Address
				if address == "" {
					address = "127.0.0.1"
				}
				endpoint := fmt.Sprintf("%s:%d", address, svc.Port)
				endpoints = append(endpoints, endpoint)
			}
		}

		if len(endpoints) > 0 {
			logx.Infof("[consul_resolve] 成功: 从Agent解析服务端点, service=%s, endpoints=%v", serviceName, endpoints)
			return endpoints, nil
		}

		logx.Errorf("[consul_resolve] 失败: 在Consul中未找到服务, service=%s", serviceName)
		return nil, fmt.Errorf("service %s not found in consul", serviceName)
	}

	var endpoints []string
	for _, service := range services {
		address := service.Service.Address
		port := service.Service.Port
		if address == "" {
			address = "127.0.0.1"
		}
		endpoint := fmt.Sprintf("%s:%d", address, port)
		endpoints = append(endpoints, endpoint)
	}

	logx.Infof("[consul_resolve] 成功: 从健康检查解析服务端点, service=%s, endpoints=%v", serviceName, endpoints)
	return endpoints, nil
}
