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
	"SLGaming/back/services/user/userclient"

	"github.com/apache/rocketmq-client-go/v2/primitive"
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
	start := time.Now()

	helper.LogRequest(l.Logger, helper.OpCompleteOrder, map[string]interface{}{
		"order_id":    in.GetOrderId(),
		"operator_id": in.GetOperatorId(),
	})

	if in.GetOrderId() == 0 {
		metrics.OrderCompleteTotal.WithLabelValues("invalid_argument").Inc()
		return nil, status.Error(codes.InvalidArgument, "order_id is required")
	}
	if in.GetOperatorId() == 0 {
		metrics.OrderCompleteTotal.WithLabelValues("invalid_argument").Inc()
		return nil, status.Error(codes.InvalidArgument, "operator_id is required")
	}

	// 使用分布式锁防止并发完成订单
	// 锁的 key 基于 order_id，防止同一订单被并发完成
	lockKey := fmt.Sprintf("complete_order:%d", in.GetOrderId())

	// 如果分布式锁未初始化，直接执行（降级处理）
	if l.svcCtx.DistributedLock == nil {
		helper.LogWarning(l.Logger, helper.OpCompleteOrder, "distributed lock not initialized, skipping lock", nil)
		return l.doCompleteOrder(in, start)
	}

	var result *order.CompleteOrderResponse
	var completeErr error

	lockOptions := &lock.LockOptions{
		TTL:           30 * time.Second,
		RetryInterval: 100 * time.Millisecond,
		MaxWaitTime:   5 * time.Second,
	}

	lockStart := time.Now()
	err := l.svcCtx.DistributedLock.WithLock(l.ctx, lockKey, lockOptions, func() error {
		result, completeErr = l.doCompleteOrder(in, start)
		return completeErr
	})

	if err != nil {
		metrics.LockAcquireTotal.WithLabelValues("complete_order", "failed").Inc()
		if err == context.DeadlineExceeded || err == context.Canceled {
			metrics.OrderCompleteTotal.WithLabelValues("lock_timeout").Inc()
			return nil, status.Error(codes.DeadlineExceeded, "acquire lock timeout, please try again later")
		}
		helper.LogError(l.Logger, helper.OpCompleteOrder, "complete order with lock failed", err, nil)
		metrics.OrderCompleteTotal.WithLabelValues("lock_failed").Inc()
		return nil, status.Error(codes.Internal, "complete order failed")
	}

	metrics.LockAcquireTotal.WithLabelValues("complete_order", "success").Inc()
	metrics.LockAcquireDuration.WithLabelValues("complete_order").Observe(time.Since(lockStart).Seconds())

	return result, completeErr
}

