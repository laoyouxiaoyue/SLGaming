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

type ListGameSkillsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListGameSkillsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListGameSkillsLogic {
	return &ListGameSkillsLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *ListGameSkillsLogic) ListGameSkills(in *user.ListGameSkillsRequest) (*user.ListGameSkillsResponse, error) {
	db := l.svcCtx.DB().WithContext(l.ctx)

	var skills []model.GameSkill
	if err := db.Order("name asc").Find(&skills).Error; err != nil {
		l.Errorf("list game skills failed: %v", err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	resp := make([]*user.GameSkill, 0, len(skills))
	for i := range skills {
		resp = append(resp, &user.GameSkill{
			Id:          skills[i].ID,
			Name:        skills[i].Name,
			Description: skills[i].Description,
		})
	}

	return &user.ListGameSkillsResponse{Skills: resp}, nil
}
