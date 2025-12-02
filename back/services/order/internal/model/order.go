package model

import (
	"SLGaming/back/pkg/snowflake"
	"time"

	"gorm.io/gorm"
)

// 订单状态常量
const (
	OrderStatusCreated   = 1 // 已创建，待支付/扣款
	OrderStatusPaid      = 2 // 已支付/已扣款，待陪玩接单
	OrderStatusAccepted  = 3 // 陪玩已接单，待开始服务
	OrderStatusInService = 4 // 服务中
	OrderStatusCompleted = 5 // 已完成
	OrderStatusCancelled = 6 // 已取消
	OrderStatusRated     = 7 // 已评价
)

// BaseModel 基础模型（与 user 服务风格保持一致）
type BaseModel struct {
	ID        uint64         `gorm:"primaryKey;autoIncrement:false" json:"id,string"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

// Order 订单表
type Order struct {
	BaseModel

	// 展示订单号（例如：202501010001），便于用户查看和客服查询
	OrderNo string `gorm:"size:32;uniqueIndex;not null;comment:订单号" json:"order_no"`

	// 老板 & 陪玩
	BossID      uint64 `gorm:"not null;index;comment:老板ID" json:"boss_id,string"`
	CompanionID uint64 `gorm:"not null;index;comment:陪玩ID" json:"companion_id,string"`

	// 游戏信息
	GameName string `gorm:"size:64;not null;comment:游戏名称" json:"game_name"`
	GameMode string `gorm:"size:64;comment:游戏模式/段位" json:"game_mode"`

	// 服务时长（分钟）
	DurationMinutes int32 `gorm:"not null;comment:服务时长(分钟)" json:"duration_minutes"`

	// 金额信息（帅币）
	PricePerHour int64 `gorm:"not null;default:0;comment:每小时价格(帅币)" json:"price_per_hour"`
	TotalAmount  int64 `gorm:"not null;default:0;comment:订单总价(帅币)" json:"total_amount"`

	// 状态
	Status int32 `gorm:"not null;index;comment:订单状态" json:"status"`

	// 时间字段
	PaidAt      *time.Time `gorm:"comment:支付时间" json:"paid_at"`
	AcceptedAt  *time.Time `gorm:"comment:接单时间" json:"accepted_at"`
	StartAt     *time.Time `gorm:"comment:开始服务时间" json:"start_at"`
	CompletedAt *time.Time `gorm:"comment:完成时间" json:"completed_at"`
	CancelledAt *time.Time `gorm:"comment:取消时间" json:"cancelled_at"`

	// 评价
	Rating  float64 `gorm:"type:decimal(3,2);not null;default:0;comment:评分(0-5)" json:"rating"`
	Comment string  `gorm:"type:text;comment:评价内容" json:"comment"`

	// 取消信息
	CancelReason string `gorm:"size:255;comment:取消原因" json:"cancel_reason"`
}

func (o *Order) TableName() string {
	return "orders"
}

// BeforeCreate 钩子：生成雪花主键
func (o *Order) BeforeCreate(tx *gorm.DB) error {
	if o.ID == 0 {
		o.ID = uint64(snowflake.GenID())
	}
	return nil
}
