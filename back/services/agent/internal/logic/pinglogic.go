package logic

import (
	"context"

	"SLGaming/back/services/agent/agent"
	"SLGaming/back/services/agent/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type PingLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PingLogic {
	return &PingLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 健康检查
func (l *PingLogic) Ping(in *agent.PingRequest) (*agent.PingResponse, error) {
	// todo: add your logic here and delete this line

	return &agent.PingResponse{}, nil
}
