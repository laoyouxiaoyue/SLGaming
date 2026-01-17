package config

import "github.com/zeromicro/go-zero/zrpc"

// NacosConf Nacos 配置结构
type NacosConf struct {
	Hosts     []string `json:",optional"` // Nacos 服务器地址列表
	Namespace string   `json:",optional"` // 命名空间
	Group     string   `json:",optional"` // 配置组
	DataId    string   `json:",optional"` // 配置 DataId
	Username  string   `json:",optional"` // 用户名
	Password  string   `json:",optional"` // 密码
}

type ConsulConf struct {
	Address string            `json:",optional"`
	Token   string            `json:",optional"`
	Service ConsulServiceConf `json:",optional"`
}

type ConsulServiceConf struct {
	Name          string   `json:",optional"`
	ID            string   `json:",optional"`
	Address       string   `json:",optional"`
	Tags          []string `json:",optional"`
	CheckInterval string   `json:",optional"`
	CheckTimeout  string   `json:",optional"`
}

// LLMConf 大模型配置
type LLMConf struct {
	APIKey string `json:",optional"` // API Key，用于调用大模型服务
	Model  string `json:",optional"` // 模型名称，如 "gpt-4", "gpt-3.5-turbo", "claude-3", "qwen-plus" 等
}

// MilvusConf Milvus 向量数据库配置
type MilvusConf struct {
	Address  string `json:",optional"` // Milvus 服务地址，格式: host:port，如 "127.0.0.1:19530"
	Username string `json:",optional"` // 用户名（如果开启认证）
	Password string `json:",optional"` // 密码（如果开启认证）
	DBName   string `json:",optional"` // 数据库名称，默认为 "default"
}

type Config struct {
	zrpc.RpcServerConf
	Nacos  NacosConf  `json:",optional"` // Nacos 配置
	Consul ConsulConf `json:",optional"` // Consul 配置
	LLM    LLMConf    `json:",optional"` // 大模型配置
	Milvus MilvusConf `json:",optional"` // Milvus 向量数据库配置
}
