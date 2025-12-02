package logic

import (
	"context"

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

	return &user.UpdateCompanionStatsResponse{
		Profile: toCompanionInfo(&p),
	}, nil
}
