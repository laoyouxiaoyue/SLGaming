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

type StartOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewStartOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *StartOrderLogic {
	return &StartOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *StartOrderLogic) StartOrder(in *order.StartOrderRequest) (*order.StartOrderResponse, error) {
	if in.GetOrderId() == 0 || in.GetCompanionId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "order_id and companion_id are required")
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

	if o.CompanionID != in.GetCompanionId() {
		return nil, status.Error(codes.PermissionDenied, "not allowed to start this order")
	}
	if o.Status != model.OrderStatusAccepted {
		return nil, status.Error(codes.FailedPrecondition, "order is not in accepted state")
	}

	now := time.Now()
	o.Status = model.OrderStatusInService
	o.StartAt = &now

	if err := db.Save(&o).Error; err != nil {
		l.Errorf("start order failed: %v", err)
		return nil, status.Error(codes.Internal, "start order failed")
	}

	return &order.StartOrderResponse{
		Order: toOrderInfo(&o),
	}, nil
}
