package svc

import (
	"context"
	"sync"
	"time"

	"SLGaming/back/services/agent/internal/config"
	"SLGaming/back/services/agent/internal/ioc"

	cli "github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/zeromicro/go-zero/core/logx"
)

type ServiceContext struct {
	mu           sync.RWMutex
	config       config.Config
	MilvusClient cli.Client
}

func NewServiceContext(c config.Config) *ServiceContext {
	ctx := &ServiceContext{
		config: c,
	}

	// 初始化 Milvus 客户端和 Indexer（如果配置了）
	if c.Milvus.Address != "" {
		// 使用带超时的 context，避免阻塞启动
		initCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		milvusClient, err := ioc.InitMilvus(initCtx, c.Milvus)
		if err != nil {
			logx.Errorf("init milvus failed: %v", err)
		} else {
			ctx.MilvusClient = milvusClient
			logx.Infof("Milvus 已初始化, address=%s", c.Milvus.Address)
		}
	} else {
		logx.Infof("Milvus 未配置，向量数据库功能不可用")
	}

	return ctx
}

// Config returns the latest configuration snapshot.
// Config 返回最新的配置快照（复制值，避免外部修改底层状态）。
func (s *ServiceContext) Config() config.Config {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config
}

// UpdateConfig 更新配置。
func (s *ServiceContext) UpdateConfig(newCfg config.Config) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.config = newCfg
	return nil
}
