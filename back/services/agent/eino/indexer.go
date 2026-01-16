package eino

import (
	"context"

	"github.com/cloudwego/eino/components/indexer"
	"github.com/cloudwego/eino/schema"
)

type IndexerImpl struct {
	config *IndexerConfig
}

type IndexerConfig struct {
}

// newIndexer component initialization function of node 'AddCompanionIndexer' in graph 'SLGaming'
func newIndexer(ctx context.Context) (idr indexer.Indexer, err error) {
	// TODO Modify component configuration here.
	config := &IndexerConfig{}
	idr = &IndexerImpl{config: config}
	return idr, nil
}

func (impl *IndexerImpl) Store(ctx context.Context, docs []*schema.Document, opts ...indexer.Option) ([]string, error) {
	panic("implement me")
}
