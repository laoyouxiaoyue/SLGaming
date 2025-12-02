package logic

import (
	"context"
	"time"

	"SLGaming/back/services/order/internal/model"
	"SLGaming/back/services/order/internal/svc"
	"SLGaming/back/services/order/order"
	"SLGaming/back/services/user/userclient"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type CancelOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCancelOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CancelOrderLogic {
	return &CancelOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CancelOrderLogic) CancelOrder(in *order.CancelOrderRequest) (*order.CancelOrderResponse, error) {
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

	// 已完成或已取消的订单不能再次取消
	if o.Status == model.OrderStatusCompleted || o.Status == model.OrderStatusCancelled {
		return nil, status.Error(codes.FailedPrecondition, "order is already finished or cancelled")
	}

	now := time.Now()
	o.Status = model.OrderStatusCancelled
	o.CancelledAt = &now
	o.CancelReason = in.GetReason()

	if err := db.Save(&o).Error; err != nil {
		l.Errorf("cancel order failed: %v", err)
		return nil, status.Error(codes.Internal, "cancel order failed")
	}

	// 根据订单状态和阶段，调用 user 钱包接口做退款逻辑（简单版本：已支付但未完成的订单全额退款）
	if l.svcCtx.UserRPC != nil && (o.Status == model.OrderStatusPaid || o.Status == model.OrderStatusAccepted) && o.TotalAmount > 0 {
		_, err := l.svcCtx.UserRPC.Recharge(l.ctx, &userclient.RechargeRequest{
			UserId:     o.BossID,
			Amount:     o.TotalAmount,
			BizOrderId: o.OrderNo,
			Remark:     "order refund",
		})
		if err != nil {
			if st, ok := status.FromError(err); ok {
				l.Errorf("refund wallet failed: code=%v, msg=%s", st.Code(), st.Message())
				return nil, err
			}
			l.Errorf("refund wallet failed: %v", err)
			return nil, status.Error(codes.Internal, "refund wallet failed")
		}
	}

	return &order.CancelOrderResponse{
		Order: toOrderInfo(&o),
	}, nil
}
