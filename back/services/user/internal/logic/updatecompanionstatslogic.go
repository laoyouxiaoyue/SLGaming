package logic

import (
	"context"
	"strconv"

	"SLGaming/back/services/user/internal/model"
	"SLGaming/back/services/user/internal/svc"
	"SLGaming/back/services/user/user"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type UpdateCompanionStatsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateCompanionStatsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateCompanionStatsLogic {
	return &UpdateCompanionStatsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateCompanionStatsLogic) UpdateCompanionStats(in *user.UpdateCompanionStatsRequest) (*user.UpdateCompanionStatsResponse, error) {
	if in.GetUserId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	if in.GetNewRating() < 0 || in.GetNewRating() > 5 {
		return nil, status.Error(codes.InvalidArgument, "new_rating must be between 0 and 5")
	}
	if in.GetDeltaOrders() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "delta_orders must be positive")
	}

	db := l.svcCtx.DB().WithContext(l.ctx)

	var p model.CompanionProfile
	if err := db.Where("user_id = ?", in.GetUserId()).First(&p).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Error(codes.NotFound, "companion profile not found")
		}
		l.Errorf("get companion profile failed: %v", err)
		return nil, status.Error(codes.Internal, "get companion profile failed")
	}

	oldOrders := p.TotalOrders
	p.TotalOrders += in.GetDeltaOrders()

	// 重新计算加权平均评分： (旧评分*旧单数 + 本次评分*增量) / 新总单数
	if p.TotalOrders > 0 {
		p.Rating = (p.Rating*float64(oldOrders) + in.GetNewRating()*float64(in.GetDeltaOrders())) / float64(p.TotalOrders)
	}

	if err := db.Save(&p).Error; err != nil {
		l.Errorf("update companion stats failed: %v", err)
		return nil, status.Error(codes.Internal, "update companion stats failed")
	}

	// 直接更新 Redis 排名 ZSet（如果配置了 Redis）
	// 失败只记录日志，不影响主流程
	if l.svcCtx.Redis != nil {
		userIDStr := strconv.FormatUint(p.UserID, 10)

		// 更新评分排名 ZSet（乘以 10000 转为整数，保持精度）
		ratingScore := int64(p.Rating * 10000)
		_, err := l.svcCtx.Redis.Zadd("ranking:rating", ratingScore, userIDStr)
		if err != nil {
			l.Errorf("update rating ranking failed: user_id=%d, rating=%f, err=%v", p.UserID, p.Rating, err)
		}

		// 更新接单数排名 ZSet
		_, err = l.svcCtx.Redis.Zadd("ranking:orders", p.TotalOrders, userIDStr)
		if err != nil {
			l.Errorf("update orders ranking failed: user_id=%d, total_orders=%d, err=%v", p.UserID, p.TotalOrders, err)
		}
	}

	return &user.UpdateCompanionStatsResponse{
		Profile: toCompanionInfo(&p),
	}, nil
}
