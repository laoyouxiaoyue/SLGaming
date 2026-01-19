package embedder

import (
	"context"
	"fmt"
	"time"

	"github.com/cloudwego/eino-ext/components/embedding/dashscope"
	"github.com/cloudwego/eino/components/embedding"
	"github.com/zeromicro/go-zero/core/logx"
)

// DashScopeEmbedder DashScope 自定义 Embedder
type DashScopeEmbedder struct {
	embedder *dashscope.Embedder
	model    string
	logger   logx.Logger
}

// NewDashScopeEmbedder 创建 DashScope Embedder
func NewDashScopeEmbedder(ctx context.Context, apiKey, modelName string) (*DashScopeEmbedder, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("api key is required")
	}
	if modelName == "" {
		return nil, fmt.Errorf("model name is required")
	}

	// 配置 DashScope Embedding
	embeddingCfg := &dashscope.EmbeddingConfig{
		APIKey:  apiKey,
		Model:   modelName,
		Timeout: 30 * time.Second,
	}

	// 创建 DashScope embedder
	emb, err := dashscope.NewEmbedder(ctx, embeddingCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create dashscope embedder: %w", err)
	}

	return &DashScopeEmbedder{
		embedder: emb,
		model:    modelName,
		logger:   logx.WithContext(ctx),
	}, nil
}

// EmbedStrings 实现 eino embedding.Embedder 接口，对文本列表进行向量化
func (e *DashScopeEmbedder) EmbedStrings(ctx context.Context, texts []string, opts ...embedding.Option) ([][]float64, error) {
	if len(texts) == 0 {
		return nil, fmt.Errorf("texts cannot be empty")
	}

	// 调用 DashScope embedder，传递 options
	vectors, err := e.embedder.EmbedStrings(ctx, texts, opts...)
	if err != nil {
		e.logger.Errorf("dashscope create embeddings failed: %v", err)
		return nil, fmt.Errorf("create embeddings failed: %w", err)
	}

	e.logger.Infof("successfully embedded %d texts using dashscope model %s", len(texts), e.model)
	return vectors, nil
}
