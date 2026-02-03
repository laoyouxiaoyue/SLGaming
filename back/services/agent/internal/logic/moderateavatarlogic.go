package logic

import (
	"context"

	"SLGaming/back/services/agent/agent"
	"SLGaming/back/services/agent/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ModerateAvatarLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewModerateAvatarLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ModerateAvatarLogic {
	return &ModerateAvatarLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 头像多模态审核
func (l *ModerateAvatarLogic) ModerateAvatar(in *agent.ModerateAvatarRequest) (*agent.ModerateAvatarResponse, error) {
	// todo: add your logic here and delete this line

	return &agent.ModerateAvatarResponse{}, nil
}
