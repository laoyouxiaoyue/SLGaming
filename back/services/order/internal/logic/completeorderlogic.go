package logic

import (
	"context"
	"encoding/json"
	"fmt"

	"SLGaming/back/services/order/internal/helper"
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
	helper.LogRequest(l.Logger, helper.OpCompleteOrder, map[string]interface{}{
		"order_id":    in.GetOrderId(),
		"operator_id": in.GetOperatorId(),
	})

	if in.GetOrderId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "order_id is required")
	}
	if in.GetOperatorId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "operator_id is required")
	}

	db := l.svcCtx.DB.WithContext(l.ctx)

	var o model.Order
	if err := db.Where("id = ?", in.GetOrderId()).First(&o).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Error(codes.NotFound, "order not found")
		}
		helper.LogError(l.Logger, helper.OpCompleteOrder, "get order failed", err, map[string]interface{}{
			"order_id": in.GetOrderId(),
		})
		return nil, status.Error(codes.Internal, "get order failed")
	}

	if o.BossID != in.GetOperatorId() {
		if l.svcCtx.UserRPC == nil {
			return nil, status.Error(codes.PermissionDenied, "only boss or admin can complete order")
		}
		userResp, err := l.svcCtx.UserRPC.GetUser(l.ctx, &userclient.GetUserRequest{Id: in.GetOperatorId()})
		if err != nil {
			helper.LogError(l.Logger, helper.OpCompleteOrder, "get operator role failed", err, map[string]interface{}{
				"operator_id": in.GetOperatorId(),
			})
			return nil, status.Error(codes.Internal, "get operator role failed")
		}
		if userResp.GetUser() == nil || userResp.GetUser().GetRole() != 3 {
			helper.LogWarning(l.Logger, helper.OpCompleteOrder, "permission denied: not boss or admin", map[string]interface{}{
				"order_id":    in.GetOrderId(),
				"operator_id": in.GetOperatorId(),
				"boss_id":     o.BossID,
			})
			return nil, status.Error(codes.PermissionDenied, "only boss or admin can complete order")
		}
	}

	if o.Status != model.OrderStatusInService && o.Status != model.OrderStatusAccepted {
		helper.LogWarning(l.Logger, helper.OpCompleteOrder, "order is not in progress", map[string]interface{}{
			"order_id": in.GetOrderId(),
			"status":   o.Status,
		})
		return nil, status.Error(codes.FailedPrecondition, "order is not in progress")
	}

	if l.svcCtx.OrderEventTxProducer == nil {
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
		return nil, status.Error(codes.Internal, "marshal completed event failed")
	}
	msg := primitive.NewMessage(orderMQ.OrderEventTopic(), msgBody)
	msg.WithTag(orderMQ.EventTypeCompleted())

	txRes, err := l.svcCtx.OrderEventTxProducer.SendMessageInTransaction(l.ctx, msg)
	if err != nil {
		helper.LogError(l.Logger, helper.OpCompleteOrder, "send transactional message failed", err, map[string]interface{}{
			"result": fmt.Sprintf("%+v", txRes),
		})
		return nil, status.Error(codes.Internal, "complete order failed")
	}

	var updatedOrder model.Order
	if err := db.Where("order_no = ?", o.OrderNo).First(&updatedOrder).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			helper.LogError(l.Logger, helper.OpCompleteOrder, "complete order transaction rolled back", err, map[string]interface{}{
				"order_no": o.OrderNo,
			})
			return nil, status.Error(codes.Internal, "complete order transaction rolled back")
		}
		helper.LogError(l.Logger, helper.OpCompleteOrder, "query order after transactional message failed", err, nil)
		return nil, status.Error(codes.Internal, "complete order failed")
	}

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
