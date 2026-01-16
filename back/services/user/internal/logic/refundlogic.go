package logic

import (
	"context"
	"encoding/json"

	"SLGaming/back/services/user/internal/helper"
	userMQ "SLGaming/back/services/user/internal/mq"
	"SLGaming/back/services/user/internal/svc"

	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RefundLogic 退款逻辑（幂等）
// 同一笔业务单号（bizOrderId）多次调用，只会实际退款一次，并通过 RocketMQ 事务消息发送退款成功事件
type RefundLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRefundLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RefundLogic {
	return &RefundLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Refund 为指定用户执行退款（amount 为正数），bizOrderId 用于幂等控制
// orderNo 用于在退款成功事件中携带订单号，便于订单服务更新状态
func (l *RefundLogic) Refund(userID uint64, amount int64, bizOrderID, orderNo, remark string) error {
	if userID == 0 {
		return status.Error(codes.InvalidArgument, "user_id is required")
	}
	if amount <= 0 {
		return status.Error(codes.InvalidArgument, "amount must be positive")
	}
	if bizOrderID == "" {
		return status.Error(codes.InvalidArgument, "biz_order_id is required for refund idempotency")
	}

	// 使用 RocketMQ 半消息实现：发送 ORDER_REFUND_SUCCEEDED 事务消息，
	// 在本地事务（ExecuteUserEventTx）中完成钱包退款和流水记录。
	if l.svcCtx.EventTxProducer == nil {
		return status.Error(codes.FailedPrecondition, "user transaction producer not initialized")
	}

	payload := &userMQ.RefundSucceededPayload{
		OrderID:    0, // 如需精确到订单 ID，可在调用处扩展
		OrderNo:    orderNo,
		BizOrderID: bizOrderID,
		UserID:     userID,
		Amount:     amount,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		helper.LogError(l.Logger, helper.OpRefund, "marshal refund tx payload failed", err, map[string]interface{}{
			"user_id":      userID,
			"biz_order_id": bizOrderID,
		})
		return status.Error(codes.Internal, "marshal refund payload failed")
	}

	msg := primitive.NewMessage(userMQ.UserEventTopic(), body)
	msg.WithTag(userMQ.EventTypeRefundSucceeded())

	txRes, err := l.svcCtx.EventTxProducer.SendMessageInTransaction(l.ctx, msg)
	if err != nil {
		helper.LogError(l.Logger, helper.OpRefund, "send refund transactional message failed", err, map[string]interface{}{
			"user_id":      userID,
			"biz_order_id": bizOrderID,
			"tx_result":    txRes,
		})
		return status.Error(codes.Internal, "send refund transactional message failed")
	}

	// 此时本地事务（ExecuteUserEventTx）已执行完成，但是否成功由 RocketMQ + CheckUserEventTx 决定；
	// 对调用方而言，视为“已发起退款”，幂等由 bizOrderID 保证。
	return nil
}
