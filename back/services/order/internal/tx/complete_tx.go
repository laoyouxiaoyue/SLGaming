package tx

import (
	"context"
	"time"

	"SLGaming/back/services/order/internal/model"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

// OrderCompletedPayload 订单完成事件负载（与 user 服务中消费的 payload 对应）
type OrderCompletedPayload struct {
	OrderID     uint64 `json:"order_id"`
	OrderNo     string `json:"order_no"`
	BossID      uint64 `json:"boss_id"`
	CompanionID uint64 `json:"companion_id"`
	Amount      int64  `json:"amount"`
	BizOrderID  string `json:"biz_order_id"`
}

// ExecuteCompleteOrderTx 在本地事务中更新订单状态为已完成（COMPLETED）
// 返回 error 为 nil 表示事务可提交；非 nil 表示需要回滚事务消息。
func ExecuteCompleteOrderTx(ctx context.Context, db *gorm.DB, p *OrderCompletedPayload) error {
	if db == nil || p == nil {
		return gorm.ErrInvalidDB
	}
	if p.OrderNo == "" {
		logx.Errorf("ExecuteCompleteOrderTx: invalid payload: order_no is empty")
		return gorm.ErrInvalidData
	}

	now := time.Now()

	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var o model.Order
		if err := tx.Where("order_no = ?", p.OrderNo).First(&o).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				// 订单不存在，幂等返回
				return nil
			}
			return err
		}

		// 更新订单状态为已完成
		o.Status = model.OrderStatusCompleted
		o.CompletedAt = &now

		if err := tx.Save(&o).Error; err != nil {
			return err
		}
		return nil
	})
}

// CheckCompleteOrderTx 事务回查：根据订单状态判断本地事务是否成功。
// 返回 (true, nil) 表示应提交消息；(false, nil) 表示应回滚；error 表示保持 UNKNOW。
func CheckCompleteOrderTx(ctx context.Context, db *gorm.DB, p *OrderCompletedPayload) (bool, error) {
	if db == nil || p == nil {
		return false, gorm.ErrInvalidDB
	}
	if p.OrderNo == "" {
		return false, nil
	}

	var o model.Order
	if err := db.WithContext(ctx).Where("order_no = ?", p.OrderNo).First(&o).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// 本地没有这条订单，则回滚消息
			return false, nil
		}
		logx.Errorf("CheckCompleteOrderTx: query order failed: %v", err)
		return false, err
	}

	// 如果订单状态是 COMPLETED 或 RATED，认为本地事务成功
	if o.Status == model.OrderStatusCompleted || o.Status == model.OrderStatusRated {
		return true, nil
	}

	return false, nil
}
