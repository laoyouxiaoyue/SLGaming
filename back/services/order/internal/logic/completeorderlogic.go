package logic

import (
	"context"
	"time"

	"SLGaming/back/services/order/internal/model"
	"SLGaming/back/services/order/internal/svc"
	"SLGaming/back/services/order/order"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type CompleteOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCompleteOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CompleteOrderLogic {
	return &CompleteOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CompleteOrderLogic) CompleteOrder(in *order.CompleteOrderRequest) (*order.CompleteOrderResponse, error) {
	if in.GetOrderId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "order_id is required")
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

	if o.Status != model.OrderStatusInService && o.Status != model.OrderStatusAccepted {
		return nil, status.Error(codes.FailedPrecondition, "order is not in progress")
	}

	now := time.Now()
	o.Status = model.OrderStatusCompleted
	o.CompletedAt = &now

	if err := db.Save(&o).Error; err != nil {
		l.Errorf("complete order failed: %v", err)
		return nil, status.Error(codes.Internal, "complete order failed")
	}

	return &order.CompleteOrderResponse{
		Order: toOrderInfo(&o),
	}, nil
}
