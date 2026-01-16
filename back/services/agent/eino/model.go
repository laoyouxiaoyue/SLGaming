package eino

import (
	"context"

	"github.com/cloudwego/eino-ext/components/model/ark"
	"github.com/cloudwego/eino/components/model"
)

// newChatModel component initialization function of node 'MasterNode' in graph 'SLGaming'
func newChatModel(ctx context.Context) (cm model.ChatModel, err error) {
	// TODO Modify component configuration here.
	config := &ark.ChatModelConfig{
		APIKey: "a9d50ceb-485a-4a2e-9042-5a7c87b24b57",
		Model:  "doubao-seed-1-8-251228"}
	cm, err = ark.NewChatModel(ctx, config)
	if err != nil {
		return nil, err
	}
	return cm, nil
}

// newChatModel1 component initialization function of node 'RecommendChat' in graph 'SLGaming'
func newChatModel1(ctx context.Context) (cm model.ChatModel, err error) {
	// TODO Modify component configuration here.
	config := &ark.ChatModelConfig{
		APIKey: "a9d50ceb-485a-4a2e-9042-5a7c87b24b57"}
	cm, err = ark.NewChatModel(ctx, config)
	if err != nil {
		return nil, err
	}
	return cm, nil
}