// doCompleteOrder 执行实际的订单完成逻辑（不加锁）
func (l *CompleteOrderLogic) doCompleteOrder(in *order.CompleteOrderRequest, start time.Time) (*order.CompleteOrderResponse, error) {
	db := l.svcCtx.DB.WithContext(l.ctx)

	var o model.Order
	if err := db.Where("id = ?", in.GetOrderId()).First(&o).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			metrics.OrderCompleteTotal.WithLabelValues("not_found").Inc()
			metrics.OrderCompleteDuration.WithLabelValues().Observe(time.Since(start).Seconds())
			return nil, status.Error(codes.NotFound, "order not found")
		}
		helper.LogError(l.Logger, helper.OpCompleteOrder, "get order failed", err, map[string]interface{}{
			"order_id": in.GetOrderId(),
		})
		metrics.OrderCompleteTotal.WithLabelValues("query_failed").Inc()
		metrics.OrderCompleteDuration.WithLabelValues().Observe(time.Since(start).Seconds())
		return nil, status.Error(codes.Internal, "get order failed")
	}

	if o.BossID != in.GetOperatorId() {
		if l.svcCtx.UserRPC == nil {
			metrics.OrderCompleteTotal.WithLabelValues("not_authorized").Inc()
			metrics.OrderCompleteDuration.WithLabelValues().Observe(time.Since(start).Seconds())
			return nil, status.Error(codes.PermissionDenied, "only boss or admin can complete order")
		}
		userResp, err := l.svcCtx.UserRPC.GetUser(l.ctx, &userclient.GetUserRequest{Id: in.GetOperatorId()})
		if err != nil {
			helper.LogError(l.Logger, helper.OpCompleteOrder, "get operator role failed", err, map[string]interface{}{
				"operator_id": in.GetOperatorId(),
			})
			metrics.OrderCompleteTotal.WithLabelValues("role_query_failed").Inc()
			metrics.OrderCompleteDuration.WithLabelValues().Observe(time.Since(start).Seconds())
			return nil, status.Error(codes.Internal, "get operator role failed")
		}
		if userResp.GetUser() == nil || userResp.GetUser().GetRole() != 3 {
			helper.LogWarning(l.Logger, helper.OpCompleteOrder, "permission denied: not boss or admin", map[string]interface{}{
				"order_id":    in.GetOrderId(),
				"operator_id": in.GetOperatorId(),
				"boss_id":     o.BossID,
			})
			metrics.OrderCompleteTotal.WithLabelValues("not_authorized").Inc()
			metrics.OrderCompleteDuration.WithLabelValues().Observe(time.Since(start).Seconds())
			return nil, status.Error(codes.PermissionDenied, "only boss or admin can complete order")
		}
	}

	if o.Status != model.OrderStatusInService && o.Status != model.OrderStatusAccepted {
		helper.LogWarning(l.Logger, helper.OpCompleteOrder, "order is not in progress", map[string]interface{}{
			"order_id": in.GetOrderId(),
			"status":   o.Status,
		})
		metrics.OrderCompleteTotal.WithLabelValues("invalid_status").Inc()
		metrics.OrderCompleteDuration.WithLabelValues().Observe(time.Since(start).Seconds())
		return nil, status.Error(codes.FailedPrecondition, "order is not in progress")
	}

	if l.svcCtx.OrderEventTxProducer == nil {
		metrics.OrderCompleteTotal.WithLabelValues("producer_not_initialized").Inc()
		metrics.OrderCompleteDuration.WithLabelValues().Observe(time.Since(start).Seconds())
		return nil, status.Error(codes.FailedPrecondition, "order transaction producer not initialized")
	}

	payload := &tx.OrderCompletedPayload{
		OrderID:     o.ID,
		OrderNo:     o.OrderNo,
		BossID:      o.BossID,
		CompanionID: o.CompanionID,
		Amount:      o.TotalAmount,
		BizOrderID:  o.OrderNo,
	}

	msgBody, err := json.Marshal(payload)
	if err != nil {
		helper.LogError(l.Logger, helper.OpCompleteOrder, "marshal completed payload failed", err, nil)
		metrics.OrderCompleteTotal.WithLabelValues("marshal_failed").Inc()
		metrics.OrderCompleteDuration.WithLabelValues().Observe(time.Since(start).Seconds())
		return nil, status.Error(codes.Internal, "marshal completed event failed")
	}
	msg := primitive.NewMessage(orderMQ.OrderEventTopic(), msgBody)
	msg.WithTag(orderMQ.EventTypeCompleted())

	txRes, err := l.svcCtx.OrderEventTxProducer.SendMessageInTransaction(l.ctx, msg)
	if err != nil {
		helper.LogError(l.Logger, helper.OpCompleteOrder, "send transactional message failed", err, map[string]interface{}{
			"result": fmt.Sprintf("%+v", txRes),
		})
		metrics.OrderCompleteTotal.WithLabelValues("tx_message_failed").Inc()
		metrics.OrderCompleteDuration.WithLabelValues().Observe(time.Since(start).Seconds())
		return nil, status.Error(codes.Internal, "complete order failed")
	}

	var updatedOrder model.Order
	if err := db.Where("order_no = ?", o.OrderNo).First(&updatedOrder).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			helper.LogError(l.Logger, helper.OpCompleteOrder, "complete order transaction rolled back", err, map[string]interface{}{
				"order_no": o.OrderNo,
			})
			metrics.OrderCompleteTotal.WithLabelValues("tx_rolled_back").Inc()
			metrics.OrderCompleteDuration.WithLabelValues().Observe(time.Since(start).Seconds())
			return nil, status.Error(codes.Internal, "complete order transaction rolled back")
		}
		helper.LogError(l.Logger, helper.OpCompleteOrder, "query order after transactional message failed", err, nil)
		metrics.OrderCompleteTotal.WithLabelValues("query_failed").Inc()
		metrics.OrderCompleteDuration.WithLabelValues().Observe(time.Since(start).Seconds())
		return nil, status.Error(codes.Internal, "complete order failed")
	}

	metrics.OrderCompleteTotal.WithLabelValues("success").Inc()
	metrics.OrderCompleteDuration.WithLabelValues().Observe(time.Since(start).Seconds())
	metrics.OrderStatusTransition.WithLabelValues(fmt.Sprintf("%d", o.Status), fmt.Sprintf("%d", updatedOrder.Status)).Inc()

	helper.LogSuccess(l.Logger, helper.OpCompleteOrder, map[string]interface{}{
		"order_id":     updatedOrder.ID,
		"order_no":     updatedOrder.OrderNo,
		"companion_id": updatedOrder.CompanionID,
		"boss_id":      updatedOrder.BossID,
		"amount":       updatedOrder.TotalAmount,
	})

	return &order.CompleteOrderResponse{
		Order: toOrderInfo(&updatedOrder),
	}, nil
}
