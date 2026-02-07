package logic

import (
	"context"
	"encoding/json"

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
		l.Errorf("get order failed: %v", err)
		return nil, status.Error(codes.Internal, "get order failed")
	}

	if o.BossID != in.GetOperatorId() {
		if l.svcCtx.UserRPC == nil {
			return nil, status.Error(codes.PermissionDenied, "only boss or admin can complete order")
		}
		userResp, err := l.svcCtx.UserRPC.GetUser(l.ctx, &userclient.GetUserRequest{Id: in.GetOperatorId()})
		if err != nil {
			l.Errorf("get operator role failed: %v", err)
			return nil, status.Error(codes.Internal, "get operator role failed")
		}
		if userResp.GetUser() == nil || userResp.GetUser().GetRole() != 3 {
			return nil, status.Error(codes.PermissionDenied, "only boss or admin can complete order")
		}
	}

	if o.Status != model.OrderStatusInService && o.Status != model.OrderStatusAccepted {
		return nil, status.Error(codes.FailedPrecondition, "order is not in progress")
	}

	// 使用 RocketMQ 事务消息发送 ORDER_COMPLETED，并在本地事务中更新订单状态
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

	// 构造事务消息
	msgBody, err := json.Marshal(payload)
	if err != nil {
		l.Errorf("marshal completed payload failed: %v", err)
		return nil, status.Error(codes.Internal, "marshal completed event failed")
	}
	msg := primitive.NewMessage(orderMQ.OrderEventTopic(), msgBody)
	msg.WithTag(orderMQ.EventTypeCompleted())

	txRes, err := l.svcCtx.OrderEventTxProducer.SendMessageInTransaction(l.ctx, msg)
	if err != nil {
		l.Errorf("send transactional message failed: %v, result=%+v", err, txRes)
		return nil, status.Error(codes.Internal, "complete order failed")
	}

	// 此时本地事务（ExecuteOrderTx -> ExecuteCompleteOrderTx）已经执行完成，
	// 但是否成功需要通过查询订单确认
	var updatedOrder model.Order
	if err := db.Where("order_no = ?", o.OrderNo).First(&updatedOrder).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			l.Errorf("complete order transaction rolled back, order not found, order_no=%s", o.OrderNo)
			return nil, status.Error(codes.Internal, "complete order transaction rolled back")
		}
		l.Errorf("query order after transactional message failed: %v", err)
		return nil, status.Error(codes.Internal, "complete order failed")
	}

	return &order.CompleteOrderResponse{
		Order: toOrderInfo(&updatedOrder),
	}, nil
}
