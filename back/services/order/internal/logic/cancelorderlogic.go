package logic

import (
	"context"
	"encoding/json"
	"time"

	"SLGaming/back/services/order/internal/model"
	"SLGaming/back/services/order/internal/svc"
	"SLGaming/back/services/order/order"

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

// orderCancelledEventPayload 订单取消事件负载结构
type orderCancelledEventPayload struct {
	OrderID     uint64 `json:"order_id"`
	OrderNo     string `json:"order_no"`
	BossID      uint64 `json:"boss_id"`
	CompanionID uint64 `json:"companion_id"`
	Amount      int64  `json:"amount"`
	// 这里使用订单号作为业务幂等键，钱包侧可据此做幂等控制
	BizOrderID string `json:"biz_order_id"`
}

// buildOrderCancelledPayload 构造 ORDER_CANCELLED 事件的 JSON 负载
func buildOrderCancelledPayload(o *model.Order) string {
	payload := orderCancelledEventPayload{
		OrderID:     o.ID,
		OrderNo:     o.OrderNo,
		BossID:      o.BossID,
		CompanionID: o.CompanionID,
		Amount:      o.TotalAmount,
		BizOrderID:  o.OrderNo,
	}
	b, err := json.Marshal(payload)
	if err != nil {
		// 理论上不应该失败，失败时返回空字符串，上层会记录错误
		return ""
	}
	return string(b)
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

	// 仅允许从已支付/已接单进入“取消并退款”流程
	if o.Status != model.OrderStatusPaid && o.Status != model.OrderStatusAccepted {
		return nil, status.Error(codes.FailedPrecondition, "order status not allowed to cancel with refund")
	}

	now := time.Now()

	// 在一个事务中：更新订单状态为 CANCEL_REFUNDING + 写一条 ORDER_CANCELLED outbox 事件
	if err := db.Transaction(func(tx *gorm.DB) error {
		// 更新订单状态为取消中（等待退款）
		o.Status = model.OrderStatusCancelRefunding
		o.CancelledAt = &now
		o.CancelReason = in.GetReason()

		if err := tx.Save(&o).Error; err != nil {
			l.Errorf("update order status to cancel_refunding failed: %v", err)
			return status.Error(codes.Internal, "cancel order failed")
		}

		// 写一条 ORDER_CANCELLED 事件到 outbox，由后台任务异步发送到 MQ
		evt := &model.OrderEventOutbox{
			EventType: "ORDER_CANCELLED",
			Payload:   buildOrderCancelledPayload(&o),
			Status:    "PENDING",
		}

		if err := tx.Create(evt).Error; err != nil {
			l.Errorf("create order event outbox failed: %v", err)
			return status.Error(codes.Internal, "cancel order failed")
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return &order.CancelOrderResponse{
		Order: toOrderInfo(&o),
	}, nil
}
