package logic

import (
	"context"

	"SLGaming/back/services/agent/agent"
	"SLGaming/back/services/agent/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMessageHistoryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetMessageHistoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMessageHistoryLogic {
	return &GetMessageHistoryLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取会话消息历史
func (l *GetMessageHistoryLogic) GetMessageHistory(in *agent.GetMessageHistoryRequest) (*agent.GetMessageHistoryResponse, error) {
	// todo: add your logic here and delete this line

	return &agent.GetMessageHistoryResponse{}, nil
}
