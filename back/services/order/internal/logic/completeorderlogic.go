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

// orderCompletedEventPayload 订单完成事件负载结构
type orderCompletedEventPayload struct {
	OrderID     uint64 `json:"order_id"`
	OrderNo     string `json:"order_no"`
	BossID      uint64 `json:"boss_id"`
	CompanionID uint64 `json:"companion_id"`
	Amount      int64  `json:"amount"`
	// 这里使用订单号作为业务幂等键，钱包侧可据此做幂等控制
	BizOrderID string `json:"biz_order_id"`
}

// buildOrderCompletedPayload 构造 ORDER_COMPLETED 事件的 JSON 负载
func buildOrderCompletedPayload(o *model.Order) string {
	payload := orderCompletedEventPayload{
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

	// 在一个事务中：更新订单状态为 COMPLETED + 写一条 ORDER_COMPLETED outbox 事件
	if err := db.Transaction(func(tx *gorm.DB) error {
		// 更新订单状态为已完成
		o.Status = model.OrderStatusCompleted
		o.CompletedAt = &now

		if err := tx.Save(&o).Error; err != nil {
			l.Errorf("update order status to completed failed: %v", err)
			return status.Error(codes.Internal, "complete order failed")
		}

		// 写一条 ORDER_COMPLETED 事件到 outbox，由后台任务异步发送到 MQ
		// 用户服务会消费这个事件，给陪玩充值
		evt := &model.OrderEventOutbox{
			EventType: "ORDER_COMPLETED",
			Payload:   buildOrderCompletedPayload(&o),
			Status:    "PENDING",
		}

		if err := tx.Create(evt).Error; err != nil {
			l.Errorf("create order event outbox failed: %v", err)
			return status.Error(codes.Internal, "complete order failed")
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return &order.CompleteOrderResponse{
		Order: toOrderInfo(&o),
	}, nil
}
