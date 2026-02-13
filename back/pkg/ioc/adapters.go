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
		Meta          map[string]string
		CheckInterval string
		CheckTimeout  string
		CheckHTTP     string
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

func (c *ConsulConfigAdapter) GetServiceMeta() map[string]string {
	return c.Service.Meta
}

func (c *ConsulConfigAdapter) GetCheckHTTP() string {
	return c.Service.CheckHTTP
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

// RocketMQConfigAdapter RocketMQ 配置适配器
type RocketMQConfigAdapter struct {
	NameServers []string
	Namespace   string
	AccessKey   string
	SecretKey   string
}

func (r *RocketMQConfigAdapter) GetNameServers() []string {
	return r.NameServers
}

func (r *RocketMQConfigAdapter) GetNamespace() string {
	return r.Namespace
}

func (r *RocketMQConfigAdapter) GetAccessKey() string {
	return r.AccessKey
}

func (r *RocketMQConfigAdapter) GetSecretKey() string {
	return r.SecretKey
}

// RedisConfigAdapter Redis 配置适配器
type RedisConfigAdapter struct {
	Host string
	Type string
	Pass string
	Tls  bool
}

func (r *RedisConfigAdapter) GetHost() string {
	return r.Host
}

func (r *RedisConfigAdapter) GetType() string {
	return r.Type
}

func (r *RedisConfigAdapter) GetPass() string {
	return r.Pass
}

func (r *RedisConfigAdapter) GetTls() bool {
	return r.Tls
}

// MilvusConfigAdapter Milvus 配置适配器
type MilvusConfigAdapter struct {
	Address  string
	Username string
	Password string
	DBName   string
}

func (m *MilvusConfigAdapter) GetAddress() string {
	return m.Address
}

func (m *MilvusConfigAdapter) GetUsername() string {
	return m.Username
}

func (m *MilvusConfigAdapter) GetPassword() string {
	return m.Password
}

func (m *MilvusConfigAdapter) GetDBName() string {
	return m.DBName
}
