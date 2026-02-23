package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"SLGaming/back/pkg/lock"
	"SLGaming/back/services/order/internal/helper"
	"SLGaming/back/services/order/internal/metrics"
	"SLGaming/back/services/order/internal/model"
	orderMQ "SLGaming/back/services/order/internal/mq"
	"SLGaming/back/services/order/internal/svc"
	"SLGaming/back/services/order/internal/tx"
	"SLGaming/back/services/order/order"

	"github.com/apache/rocketmq-client-go/v2/primitive"
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
	start := time.Now()

	helper.LogRequest(l.Logger, helper.OpCancelOrder, map[string]interface{}{
		"order_id":    in.GetOrderId(),
		"operator_id": in.GetOperatorId(),
		"reason":      in.GetReason(),
	})

	if in.GetOrderId() == 0 {
		metrics.OrderCancelTotal.WithLabelValues("invalid_argument", "false").Inc()
		return nil, status.Error(codes.InvalidArgument, "order_id is required")
	}

	// 使用分布式锁防止并发取消订单
	// 锁的 key 基于 order_id，防止同一订单被并发取消
	lockKey := fmt.Sprintf("cancel_order:%d", in.GetOrderId())

	// 如果分布式锁未初始化，直接执行（降级处理）
	if l.svcCtx.DistributedLock == nil {
		helper.LogWarning(l.Logger, helper.OpCancelOrder, "distributed lock not initialized, skipping lock", nil)
		return l.doCancelOrder(in, start)
	}

	var result *order.CancelOrderResponse
	var cancelErr error

	lockOptions := &lock.LockOptions{
		TTL:           30 * time.Second,
		RetryInterval: 100 * time.Millisecond,
		MaxWaitTime:   5 * time.Second,
	}

	lockStart := time.Now()
	err := l.svcCtx.DistributedLock.WithLock(l.ctx, lockKey, lockOptions, func() error {
		result, cancelErr = l.doCancelOrder(in, start)
		return cancelErr
	})

	if err != nil {
		metrics.LockAcquireTotal.WithLabelValues("cancel_order", "failed").Inc()
		if err == context.DeadlineExceeded || err == context.Canceled {
			metrics.OrderCancelTotal.WithLabelValues("lock_timeout", "false").Inc()
			return nil, status.Error(codes.DeadlineExceeded, "acquire lock timeout, please try again later")
		}
		helper.LogError(l.Logger, helper.OpCancelOrder, "cancel order with lock failed", err, nil)
		metrics.OrderCancelTotal.WithLabelValues("lock_failed", "false").Inc()
		return nil, status.Error(codes.Internal, "cancel order failed")
	}

	metrics.LockAcquireTotal.WithLabelValues("cancel_order", "success").Inc()
	metrics.LockAcquireDuration.WithLabelValues("cancel_order").Observe(time.Since(lockStart).Seconds())

	return result, cancelErr
}

