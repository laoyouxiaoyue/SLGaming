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

type DeleteOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteOrderLogic {
	return &DeleteOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteOrderLogic) DeleteOrder(in *order.DeleteOrderRequest) (*order.DeleteOrderResponse, error) {
	if in.GetOrderId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "order_id is required")
	}

	if in.GetOperatorId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "operator_id is required")
	}

	db := l.svcCtx.DB.WithContext(l.ctx)

	// 查询订单
	var o model.Order
	if err := db.Where("id = ?", in.GetOrderId()).First(&o).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Error(codes.NotFound, "order not found")
		}
		l.Errorf("delete order failed: get order failed, order_id=%d, error=%v", in.GetOrderId(), err)
		return nil, status.Error(codes.Internal, "get order failed")
	}

	// 校验订单状态：只允许删除已完成、已取消、已评价的订单
	if !isOrderDeletable(o.Status) {
		l.Infof("delete order failed: order status not deletable, order_id=%d, status=%d",
			o.ID, o.Status)
		return nil, status.Error(codes.FailedPrecondition, "only completed, cancelled or rated orders can be deleted")
	}

	// 权限校验：只有老板或陪玩可以删除自己的订单
	operatorID := in.GetOperatorId()
	now := time.Now()

	if operatorID == o.BossID {
		// 老板删除：设置 BossDeletedAt
		if err := db.Model(&o).Update("boss_deleted_at", now).Error; err != nil {
			l.Errorf("delete order failed: update boss_deleted_at failed, order_id=%d, error=%v", o.ID, err)
			return nil, status.Error(codes.Internal, "delete order failed")
		}
		l.Infof("order deleted by boss: order_id=%d, boss_id=%d", o.ID, operatorID)
	} else if operatorID == o.CompanionID {
		// 陪玩删除：设置 CompanionDeletedAt
		if err := db.Model(&o).Update("companion_deleted_at", now).Error; err != nil {
			l.Errorf("delete order failed: update companion_deleted_at failed, order_id=%d, error=%v", o.ID, err)
			return nil, status.Error(codes.Internal, "delete order failed")
		}
		l.Infof("order deleted by companion: order_id=%d, companion_id=%d", o.ID, operatorID)
	} else {
		l.Errorf("delete order failed: permission denied, order_id=%d, operator_id=%d, boss_id=%d, companion_id=%d",
			o.ID, operatorID, o.BossID, o.CompanionID)
		return nil, status.Error(codes.PermissionDenied, "only the boss or companion of this order can delete it")
	}

	return &order.DeleteOrderResponse{
		Success: true,
	}, nil
}

// isOrderDeletable 判断订单状态是否可删除
func isOrderDeletable(status int32) bool {
	return status == model.OrderStatusCompleted ||
		status == model.OrderStatusCancelled ||
		status == model.OrderStatusRated
}
