package logic

import (
	"context"
	"time"

	"SLGaming/back/services/order/internal/helper"
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
	helper.LogRequest(l.Logger, helper.OpAcceptOrder, map[string]interface{}{
		"order_id":     in.GetOrderId(),
		"companion_id": in.GetCompanionId(),
	})

	if in.GetOrderId() == 0 || in.GetCompanionId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "order_id and companion_id are required")
	}

	db := l.svcCtx.DB.WithContext(l.ctx)

	var o model.Order
	if err := db.Where("id = ?", in.GetOrderId()).First(&o).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Error(codes.NotFound, "order not found")
		}
		helper.LogError(l.Logger, helper.OpAcceptOrder, "get order failed", err, map[string]interface{}{
			"order_id": in.GetOrderId(),
		})
		return nil, status.Error(codes.Internal, "get order failed")
	}

	if o.CompanionID != in.GetCompanionId() {
		helper.LogWarning(l.Logger, helper.OpAcceptOrder, "not allowed to accept this order", map[string]interface{}{
			"order_id":     in.GetOrderId(),
			"companion_id": in.GetCompanionId(),
			"expected_id":  o.CompanionID,
		})
		return nil, status.Error(codes.PermissionDenied, "not allowed to accept this order")
	}
	if o.Status != model.OrderStatusPaid && o.Status != model.OrderStatusCreated {
		helper.LogWarning(l.Logger, helper.OpAcceptOrder, "order is not pending for accept", map[string]interface{}{
			"order_id": in.GetOrderId(),
			"status":   o.Status,
		})
		return nil, status.Error(codes.FailedPrecondition, "order is not pending for accept")
	}

	now := time.Now()
	o.Status = model.OrderStatusAccepted
	o.AcceptedAt = &now

	if err := db.Save(&o).Error; err != nil {
		helper.LogError(l.Logger, helper.OpAcceptOrder, "accept order failed", err, map[string]interface{}{
			"order_id": in.GetOrderId(),
		})
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
