package eino

import (
	"context"

	"github.com/cloudwego/eino-ext/components/embedding/ark"
	"github.com/cloudwego/eino/components/embedding"
)

// newEmbedding component initialization function of node 'AddCompanionEmbedding' in graph 'SLGaming'
func newEmbedding(ctx context.Context) (eb embedding.Embedder, err error) {
	// TODO Modify component configuration here.
	config := &ark.EmbeddingConfig{}
	eb, err = ark.NewEmbedder(ctx, config)
	if err != nil {
		return nil, err
	}
	return eb, nil
}
