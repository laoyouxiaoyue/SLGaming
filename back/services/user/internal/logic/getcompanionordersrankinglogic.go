package logic

import (
	"context"
	"time"

	"SLGaming/back/services/user/internal/helper"
	"SLGaming/back/services/user/internal/metrics"
	"SLGaming/back/services/user/internal/svc"
	"SLGaming/back/services/user/user"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GetCompanionOrdersRankingLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetCompanionOrdersRankingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCompanionOrdersRankingLogic {
	return &GetCompanionOrdersRankingLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetCompanionOrdersRankingLogic) GetCompanionOrdersRanking(in *user.GetCompanionOrdersRankingRequest) (*user.GetCompanionOrdersRankingResponse, error) {
	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime).Seconds()
		metrics.RankingQueryDuration.WithLabelValues("orders").Observe(duration)
	}()

	if l.svcCtx.Redis == nil {
		metrics.RankingQueryTotal.WithLabelValues("orders", "error").Inc()
		return nil, status.Error(codes.FailedPrecondition, "redis not configured")
	}

	// 使用公共排名查询逻辑（ZSet只存前100名）
	result, err := helper.QueryRankingWithProfiles(
		l.ctx,
		l.svcCtx,
		l.Logger,
		"ranking:orders",
		in.GetPage(),
		in.GetPageSize(),
		&helper.OrdersRankingBuilder{},
	)
	if err != nil {
		metrics.RankingQueryTotal.WithLabelValues("orders", "error").Inc()
		return nil, err
	}

	metrics.RankingQueryTotal.WithLabelValues("orders", "success").Inc()

	return &user.GetCompanionOrdersRankingResponse{
		Rankings: result.Rankings,
		Total:    result.Total,
		Page:     result.Page,
		PageSize: result.PageSize,
	}, nil
}
