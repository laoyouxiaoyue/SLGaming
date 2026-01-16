package mq

import (
	"context"
	"encoding/json"

	"SLGaming/back/services/order/internal/tx"

	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

const (
	orderEventTopic         = "order_events"
	eventTypePaymentPending = "ORDER_PAYMENT_PENDING"
	eventTypeCancelled      = "ORDER_CANCELLED"
	eventTypeCompleted      = "ORDER_COMPLETED"
)

// OrderEventTopic 返回订单事件主题
func OrderEventTopic() string {
	return orderEventTopic
}

// EventTypePaymentPending 返回支付待处理事件类型
func EventTypePaymentPending() string {
	return eventTypePaymentPending
}

// EventTypeCancelled 返回订单取消事件类型
func EventTypeCancelled() string {
	return eventTypeCancelled
}

// EventTypeCompleted 返回订单完成事件类型
func EventTypeCompleted() string {
	return eventTypeCompleted
}

// NOTE: 所有订单事务的本地事务逻辑已迁移到 tx 包中。

// ExecuteOrderTx 通用的订单事务执行器：根据消息 tag 分发到不同的处理函数
func ExecuteOrderTx(ctx context.Context, db *gorm.DB, msg *primitive.Message) primitive.LocalTransactionState {
	if msg == nil || db == nil {
		return primitive.RollbackMessageState
	}

	eventType := msg.GetTags()
	switch eventType {
	case eventTypePaymentPending:
		var payload tx.OrderPaymentPendingPayload
		if err := json.Unmarshal(msg.Body, &payload); err != nil {
			logx.Errorf("ExecuteOrderTx: unmarshal payment pending payload failed: %v, body=%s", err, string(msg.Body))
			return primitive.RollbackMessageState
		}
		if err := tx.ExecuteCreateOrderTx(ctx, db, &payload); err != nil {
			return primitive.RollbackMessageState
		}
		return primitive.CommitMessageState
	case eventTypeCancelled:
		var payload tx.OrderCancelledPayload
		if err := json.Unmarshal(msg.Body, &payload); err != nil {
			logx.Errorf("ExecuteOrderTx: unmarshal cancelled payload failed: %v, body=%s", err, string(msg.Body))
			return primitive.RollbackMessageState
		}
		if err := tx.ExecuteCancelOrderTx(ctx, db, &payload); err != nil {
			return primitive.RollbackMessageState
		}
		return primitive.CommitMessageState
	case eventTypeCompleted:
		var payload tx.OrderCompletedPayload
		if err := json.Unmarshal(msg.Body, &payload); err != nil {
			logx.Errorf("ExecuteOrderTx: unmarshal completed payload failed: %v, body=%s", err, string(msg.Body))
			return primitive.RollbackMessageState
		}
		if err := tx.ExecuteCompleteOrderTx(ctx, db, &payload); err != nil {
			return primitive.RollbackMessageState
		}
		return primitive.CommitMessageState
	default:
		logx.Errorf("ExecuteOrderTx: unknown event type: %s", eventType)
		return primitive.RollbackMessageState
	}
}

// CheckOrderTx 通用的订单事务回查器：根据消息 tag 分发到不同的处理函数
func CheckOrderTx(ctx context.Context, db *gorm.DB, msg *primitive.Message) primitive.LocalTransactionState {
	if msg == nil || db == nil {
		return primitive.UnknowState
	}

	eventType := msg.GetTags()
	switch eventType {
	case eventTypePaymentPending:
		var payload tx.OrderPaymentPendingPayload
		if err := json.Unmarshal(msg.Body, &payload); err != nil {
			logx.Errorf("CheckOrderTx: unmarshal payment pending payload failed: %v, body=%s", err, string(msg.Body))
			return primitive.UnknowState
		}
		ok, err := tx.CheckCreateOrderTx(ctx, db, &payload)
		if err != nil {
			return primitive.UnknowState
		}
		if ok {
			return primitive.CommitMessageState
		}
		return primitive.RollbackMessageState
	case eventTypeCancelled:
		var payload tx.OrderCancelledPayload
		if err := json.Unmarshal(msg.Body, &payload); err != nil {
			logx.Errorf("CheckOrderTx: unmarshal cancelled payload failed: %v, body=%s", err, string(msg.Body))
			return primitive.UnknowState
		}
		ok, err := tx.CheckCancelOrderTx(ctx, db, &payload)
		if err != nil {
			return primitive.UnknowState
		}
		if ok {
			return primitive.CommitMessageState
		}
		return primitive.RollbackMessageState
	case eventTypeCompleted:
		var payload tx.OrderCompletedPayload
		if err := json.Unmarshal(msg.Body, &payload); err != nil {
			logx.Errorf("CheckOrderTx: unmarshal completed payload failed: %v, body=%s", err, string(msg.Body))
			return primitive.UnknowState
		}
		ok, err := tx.CheckCompleteOrderTx(ctx, db, &payload)
		if err != nil {
			return primitive.UnknowState
		}
		if ok {
			return primitive.CommitMessageState
		}
		return primitive.RollbackMessageState
	default:
		logx.Errorf("CheckOrderTx: unknown event type: %s", eventType)
		return primitive.UnknowState
	}
}
