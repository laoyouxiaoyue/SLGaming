package logic

import (
	"context"

	"SLGaming/back/services/agent/agent"
	"SLGaming/back/services/agent/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type SendMessageLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSendMessageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendMessageLogic {
	return &SendMessageLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 发送用户消息，返回 Agent 回复
func (l *SendMessageLogic) SendMessage(in *agent.SendMessageRequest) (*agent.SendMessageResponse, error) {
	// todo: add your logic here and delete this line

	return &agent.SendMessageResponse{}, nil
}
