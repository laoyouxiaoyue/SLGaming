package ioc

import (
	"context"

	"SLGaming/back/pkg/ioc"
	"SLGaming/back/services/agent/internal/config"

	cli "github.com/milvus-io/milvus-sdk-go/v2/client"
)

// InitMilvus 根据配置初始化 Milvus 客户端
func InitMilvus(ctx context.Context, cfg config.MilvusConf) (cli.Client, error) {
	dbName := cfg.DBName
	if dbName == "" {
		dbName = "default"
	}

	adapter := &ioc.MilvusConfigAdapter{
		Address:  cfg.Address,
		Username: cfg.Username,
		Password: cfg.Password,
		DBName:   dbName,
	}
	return ioc.InitMilvus(ctx, adapter)
}
