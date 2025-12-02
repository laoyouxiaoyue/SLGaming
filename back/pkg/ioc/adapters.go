package ioc

import "time"

// ConsulConfigAdapter Consul 配置适配器
type ConsulConfigAdapter struct {
	Address string
	Token   string
	Service struct {
		Name          string
		ID            string
		Address       string
		Tags          []string
		CheckInterval string
		CheckTimeout  string
	}
}

func (c *ConsulConfigAdapter) GetAddress() string {
	return c.Address
}

func (c *ConsulConfigAdapter) GetToken() string {
	return c.Token
}

func (c *ConsulConfigAdapter) GetServiceName() string {
	return c.Service.Name
}

func (c *ConsulConfigAdapter) GetServiceID() string {
	return c.Service.ID
}

func (c *ConsulConfigAdapter) GetServiceAddress() string {
	return c.Service.Address
}

func (c *ConsulConfigAdapter) GetServiceTags() []string {
	return c.Service.Tags
}

func (c *ConsulConfigAdapter) GetCheckInterval() string {
	return c.Service.CheckInterval
}

func (c *ConsulConfigAdapter) GetCheckTimeout() string {
	return c.Service.CheckTimeout
}

// NacosConfigAdapter Nacos 配置适配器
type NacosConfigAdapter struct {
	Hosts     []string
	Namespace string
	Group     string
	DataId    string
	Username  string
	Password  string
}

func (n *NacosConfigAdapter) GetHosts() []string {
	return n.Hosts
}

func (n *NacosConfigAdapter) GetNamespace() string {
	return n.Namespace
}

func (n *NacosConfigAdapter) GetGroup() string {
	return n.Group
}

func (n *NacosConfigAdapter) GetDataId() string {
	return n.DataId
}

func (n *NacosConfigAdapter) GetUsername() string {
	return n.Username
}

func (n *NacosConfigAdapter) GetPassword() string {
	return n.Password
}

// MySQLConfigAdapter MySQL 配置适配器
type MySQLConfigAdapter struct {
	DSN             string
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

func (m *MySQLConfigAdapter) GetDSN() string {
	return m.DSN
}

func (m *MySQLConfigAdapter) GetMaxIdleConns() int {
	return m.MaxIdleConns
}

func (m *MySQLConfigAdapter) GetMaxOpenConns() int {
	return m.MaxOpenConns
}

func (m *MySQLConfigAdapter) GetConnMaxLifetime() time.Duration {
	return m.ConnMaxLifetime
}

func (m *MySQLConfigAdapter) GetConnMaxIdleTime() time.Duration {
	return m.ConnMaxIdleTime
}
