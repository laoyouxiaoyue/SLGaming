package embedder

import (
	"context"
	"fmt"

	"github.com/volcengine/volcengine-go-sdk/service/arkruntime"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime/model"
	"github.com/zeromicro/go-zero/core/logx"
)

// VolcengineEmbedder 火山引擎自定义 Embedder
type VolcengineEmbedder struct {
	client *arkruntime.Client
	model  string
	logger logx.Logger
}

// NewVolcengineEmbedder 创建火山引擎 Embedder
func NewVolcengineEmbedder(ctx context.Context, apiKey, modelName string) (*VolcengineEmbedder, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("api key is required")
	}
	if modelName == "" {
		return nil, fmt.Errorf("model name is required")
	}

	client := arkruntime.NewClientWithApiKey(apiKey)

	return &VolcengineEmbedder{
		client: client,
		model:  modelName,
		logger: logx.WithContext(ctx),
	}, nil
}

// EmbedStrings 实现 eino schema.Embedder 接口，对文本列表进行向量化
func (e *VolcengineEmbedder) EmbedStrings(ctx context.Context, texts []string) ([][]float64, error) {
	if len(texts) == 0 {
		return nil, fmt.Errorf("texts cannot be empty")
	}

	// 构建请求
	req := model.EmbeddingRequest{
		Model: e.model,
		Input: texts,
	}

	// 调用火山引擎 API
	resp, err := e.client.CreateEmbeddings(ctx, req)
	if err != nil {
		e.logger.Errorf("volcengine create embeddings failed: %v", err)
		return nil, fmt.Errorf("create embeddings failed: %w", err)
	}

	// 检查响应
	if resp.Data == nil || len(resp.Data) == 0 {
		return nil, fmt.Errorf("empty embedding response")
	}

	// 转换响应格式：将 []float32 转换为 []float64
	results := make([][]float64, len(resp.Data))
	for i, embedding := range resp.Data {
		if embedding.Embedding == nil {
			return nil, fmt.Errorf("embedding data is nil at index %d", i)
		}

		// 将 float32 向量转换为 float64
		vec := make([]float64, len(embedding.Embedding))
		for j, v := range embedding.Embedding {
			vec[j] = float64(v)
		}
		results[i] = vec
	}

	e.logger.Infof("successfully embedded %d texts using model %s", len(texts), e.model)
	return results, nil
}
