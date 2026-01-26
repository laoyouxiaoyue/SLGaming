package logic

import (
	"context"
	"strings"

	"SLGaming/back/services/user/internal/model"
	"SLGaming/back/services/user/internal/svc"
	"SLGaming/back/services/user/user"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type UpdateGameSkillLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateGameSkillLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateGameSkillLogic {
	return &UpdateGameSkillLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *UpdateGameSkillLogic) UpdateGameSkill(in *user.UpdateGameSkillRequest) (*user.UpdateGameSkillResponse, error) {
	if in.GetId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	name := strings.TrimSpace(in.GetName())
	if name == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}

	db := l.svcCtx.DB().WithContext(l.ctx)

	var gs model.GameSkill
	if err := db.First(&gs, "id = ?", in.GetId()).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Error(codes.NotFound, "game skill not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	// 若改名，需检查唯一
	if gs.Name != name {
		var count int64
		if err := db.Model(&model.GameSkill{}).Where("name = ?", name).Count(&count).Error; err != nil {
			l.Errorf("check duplicate on update failed: %v", err)
			return nil, status.Error(codes.Internal, err.Error())
		}
		if count > 0 {
			return nil, status.Error(codes.AlreadyExists, "skill already exists")
		}
	}

	gs.Name = name
	gs.Description = strings.TrimSpace(in.GetDescription())

	if err := db.Save(&gs).Error; err != nil {
		l.Errorf("update game skill failed: %v", err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &user.UpdateGameSkillResponse{Skill: &user.GameSkill{
		Id:          gs.ID,
		Name:        gs.Name,
		Description: gs.Description,
	}}, nil
}
