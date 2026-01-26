package logic

import (
	"context"

	"SLGaming/back/services/user/internal/model"
	"SLGaming/back/services/user/internal/svc"
	"SLGaming/back/services/user/user"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type DeleteGameSkillLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteGameSkillLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteGameSkillLogic {
	return &DeleteGameSkillLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *DeleteGameSkillLogic) DeleteGameSkill(in *user.DeleteGameSkillRequest) (*user.DeleteGameSkillResponse, error) {
	if in.GetId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	db := l.svcCtx.DB().WithContext(l.ctx)

	res := db.Delete(&model.GameSkill{}, in.GetId())
	if res.Error != nil {
		l.Errorf("delete game skill failed: %v", res.Error)
		return nil, status.Error(codes.Internal, res.Error.Error())
	}
	if res.RowsAffected == 0 {
		return nil, status.Error(codes.NotFound, "game skill not found")
	}

	return &user.DeleteGameSkillResponse{Success: true}, nil
}
