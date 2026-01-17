package logic

import (
	"context"

	"SLGaming/back/services/agent/agent"
	"SLGaming/back/services/agent/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type RecommendCompanionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRecommendCompanionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RecommendCompanionLogic {
	return &RecommendCompanionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 根据用户输入推荐陪玩
func (l *RecommendCompanionLogic) RecommendCompanion(in *agent.RecommendCompanionRequest) (*agent.RecommendCompanionResponse, error) {
	// todo: add your logic here and delete this line

	return &agent.RecommendCompanionResponse{}, nil
}
