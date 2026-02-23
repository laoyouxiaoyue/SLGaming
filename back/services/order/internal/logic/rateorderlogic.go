package logic

import (
	"context"

	"SLGaming/back/services/order/internal/model"
	"SLGaming/back/services/order/internal/svc"
	"SLGaming/back/services/order/order"
	"SLGaming/back/services/user/userclient"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type RateOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRateOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RateOrderLogic {
	return &RateOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RateOrderLogic) RateOrder(in *order.RateOrderRequest) (*order.RateOrderResponse, error) {
	if in.GetOrderId() == 0 || in.GetBossId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "order_id and boss_id are required")
	}
	if in.GetRating() < 0 || in.GetRating() > 5 {
		return nil, status.Error(codes.InvalidArgument, "rating must be between 0 and 5")
	}

	db := l.svcCtx.DB.WithContext(l.ctx)

	var o model.Order
	if err := db.Where("id = ?", in.GetOrderId()).First(&o).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Error(codes.NotFound, "order not found")
		}
		l.Errorf("get order failed: %v", err)
		return nil, status.Error(codes.Internal, "get order failed")
	}

	// 只能由该订单的老板评价，且订单必须已完成
	if o.BossID != in.GetBossId() {
		return nil, status.Error(codes.PermissionDenied, "not allowed to rate this order")
	}
	if o.Status != model.OrderStatusCompleted {
		return nil, status.Error(codes.FailedPrecondition, "order is not completed")
	}
	if o.Rating > 0 {
		return nil, status.Error(codes.AlreadyExists, "order already rated")
	}

	o.Rating = in.GetRating()
	o.Comment = in.GetComment()
	o.Status = model.OrderStatusRated

	if err := db.Save(&o).Error; err != nil {
		l.Errorf("rate order failed: %v", err)
		return nil, status.Error(codes.Internal, "rate order failed")
	}

	// 同步更新陪玩的评分和接单统计
	if l.svcCtx.UserRPC != nil {
		_, err := l.svcCtx.UserRPC.UpdateCompanionStats(l.ctx, &userclient.UpdateCompanionStatsRequest{
			UserId:      o.CompanionID,
			DeltaOrders: 1,
			NewRating:   in.GetRating(),
		})
		if err != nil {
			if st, ok := status.FromError(err); ok {
				l.Errorf("update companion stats failed: code=%v, msg=%s", st.Code(), st.Message())
			} else {
				l.Errorf("update companion stats failed: %v", err)
			}
		}
	}

	return &order.RateOrderResponse{
		Order: toOrderInfo(&o),
	}, nil
}
