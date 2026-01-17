package ioc

import (
	"context"
	"fmt"

	cli "github.com/milvus-io/milvus-sdk-go/v2/client"
)

// MilvusConfig 定义 Milvus 配置接口
type MilvusConfig interface {
	GetAddress() string
	GetUsername() string
	GetPassword() string
	GetDBName() string
}

// InitMilvus 根据配置初始化 Milvus 客户端
func InitMilvus(ctx context.Context, cfg MilvusConfig) (cli.Client, error) {
	address := cfg.GetAddress()
	if address == "" {
		return nil, fmt.Errorf("milvus address is empty")
	}

	clientConfig := cli.Config{
		Address: address,
		DBName:  cfg.GetDBName(),
	}

	// 如果配置了用户名和密码，设置认证
	if username := cfg.GetUsername(); username != "" {
		clientConfig.Username = username
		clientConfig.Password = cfg.GetPassword()
	}

	milvusClient, err := cli.NewClient(ctx, clientConfig)
	if err != nil {
		return nil, fmt.Errorf("new milvus client: %w", err)
	}

	return milvusClient, nil
}
