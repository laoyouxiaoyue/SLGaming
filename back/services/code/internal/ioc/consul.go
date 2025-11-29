package ioc

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"SLGaming/back/services/code/internal/config"

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

type consulRegistrar struct {
	client    *api.Client
	serviceID string
}

func RegisterConsul(cfg config.ConsulConf, listenOn string) (*consulRegistrar, error) {
	if cfg.Address == "" || cfg.Service.Name == "" {
		return nil, nil
	}

	host, portStr, err := net.SplitHostPort(listenOn)
	if err != nil {
		logx.Errorf("解析监听地址失败: listenOn=%s, error=%v", listenOn, err)
		return nil, fmt.Errorf("split listen address: %w", err)
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		logx.Errorf("无效的端口号: port=%s, error=%v", portStr, err)
		return nil, fmt.Errorf("invalid listen port: %w", err)
	}

	// 确定服务注册地址
	serviceAddr := cfg.Service.Address
	if strings.TrimSpace(serviceAddr) == "" || serviceAddr == "0.0.0.0" {
		// 如果配置为空或者是 0.0.0.0，需要获取实际 IP 地址
		if host == "" || host == "0.0.0.0" {
			// 获取本机实际 IP 地址
			localIP, err := getLocalIP()
			if err != nil {
				logx.Infof("获取本机 IP 地址失败: %v, 使用 127.0.0.1 作为回退", err)
				serviceAddr = "127.0.0.1"
			} else {
				serviceAddr = localIP
			}
		} else {
			serviceAddr = host
		}
	}

	serviceID := cfg.Service.ID
	if serviceID == "" {
		serviceID = fmt.Sprintf("%s-%s", cfg.Service.Name, uuid.NewString())
	}

	client, err := api.NewClient(&api.Config{
		Address: cfg.Address,
		Token:   cfg.Token,
	})
	if err != nil {
		logx.Errorf("创建 Consul 客户端失败: consul=%s, error=%v", cfg.Address, err)
		return nil, fmt.Errorf("create consul client: %w", err)
	}

	checkInterval := cfg.Service.CheckInterval
	if checkInterval == "" {
		checkInterval = "10s"
	}
	checkTimeout := cfg.Service.CheckTimeout
	if checkTimeout == "" {
		checkTimeout = "5s"
	}

	reg := &api.AgentServiceRegistration{
		ID:      serviceID,
		Name:    cfg.Service.Name,
		Address: serviceAddr,
		Port:    port,
		Tags:    cfg.Service.Tags,
		Check: &api.AgentServiceCheck{
			GRPC:     fmt.Sprintf("%s:%d", serviceAddr, port),
			Interval: checkInterval,
			Timeout:  checkTimeout,
		},
	}

	if err := client.Agent().ServiceRegister(reg); err != nil {
		logx.Errorf("注册服务到 Consul 失败: service=%s, error=%v", reg.Name, err)
		return nil, fmt.Errorf("register consul service: %w", err)
	}
	logx.Infof("成功注册服务到 Consul: service=%s, address=%s:%d", reg.Name, serviceAddr, port)

	return &consulRegistrar{
		client:    client,
		serviceID: serviceID,
	}, nil
}

func (r *consulRegistrar) Deregister() {
	if r == nil {
		return
	}
	if err := r.client.Agent().ServiceDeregister(r.serviceID); err != nil {
		logx.Errorf("注销 Consul 服务失败: service_id=%s, error=%v", r.serviceID, err)
	}
}
