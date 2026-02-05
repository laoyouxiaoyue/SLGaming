package tx

import (
	"context"

	"SLGaming/back/services/order/internal/model"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

// OrderPaymentPendingPayload 与 user 服务中消费的 payload 对应
// 这里可以在原有字段基础上扩展订单创建所需的字段，user 侧会忽略未知字段。
type OrderPaymentPendingPayload struct {
	OrderID    uint64 `json:"order_id"`
	OrderNo    string `json:"order_no"`
	BossID     uint64 `json:"boss_id"`
	Amount     int64  `json:"amount"`
	BizOrderID string `json:"biz_order_id"`

	// 扩展字段：用于在本地事务中创建订单
	CompanionID   uint64 `json:"companion_id"`
	GameName      string `json:"game_name"`
	DurationHours int32  `json:"duration_hours"`
	PricePerHour  int64  `json:"price_per_hour"`
}

// ExecuteCreateOrderTx 在本地事务中创建订单记录，如果订单已存在则幂等返回。
// 返回 error 为 nil 表示事务可提交；非 nil 表示需要回滚事务消息。
func ExecuteCreateOrderTx(ctx context.Context, db *gorm.DB, p *OrderPaymentPendingPayload) error {
	if db == nil || p == nil {
		return gorm.ErrInvalidDB
	}
	if p.OrderNo == "" || p.BossID == 0 || p.CompanionID == 0 || p.Amount <= 0 {
		logx.Errorf("ExecuteCreateOrderTx: invalid payload: %+v", p)
		return gorm.ErrInvalidData
	}

	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 幂等：如果订单已存在（按 order_no），则直接返回
		var existed model.Order
		if err := tx.Where("order_no = ?", p.OrderNo).First(&existed).Error; err == nil {
			return nil
		} else if err != nil && err != gorm.ErrRecordNotFound {
			return err
		}

		if p.DurationHours <= 0 {
			logx.Errorf("ExecuteCreateOrderTx: invalid duration_hours: %d", p.DurationHours)
			return gorm.ErrInvalidData
		}

		o := &model.Order{
			BossID:        p.BossID,
			CompanionID:   p.CompanionID,
			GameName:      p.GameName,
			DurationHours: p.DurationHours,
			PricePerHour:  p.PricePerHour,
			TotalAmount:   p.Amount,
			Status:        model.OrderStatusCreated,
			OrderNo:       p.OrderNo,
		}

		if err := tx.Create(o).Error; err != nil {
			return err
		}
		return nil
	})
}

// CheckCreateOrderTx 事务回查：根据 OrderNo 判断本地事务是否成功。
// 返回 (true, nil) 表示应提交消息；(false, nil) 表示应回滚；error 表示保持 UNKNOW。
func CheckCreateOrderTx(ctx context.Context, db *gorm.DB, p *OrderPaymentPendingPayload) (bool, error) {
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
		logx.Errorf("CheckCreateOrderTx: query order failed: %v", err)
		return false, err
	}

	// 查到订单，认为本地事务成功
	return true, nil
}
