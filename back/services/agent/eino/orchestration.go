package eino

import (
	"context"

	"github.com/cloudwego/eino/compose"
)

func BuildSLGaming(ctx context.Context) (r compose.Runnable[string, string], err error) {
	const (
		MasterNode            = "MasterNode"
		RecommendChat         = "RecommendChat"
		RecommendNode         = "RecommendNode"
		ToolsNode5            = "ToolsNode5"
		AddCompanionEmbedding = "AddCompanionEmbedding"
		AddCompanionIndexer   = "AddCompanionIndexer"
		MasterChatTemplate    = "MasterChatTemplate"
	)
	g := compose.NewGraph[string, string]()
	masterNodeKeyOfChatModel, err := newChatModel(ctx)
	if err != nil {
		return nil, err
	}
	_ = g.AddChatModelNode(MasterNode, masterNodeKeyOfChatModel)
	recommendChatKeyOfChatModel, err := newChatModel1(ctx)
	if err != nil {
		return nil, err
	}
	_ = g.AddChatModelNode(RecommendChat, recommendChatKeyOfChatModel)
	recommendNodeKeyOfRetriever, err := newRetriever(ctx)
	if err != nil {
		return nil, err
	}
	_ = g.AddRetrieverNode(RecommendNode, recommendNodeKeyOfRetriever)
	toolsNode5KeyOfToolsNode, err := newToolsNode(ctx)
	if err != nil {
		return nil, err
	}
	_ = g.AddToolsNode(ToolsNode5, toolsNode5KeyOfToolsNode)
	addCompanionEmbeddingKeyOfEmbedding, err := newEmbedding(ctx)
	if err != nil {
		return nil, err
	}
	_ = g.AddEmbeddingNode(AddCompanionEmbedding, addCompanionEmbeddingKeyOfEmbedding)
	addCompanionIndexerKeyOfIndexer, err := newIndexer(ctx)
	if err != nil {
		return nil, err
	}
	_ = g.AddIndexerNode(AddCompanionIndexer, addCompanionIndexerKeyOfIndexer)
	masterChatTemplateKeyOfChatTemplate, err := newChatTemplate(ctx)
	if err != nil {
		return nil, err
	}
	_ = g.AddChatTemplateNode(MasterChatTemplate, masterChatTemplateKeyOfChatTemplate)
	_ = g.AddEdge(compose.START, MasterChatTemplate)
	_ = g.AddEdge(RecommendChat, compose.END)
	_ = g.AddEdge(ToolsNode5, compose.END)
	_ = g.AddEdge(AddCompanionIndexer, compose.END)
	_ = g.AddEdge(MasterChatTemplate, MasterNode)
	_ = g.AddEdge(RecommendNode, RecommendChat)
	_ = g.AddEdge(AddCompanionEmbedding, AddCompanionIndexer)
	_ = g.AddBranch(MasterNode, compose.NewGraphBranch(newBranch, map[string]bool{compose.END: true, RecommendNode: true, ToolsNode5: true, AddCompanionEmbedding: true}))
	r, err = g.Compile(ctx, compose.WithGraphName("SLGaming"), compose.WithNodeTriggerMode(compose.AnyPredecessor))
	if err != nil {
		return nil, err
	}
	return r, err
}