// doCancelOrder 执行实际的订单取消逻辑（不加锁）
func (l *CancelOrderLogic) doCancelOrder(in *order.CancelOrderRequest, start time.Time) (*order.CancelOrderResponse, error) {
	db := l.svcCtx.DB.WithContext(l.ctx)

	var o model.Order
	if err := db.Where("id = ?", in.GetOrderId()).First(&o).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			metrics.OrderCancelTotal.WithLabelValues("not_found", "false").Inc()
			metrics.OrderCancelDuration.WithLabelValues().Observe(time.Since(start).Seconds())
			return nil, status.Error(codes.NotFound, "order not found")
		}
		helper.LogError(l.Logger, helper.OpCancelOrder, "get order failed", err, map[string]interface{}{
			"order_id": in.GetOrderId(),
		})
		metrics.OrderCancelTotal.WithLabelValues("query_failed", "false").Inc()
		metrics.OrderCancelDuration.WithLabelValues().Observe(time.Since(start).Seconds())
		return nil, status.Error(codes.Internal, "get order failed")
	}

	// 已完成或已取消的订单不能再次取消
	if o.Status == model.OrderStatusCompleted || o.Status == model.OrderStatusCancelled {
		helper.LogWarning(l.Logger, helper.OpCancelOrder, "order already finished or cancelled", map[string]interface{}{
			"order_id": o.ID,
			"status":   o.Status,
		})
		metrics.OrderCancelTotal.WithLabelValues("already_finished", "false").Inc()
		metrics.OrderCancelDuration.WithLabelValues().Observe(time.Since(start).Seconds())
		return nil, status.Error(codes.FailedPrecondition, "order is already finished or cancelled")
	}

	// 正在取消中的订单不能重复取消
	if o.Status == model.OrderStatusCancelRefunding {
		helper.LogWarning(l.Logger, helper.OpCancelOrder, "order is already in cancelling", map[string]interface{}{
			"order_id": o.ID,
		})
		metrics.OrderCancelTotal.WithLabelValues("already_cancelling", "false").Inc()
		metrics.OrderCancelDuration.WithLabelValues().Observe(time.Since(start).Seconds())
		return nil, status.Error(codes.FailedPrecondition, "order is already in cancelling process")
	}

	// 服务中的订单不能取消（需要先完成或走仲裁流程）
	if o.Status == model.OrderStatusInService {
		helper.LogWarning(l.Logger, helper.OpCancelOrder, "order is in service, cannot cancel", map[string]interface{}{
			"order_id": o.ID,
		})
		metrics.OrderCancelTotal.WithLabelValues("in_service", "false").Inc()
		metrics.OrderCancelDuration.WithLabelValues().Observe(time.Since(start).Seconds())
		return nil, status.Error(codes.FailedPrecondition, "order is in service, cannot cancel")
	}

	// 权限验证：根据订单状态决定谁可以取消
	operatorID := in.GetOperatorId()
	if o.Status == model.OrderStatusAccepted {
		// 已接单状态：只有陪玩可以取消订单
		if operatorID != o.CompanionID {
			helper.LogError(l.Logger, helper.OpCancelOrder, "permission denied: only companion can cancel accepted order", nil, map[string]interface{}{
				"order_id":     o.ID,
				"operator_id":  operatorID,
				"companion_id": o.CompanionID,
			})
			metrics.OrderCancelTotal.WithLabelValues("permission_denied", "false").Inc()
			metrics.OrderCancelDuration.WithLabelValues().Observe(time.Since(start).Seconds())
			return nil, status.Error(codes.PermissionDenied, "only the companion can cancel an accepted order")
		}
	} else if o.Status == model.OrderStatusCreated || o.Status == model.OrderStatusPaid {
		// 已创建或已支付但未接单状态：只有老板可以取消订单
		if operatorID != o.BossID {
			helper.LogError(l.Logger, helper.OpCancelOrder, "permission denied: only boss can cancel unpaid/unaccepted order", nil, map[string]interface{}{
				"order_id":    o.ID,
				"operator_id": operatorID,
				"boss_id":     o.BossID,
			})
			metrics.OrderCancelTotal.WithLabelValues("permission_denied", "false").Inc()
			metrics.OrderCancelDuration.WithLabelValues().Observe(time.Since(start).Seconds())
			return nil, status.Error(codes.PermissionDenied, "only the boss can cancel an unpaid or unaccepted order")
		}
	} else {
		// 其他状态：理论上不应该到达这里，但为了安全起见，检查是否为订单相关方
		if operatorID != o.BossID && operatorID != o.CompanionID {
			helper.LogError(l.Logger, helper.OpCancelOrder, "permission denied: not order participant", nil, map[string]interface{}{
				"order_id":     o.ID,
				"operator_id":  operatorID,
				"boss_id":      o.BossID,
				"companion_id": o.CompanionID,
			})
			metrics.OrderCancelTotal.WithLabelValues("permission_denied", "false").Inc()
			metrics.OrderCancelDuration.WithLabelValues().Observe(time.Since(start).Seconds())
			return nil, status.Error(codes.PermissionDenied, "only the boss or companion of this order can cancel it")
		}
	}

	// 根据订单状态决定是否需要退款
	needRefund := o.Status == model.OrderStatusPaid || o.Status == model.OrderStatusAccepted

	helper.LogInfo(l.Logger, helper.OpCancelOrder, "cancelling order", map[string]interface{}{
		"order_id":       o.ID,
		"order_no":       o.OrderNo,
		"current_status": o.Status,
		"operator_id":    in.GetOperatorId(),
		"need_refund":    needRefund,
	})

	// 使用 RocketMQ 事务消息发送 ORDER_CANCELLED，并在本地事务中更新订单状态
	if l.svcCtx.OrderEventTxProducer == nil {
		metrics.OrderCancelTotal.WithLabelValues("producer_not_initialized", "false").Inc()
		metrics.OrderCancelDuration.WithLabelValues().Observe(time.Since(start).Seconds())
		return nil, status.Error(codes.FailedPrecondition, "order transaction producer not initialized")
	}

	payload := &tx.OrderCancelledPayload{
		OrderID:      o.ID,
		OrderNo:      o.OrderNo,
		BossID:       o.BossID,
		CompanionID:  o.CompanionID,
		Amount:       o.TotalAmount,
		BizOrderID:   o.OrderNo,
		NeedRefund:   needRefund,
		CancelReason: in.GetReason(),
	}

	// 构造事务消息
	msgBody, err := json.Marshal(payload)
	if err != nil {
		helper.LogError(l.Logger, helper.OpCancelOrder, "marshal cancelled payload failed", err, nil)
		metrics.OrderCancelTotal.WithLabelValues("marshal_failed", "false").Inc()
		metrics.OrderCancelDuration.WithLabelValues().Observe(time.Since(start).Seconds())
		return nil, status.Error(codes.Internal, "marshal cancelled event failed")
	}
	msg := primitive.NewMessage(orderMQ.OrderEventTopic(), msgBody)
	msg.WithTag(orderMQ.EventTypeCancelled())

	txRes, err := l.svcCtx.OrderEventTxProducer.SendMessageInTransaction(l.ctx, msg)
	if err != nil {
		helper.LogError(l.Logger, helper.OpCancelOrder, "send transactional message failed", err, map[string]interface{}{
			"result": fmt.Sprintf("%+v", txRes),
		})
		metrics.OrderCancelTotal.WithLabelValues("tx_message_failed", "false").Inc()
		metrics.OrderCancelDuration.WithLabelValues().Observe(time.Since(start).Seconds())
		return nil, status.Error(codes.Internal, "cancel order failed")
	}

	// 此时本地事务（ExecuteOrderTx -> ExecuteCancelOrderTx）已经执行完成，
	// 但是否成功需要通过查询订单确认
	var updatedOrder model.Order
	if err := db.Where("order_no = ?", o.OrderNo).First(&updatedOrder).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			helper.LogError(l.Logger, helper.OpCancelOrder, "cancel order transaction rolled back", err, map[string]interface{}{
				"order_no": o.OrderNo,
			})
			metrics.OrderCancelTotal.WithLabelValues("tx_rolled_back", "false").Inc()
			metrics.OrderCancelDuration.WithLabelValues().Observe(time.Since(start).Seconds())
			return nil, status.Error(codes.Internal, "cancel order transaction rolled back")
		}
		helper.LogError(l.Logger, helper.OpCancelOrder, "query order after transactional message failed", err, nil)
		metrics.OrderCancelTotal.WithLabelValues("query_failed", "false").Inc()
		metrics.OrderCancelDuration.WithLabelValues().Observe(time.Since(start).Seconds())
		return nil, status.Error(codes.Internal, "cancel order failed")
	}

	// 记录成功指标
	needRefundStr := "false"
	if needRefund {
		needRefundStr = "true"
	}
	metrics.OrderCancelTotal.WithLabelValues("success", needRefundStr).Inc()
	metrics.OrderCancelDuration.WithLabelValues().Observe(time.Since(start).Seconds())
	metrics.OrderStatusTransition.WithLabelValues(fmt.Sprintf("%d", o.Status), fmt.Sprintf("%d", updatedOrder.Status)).Inc()

	helper.LogSuccess(l.Logger, helper.OpCancelOrder, map[string]interface{}{
		"order_id":    updatedOrder.ID,
		"order_no":    updatedOrder.OrderNo,
		"status":      updatedOrder.Status,
		"need_refund": needRefund,
	})

	return &order.CancelOrderResponse{
		Order: toOrderInfo(&updatedOrder),
	}, nil
}
