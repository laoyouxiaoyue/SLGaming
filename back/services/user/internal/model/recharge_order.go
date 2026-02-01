package model

import (
	"SLGaming/back/pkg/snowflake"
	"time"

	"gorm.io/gorm"
)

const (
	RechargeStatusPending = 0
	RechargeStatusSuccess = 1
	RechargeStatusFailed  = 2
	RechargeStatusClosed  = 3
)

// RechargeOrder 充值订单
// 说明：记录充值订单的支付状态与关键信息，便于对账与追溯。
type RechargeOrder struct {
	BaseModel

	// 用户ID
	UserID uint64 `gorm:"not null;index;comment:用户ID" json:"user_id,string"`

	// 充值单号（业务订单号）
	OrderNo string `gorm:"size:64;not null;uniqueIndex;comment:充值单号" json:"order_no"`

	// 充值金额（分/帅币）
	Amount int64 `gorm:"not null;comment:充值金额" json:"amount"`

	// 状态：0=待支付,1=成功,2=失败,3=关闭
	Status int `gorm:"not null;default:0;comment:订单状态" json:"status"`

	// 支付方式/渠道（如 alipay）
	PayType string `gorm:"size:32;not null;default:'alipay';comment:支付方式" json:"pay_type"`

	// 支付平台交易号（可选）
	TradeNo string `gorm:"size:128;comment:平台交易号" json:"trade_no"`

	// 备注
	Remark string `gorm:"size:255;comment:备注" json:"remark"`

	// 支付完成时间（可选）
	PaidAt *time.Time `gorm:"comment:支付完成时间" json:"paid_at"`
}

func (r *RechargeOrder) TableName() string {
	return "recharge_orders"
}

// BeforeCreate 创建前钩子：生成 ID
func (r *RechargeOrder) BeforeCreate(tx *gorm.DB) error {
	if r.ID == 0 {
		r.ID = uint64(snowflake.GenID())
	}
	return nil
}
