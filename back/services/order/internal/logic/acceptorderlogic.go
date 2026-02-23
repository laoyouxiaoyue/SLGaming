package logic

import (
	"context"
	"fmt"
	"time"

	"SLGaming/back/pkg/lock"
	"SLGaming/back/services/order/internal/helper"
	"SLGaming/back/services/order/internal/metrics"
	"SLGaming/back/services/order/internal/model"
	"SLGaming/back/services/order/internal/svc"
	"SLGaming/back/services/order/order"
	"SLGaming/back/services/user/userclient"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type AcceptOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAcceptOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AcceptOrderLogic {
	return &AcceptOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *AcceptOrderLogic) AcceptOrder(in *order.AcceptOrderRequest) (*order.AcceptOrderResponse, error) {
	start := time.Now()

	helper.LogRequest(l.Logger, helper.OpAcceptOrder, map[string]interface{}{
		"order_id":     in.GetOrderId(),
		"companion_id": in.GetCompanionId(),
	})

	if in.GetOrderId() == 0 || in.GetCompanionId() == 0 {
		metrics.OrderAcceptTotal.WithLabelValues("invalid_argument").Inc()
		return nil, status.Error(codes.InvalidArgument, "order_id and companion_id are required")
	}

	// 使用分布式锁防止并发接单
	// 锁的 key 基于 order_id，防止同一订单被并发接单
	lockKey := fmt.Sprintf("accept_order:%d", in.GetOrderId())

	// 如果分布式锁未初始化，直接执行（降级处理）
	if l.svcCtx.DistributedLock == nil {
		helper.LogWarning(l.Logger, helper.OpAcceptOrder, "distributed lock not initialized, skipping lock", nil)
		return l.doAcceptOrder(in, start)
	}

	var result *order.AcceptOrderResponse
	var acceptErr error

	lockOptions := &lock.LockOptions{
		TTL:           30 * time.Second,
		RetryInterval: 100 * time.Millisecond,
		MaxWaitTime:   5 * time.Second,
	}

	lockStart := time.Now()
	err := l.svcCtx.DistributedLock.WithLock(l.ctx, lockKey, lockOptions, func() error {
		result, acceptErr = l.doAcceptOrder(in, start)
		return acceptErr
	})

	if err != nil {
		metrics.LockAcquireTotal.WithLabelValues("accept_order", "failed").Inc()
		if err == context.DeadlineExceeded || err == context.Canceled {
			metrics.OrderAcceptTotal.WithLabelValues("lock_timeout").Inc()
			return nil, status.Error(codes.DeadlineExceeded, "acquire lock timeout, please try again later")
		}
		helper.LogError(l.Logger, helper.OpAcceptOrder, "accept order with lock failed", err, nil)
		metrics.OrderAcceptTotal.WithLabelValues("lock_failed").Inc()
		return nil, status.Error(codes.Internal, "accept order failed")
	}

	metrics.LockAcquireTotal.WithLabelValues("accept_order", "success").Inc()
	metrics.LockAcquireDuration.WithLabelValues("accept_order").Observe(time.Since(lockStart).Seconds())

	return result, acceptErr
}

// doAcceptOrder 执行实际的订单接单逻辑（不加锁）
func (l *AcceptOrderLogic) doAcceptOrder(in *order.AcceptOrderRequest, start time.Time) (*order.AcceptOrderResponse, error) {
	db := l.svcCtx.DB.WithContext(l.ctx)

	var o model.Order
	if err := db.Where("id = ?", in.GetOrderId()).First(&o).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			metrics.OrderAcceptTotal.WithLabelValues("not_found").Inc()
			metrics.OrderAcceptDuration.WithLabelValues().Observe(time.Since(start).Seconds())
			return nil, status.Error(codes.NotFound, "order not found")
		}
		helper.LogError(l.Logger, helper.OpAcceptOrder, "get order failed", err, map[string]interface{}{
			"order_id": in.GetOrderId(),
		})
		metrics.OrderAcceptTotal.WithLabelValues("query_failed").Inc()
		metrics.OrderAcceptDuration.WithLabelValues().Observe(time.Since(start).Seconds())
		return nil, status.Error(codes.Internal, "get order failed")
	}

	if o.CompanionID != in.GetCompanionId() {
		helper.LogWarning(l.Logger, helper.OpAcceptOrder, "not allowed to accept this order", map[string]interface{}{
			"order_id":     in.GetOrderId(),
			"companion_id": in.GetCompanionId(),
			"expected_id":  o.CompanionID,
		})
		metrics.OrderAcceptTotal.WithLabelValues("not_authorized").Inc()
		metrics.OrderAcceptDuration.WithLabelValues().Observe(time.Since(start).Seconds())
		return nil, status.Error(codes.PermissionDenied, "not allowed to accept this order")
	}
	if o.Status != model.OrderStatusPaid && o.Status != model.OrderStatusCreated {
		helper.LogWarning(l.Logger, helper.OpAcceptOrder, "order is not pending for accept", map[string]interface{}{
			"order_id": in.GetOrderId(),
			"status":   o.Status,
		})
		metrics.OrderAcceptTotal.WithLabelValues("invalid_status").Inc()
		metrics.OrderAcceptDuration.WithLabelValues().Observe(time.Since(start).Seconds())
		return nil, status.Error(codes.FailedPrecondition, "order is not pending for accept")
	}

	now := time.Now()
	o.Status = model.OrderStatusAccepted
	o.AcceptedAt = &now

	if err := db.Save(&o).Error; err != nil {
		helper.LogError(l.Logger, helper.OpAcceptOrder, "accept order failed", err, map[string]interface{}{
			"order_id": in.GetOrderId(),
		})
		metrics.OrderAcceptTotal.WithLabelValues("save_failed").Inc()
		metrics.OrderAcceptDuration.WithLabelValues().Observe(time.Since(start).Seconds())
		return nil, status.Error(codes.Internal, "accept order failed")
	}

	if l.svcCtx.UserRPC != nil {
		_, err := l.svcCtx.UserRPC.UpdateCompanionProfile(l.ctx, &userclient.UpdateCompanionProfileRequest{
			UserId: in.GetCompanionId(),
			Status: 2,
		})
		if err != nil {
			if st, ok := status.FromError(err); ok {
				helper.LogWarning(l.Logger, helper.OpAcceptOrder, "update companion status to busy failed", map[string]interface{}{
					"companion_id": in.GetCompanionId(),
					"code":         st.Code(),
					"message":      st.Message(),
				})
			} else {
				helper.LogWarning(l.Logger, helper.OpAcceptOrder, "update companion status to busy failed", map[string]interface{}{
					"companion_id": in.GetCompanionId(),
					"error":        err.Error(),
				})
			}
		}
	}

	metrics.OrderAcceptTotal.WithLabelValues("success").Inc()
	metrics.OrderAcceptDuration.WithLabelValues().Observe(time.Since(start).Seconds())
	metrics.OrderStatusTransition.WithLabelValues(fmt.Sprintf("%d", model.OrderStatusPaid), fmt.Sprintf("%d", model.OrderStatusAccepted)).Inc()

	helper.LogSuccess(l.Logger, helper.OpAcceptOrder, map[string]interface{}{
		"order_id":     o.ID,
		"order_no":     o.OrderNo,
		"companion_id": o.CompanionID,
		"boss_id":      o.BossID,
	})

	return &order.AcceptOrderResponse{
		Order: toOrderInfo(&o),
	}, nil
}
