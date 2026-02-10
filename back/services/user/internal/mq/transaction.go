package mq

import (
	"context"
	"encoding/json"

	"SLGaming/back/services/user/internal/model"

	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	userEventTopic           = "order_events"           // 与订单服务共用 topic，靠 Tag 区分
	followEventTopic         = "follow_events"          // 关注事件独立 topic
	eventTypeRefundSucceeded = "ORDER_REFUND_SUCCEEDED" // 退款成功事件
	eventTypeFollowUser      = "USER_FOLLOW"            // 关注用户事件
	eventTypeUnfollowUser    = "USER_UNFOLLOW"          // 取消关注用户事件
)

// UserEventTopic 返回用户领域事件使用的 RocketMQ Topic
func UserEventTopic() string {
	return userEventTopic
}

// FollowEventTopic 返回关注事件使用的 RocketMQ Topic
func FollowEventTopic() string {
	return followEventTopic
}

// EventTypeRefundSucceeded 返回退款成功事件类型
func EventTypeRefundSucceeded() string {
	return eventTypeRefundSucceeded
}

// EventTypeFollowUser 返回关注用户事件类型
func EventTypeFollowUser() string {
	return eventTypeFollowUser
}

// EventTypeUnfollowUser 返回取消关注用户事件类型
func EventTypeUnfollowUser() string {
	return eventTypeUnfollowUser
}

// RefundSucceededPayload 用户退款成功事件负载
// 由用户服务产生，订单服务消费，用于将订单状态 CANCEL_REFUNDING -> CANCELLED。
type RefundSucceededPayload struct {
	OrderID    uint64 `json:"order_id"`
	OrderNo    string `json:"order_no"`
	BizOrderID string `json:"biz_order_id"`
	UserID     uint64 `json:"user_id"`
	Amount     int64  `json:"amount"`
}

// FollowUserPayload 关注用户事件负载
type FollowUserPayload struct {
	FollowerID  uint64 `json:"follower_id"`  // 关注者ID
	FollowingID uint64 `json:"following_id"` // 被关注者ID
}

// UnfollowUserPayload 取消关注用户事件负载
type UnfollowUserPayload struct {
	FollowerID  uint64 `json:"follower_id"`  // 关注者ID
	FollowingID uint64 `json:"following_id"` // 被关注者ID
}

// ExecuteUserEventTx 用户领域事件本地事务执行器
// 处理 ORDER_REFUND_SUCCEEDED：在一个本地事务中完成钱包退款和流水记录
func ExecuteUserEventTx(ctx context.Context, db *gorm.DB, msg *primitive.Message) primitive.LocalTransactionState {
	if msg == nil || db == nil {
		return primitive.RollbackMessageState
	}

	if msg.GetTags() != eventTypeRefundSucceeded {
		// 非当前支持的事件类型，直接回滚
		return primitive.RollbackMessageState
	}

	var payload RefundSucceededPayload
	if err := json.Unmarshal(msg.Body, &payload); err != nil {
		logx.Errorf("ExecuteUserEventTx: unmarshal refund succeeded payload failed: %v, body=%s", err, string(msg.Body))
		return primitive.RollbackMessageState
	}
	if payload.UserID == 0 || payload.Amount <= 0 || payload.BizOrderID == "" {
		logx.Errorf("ExecuteUserEventTx: invalid refund payload: %+v", payload)
		return primitive.RollbackMessageState
	}

	// 在本地事务中完成钱包退款和流水记录
	if err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var wallet model.UserWallet

		// 1. 查询或创建钱包（行级锁）
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ?", payload.UserID).
			First(&wallet).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				wallet = model.UserWallet{
					UserID:        payload.UserID,
					Balance:       0,
					FrozenBalance: 0,
				}
				if err := tx.Create(&wallet).Error; err != nil {
					return err
				}
			} else {
				return err
			}
		}

		// 2. 幂等检查：如果已经存在相同 biz_order_id + REFUND 的流水，则视为成功
		var existed model.WalletTransaction
		if err := tx.
			Where("user_id = ? AND type = ? AND biz_order_id = ?",
				payload.UserID, "REFUND", payload.BizOrderID).
			First(&existed).Error; err == nil {
			return nil
		} else if err != nil && err != gorm.ErrRecordNotFound {
			return err
		}

		// 3. 更新余额
		before := wallet.Balance
		wallet.Balance = before + payload.Amount
		if err := tx.Save(&wallet).Error; err != nil {
			return err
		}

		// 4. 写入流水
		tr := &model.WalletTransaction{
			UserID:        payload.UserID,
			WalletID:      wallet.ID,
			ChangeAmount:  payload.Amount,
			BeforeBalance: before,
			AfterBalance:  wallet.Balance,
			Type:          "REFUND",
			BizOrderID:    payload.BizOrderID,
			Remark:        "order refund",
		}
		if err := tx.Create(tr).Error; err != nil {
			return err
		}

		return nil
	}); err != nil {
		logx.Errorf("ExecuteUserEventTx: wallet refund transaction failed: %v, payload=%+v", err, payload)
		return primitive.RollbackMessageState
	}

	return primitive.CommitMessageState
}

// CheckUserEventTx 用户领域事件本地事务回查
// 对于 ORDER_REFUND_SUCCEEDED：根据钱包流水表中是否存在 REFUND + biz_order_id 判断事务是否成功
func CheckUserEventTx(ctx context.Context, db *gorm.DB, msg *primitive.Message) primitive.LocalTransactionState {
	if msg == nil || db == nil {
		return primitive.UnknowState
	}

	if msg.GetTags() != eventTypeRefundSucceeded {
		return primitive.RollbackMessageState
	}

	var payload RefundSucceededPayload
	if err := json.Unmarshal(msg.Body, &payload); err != nil {
		logx.Errorf("CheckUserEventTx: unmarshal refund succeeded payload failed: %v, body=%s", err, string(msg.Body))
		return primitive.UnknowState
	}
	if payload.BizOrderID == "" || payload.UserID == 0 {
		return primitive.RollbackMessageState
	}

	var tr model.WalletTransaction
	if err := db.WithContext(ctx).
		Where("user_id = ? AND type = ? AND biz_order_id = ?",
			payload.UserID, "REFUND", payload.BizOrderID).
		First(&tr).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// 没有找到对应流水，认为本地事务未成功，回滚消息
			return primitive.RollbackMessageState
		}
		logx.Errorf("CheckUserEventTx: query wallet transaction failed: %v", err)
		return primitive.UnknowState
	}

	// 找到对应退款流水，认为本地事务成功
	return primitive.CommitMessageState
}
