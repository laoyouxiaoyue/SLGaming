package logic

import (
	"context"
	"strings"

	"SLGaming/back/pkg/snowflake"
	"SLGaming/back/services/user/internal/model"
	"SLGaming/back/services/user/internal/svc"
	"SLGaming/back/services/user/user"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CreateGameSkillLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateGameSkillLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateGameSkillLogic {
	return &CreateGameSkillLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *CreateGameSkillLogic) CreateGameSkill(in *user.CreateGameSkillRequest) (*user.CreateGameSkillResponse, error) {
	name := strings.TrimSpace(in.GetName())
	if name == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}

	db := l.svcCtx.DB().WithContext(l.ctx)

	// 检查重名
	var count int64
	if err := db.Model(&model.GameSkill{}).Where("name = ?", name).Count(&count).Error; err != nil {
		l.Errorf("check game skill duplicate failed: %v", err)
		return nil, status.Error(codes.Internal, err.Error())
	}
	if count > 0 {
		return nil, status.Error(codes.AlreadyExists, "skill already exists")
	}

	gs := model.GameSkill{
		BaseModel:   model.BaseModel{ID: uint64(snowflake.GenID())},
		Name:        name,
		Description: strings.TrimSpace(in.GetDescription()),
	}

	if err := db.Create(&gs).Error; err != nil {
		l.Errorf("create game skill failed: %v", err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &user.CreateGameSkillResponse{Skill: &user.GameSkill{
		Id:          gs.ID,
		Name:        gs.Name,
		Description: gs.Description,
	}}, nil
}
