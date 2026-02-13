package logic

import (
	"context"

	"SLGaming/back/services/order/internal/model"
	"SLGaming/back/services/order/internal/svc"
	"SLGaming/back/services/order/order"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type GetOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOrderLogic {
	return &GetOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetOrderLogic) GetOrder(in *order.GetOrderRequest) (*order.GetOrderResponse, error) {
	db := l.svcCtx.DB.WithContext(l.ctx)

	var o model.Order
	var err error

	switch {
	case in.GetId() != 0:
		err = db.Where("id = ?", in.GetId()).First(&o).Error
	case in.GetOrderNo() != "":
		err = db.Where("order_no = ?", in.GetOrderNo()).First(&o).Error
	default:
		return nil, status.Error(codes.InvalidArgument, "id or order_no is required")
	}

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Error(codes.NotFound, "order not found")
		}
		l.Errorf("get order failed: %v", err)
		return nil, status.Error(codes.Internal, "get order failed")
	}

	// 检查订单是否对该用户已删除
	if in.GetOperatorId() != 0 {
		if in.GetOperatorId() == o.BossID && o.BossDeletedAt != nil {
			return nil, status.Error(codes.NotFound, "order not found")
		}
		if in.GetOperatorId() == o.CompanionID && o.CompanionDeletedAt != nil {
			return nil, status.Error(codes.NotFound, "order not found")
		}
	}

	return &order.GetOrderResponse{
		Order: toOrderInfo(&o),
	}, nil
}
