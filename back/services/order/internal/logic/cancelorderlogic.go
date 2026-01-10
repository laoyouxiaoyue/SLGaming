package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"SLGaming/back/pkg/lock"
	"SLGaming/back/services/order/internal/model"
	"SLGaming/back/services/order/internal/svc"
	"SLGaming/back/services/order/order"

	"github.com/google/uuid"
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

	// 使用分布式锁防止并发取消订单
	// 锁的 key 基于 order_id，防止同一订单被并发取消
	lockKey := fmt.Sprintf("cancel_order:%d", in.GetOrderId())
	lockValue := uuid.New().String()

	// 如果分布式锁未初始化，直接执行（降级处理）
	if l.svcCtx.DistributedLock == nil {
		l.Infof("distributed lock not initialized, skipping lock for order cancellation")
		return l.doCancelOrder(in)
	}

	// 使用分布式锁执行订单取消
	var result *order.CancelOrderResponse
	var cancelErr error

	lockOptions := &lock.LockOptions{
		TTL:           30,                     // 锁过期时间 30 秒
		RetryInterval: 100 * time.Millisecond, // 重试间隔 100ms
		MaxWaitTime:   5 * time.Second,        // 最大等待时间 5 秒
	}

	err := l.svcCtx.DistributedLock.WithLock(l.ctx, lockKey, lockValue, lockOptions, func() error {
		result, cancelErr = l.doCancelOrder(in)
		return cancelErr
	})

	if err != nil {
		if err == context.DeadlineExceeded || err == context.Canceled {
			return nil, status.Error(codes.DeadlineExceeded, "acquire lock timeout, please try again later")
		}
		l.Errorf("cancel order with lock failed: %v", err)
		return nil, status.Error(codes.Internal, "cancel order failed")
	}

	return result, cancelErr
}

// doCancelOrder 执行实际的订单取消逻辑（不加锁）
func (l *CancelOrderLogic) doCancelOrder(in *order.CancelOrderRequest) (*order.CancelOrderResponse, error) {
	db := l.svcCtx.DB.WithContext(l.ctx)

	var o model.Order
	if err := db.Where("id = ?", in.GetOrderId()).First(&o).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Error(codes.NotFound, "order not found")
		}
		l.Errorf("cancel order failed: get order failed, order_id=%d, error=%v", in.GetOrderId(), err)
		return nil, status.Error(codes.Internal, "get order failed")
	}

	// 已完成或已取消的订单不能再次取消
	if o.Status == model.OrderStatusCompleted || o.Status == model.OrderStatusCancelled {
		l.Infof("cancel order failed: order already finished or cancelled, order_id=%d, status=%d",
			o.ID, o.Status)
		return nil, status.Error(codes.FailedPrecondition, "order is already finished or cancelled")
	}

	// 正在取消中的订单不能重复取消
	if o.Status == model.OrderStatusCancelRefunding {
		l.Infof("cancel order failed: order is already in cancelling, order_id=%d", o.ID)
		return nil, status.Error(codes.FailedPrecondition, "order is already in cancelling process")
	}

	// 服务中的订单不能取消（需要先完成或走仲裁流程）
	if o.Status == model.OrderStatusInService {
		l.Infof("cancel order failed: order is in service, order_id=%d", o.ID)
		return nil, status.Error(codes.FailedPrecondition, "order is in service, cannot cancel")
	}

	// 权限验证：根据订单状态决定谁可以取消
	operatorID := in.GetOperatorId()
	if o.Status == model.OrderStatusAccepted {
		// 已接单状态：只有陪玩可以取消订单
		if operatorID != o.CompanionID {
			l.Errorf("cancel order failed: permission denied, only companion can cancel accepted order, order_id=%d, operator_id=%d, companion_id=%d",
				o.ID, operatorID, o.CompanionID)
			return nil, status.Error(codes.PermissionDenied, "only the companion can cancel an accepted order")
		}
	} else if o.Status == model.OrderStatusCreated || o.Status == model.OrderStatusPaid {
		// 已创建或已支付但未接单状态：只有老板可以取消订单
		if operatorID != o.BossID {
			l.Errorf("cancel order failed: permission denied, only boss can cancel unpaid/unaccepted order, order_id=%d, operator_id=%d, boss_id=%d",
				o.ID, operatorID, o.BossID)
			return nil, status.Error(codes.PermissionDenied, "only the boss can cancel an unpaid or unaccepted order")
		}
	} else {
		// 其他状态：理论上不应该到达这里，但为了安全起见，检查是否为订单相关方
		if operatorID != o.BossID && operatorID != o.CompanionID {
			l.Errorf("cancel order failed: permission denied, order_id=%d, operator_id=%d, boss_id=%d, companion_id=%d",
				o.ID, operatorID, o.BossID, o.CompanionID)
			return nil, status.Error(codes.PermissionDenied, "only the boss or companion of this order can cancel it")
		}
	}

	now := time.Now()

	// 根据订单状态决定是否需要退款
	needRefund := o.Status == model.OrderStatusPaid || o.Status == model.OrderStatusAccepted

	l.Infof("cancelling order: order_id=%d, order_no=%s, current_status=%d, operator_id=%d, need_refund=%v",
		o.ID, o.OrderNo, o.Status, in.GetOperatorId(), needRefund)

	// 在一个事务中处理取消逻辑
	if err := db.Transaction(func(tx *gorm.DB) error {
		if needRefund {
			// 已支付或已接单的订单：需要退款
			// 更新订单状态为取消中（等待退款）
			o.Status = model.OrderStatusCancelRefunding
			o.CancelledAt = &now
			o.CancelReason = in.GetReason()

			if err := tx.Save(&o).Error; err != nil {
				l.Errorf("cancel order failed: update order status to cancel_refunding failed, order_id=%d, error=%v",
					o.ID, err)
				return status.Error(codes.Internal, "cancel order failed")
			}

			// 写一条 ORDER_CANCELLED 事件到 outbox，由后台任务异步发送到 MQ，触发退款
			evt := &model.OrderEventOutbox{
				EventType: "ORDER_CANCELLED",
				Payload:   buildOrderCancelledPayload(&o),
				Status:    "PENDING",
			}

			if err := tx.Create(evt).Error; err != nil {
				l.Errorf("cancel order failed: create order event outbox failed, order_id=%d, error=%v",
					o.ID, err)
				return status.Error(codes.Internal, "cancel order failed")
			}

			l.Infof("order cancelled with refund: order_id=%d, order_no=%s, amount=%d",
				o.ID, o.OrderNo, o.TotalAmount)
		} else {
			// 已创建但未支付的订单：直接取消，不需要退款
			// 更新订单状态为已取消
			o.Status = model.OrderStatusCancelled
			o.CancelledAt = &now
			o.CancelReason = in.GetReason()

			if err := tx.Save(&o).Error; err != nil {
				l.Errorf("cancel order failed: update order status to cancelled failed, order_id=%d, error=%v",
					o.ID, err)
				return status.Error(codes.Internal, "cancel order failed")
			}

			l.Infof("order cancelled without refund: order_id=%d, order_no=%s (order was not paid)",
				o.ID, o.OrderNo)
		}

		return nil
	}); err != nil {
		// 如果是我们在事务中返回的 gRPC status error，直接透传
		if s, ok := status.FromError(err); ok {
			return nil, s.Err()
		}
		l.Errorf("cancel order transaction failed: order_id=%d, error=%v", o.ID, err)
		return nil, status.Error(codes.Internal, "cancel order failed")
	}

	return &order.CancelOrderResponse{
		Order: toOrderInfo(&o),
	}, nil
}
