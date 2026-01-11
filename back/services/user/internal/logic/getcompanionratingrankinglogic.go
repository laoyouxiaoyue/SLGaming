package logic

import (
	"context"

	"SLGaming/back/services/user/internal/helper"
	"SLGaming/back/services/user/internal/svc"
	"SLGaming/back/services/user/user"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GetCompanionRatingRankingLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetCompanionRatingRankingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCompanionRatingRankingLogic {
	return &GetCompanionRatingRankingLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 陪玩排名相关接口
func (l *GetCompanionRatingRankingLogic) GetCompanionRatingRanking(in *user.GetCompanionRatingRankingRequest) (*user.GetCompanionRatingRankingResponse, error) {
	if l.svcCtx.Redis == nil {
		return nil, status.Error(codes.FailedPrecondition, "redis not configured")
	}

	// 规范化分页参数
	pagination := helper.NormalizePagination(in.GetPage(), in.GetPageSize())

	// 使用公共排名查询逻辑
	result, err := helper.QueryRankingWithProfiles(
		l.ctx,
		l.svcCtx,
		l.Logger,
		"ranking:rating",
		pagination,
		&helper.RatingRankingBuilder{},
	)
	if err != nil {
		return nil, err
	}

	return &user.GetCompanionRatingRankingResponse{
		Rankings: result.Rankings,
		Total:    result.Total,
		Page:     result.Page,
		PageSize: result.PageSize,
	}, nil
}
