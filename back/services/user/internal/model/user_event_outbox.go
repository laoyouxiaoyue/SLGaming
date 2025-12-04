package model

import (
	"SLGaming/back/pkg/snowflake"
	"time"

	"gorm.io/gorm"
)

// UserEventOutbox 用户领域事件 Outbox 表
// 用于保证钱包/退款相关事件可靠投递到消息队列（RocketMQ）
type UserEventOutbox struct {
	ID uint64 `gorm:"primaryKey;autoIncrement:false" json:"id,string"`

	// 事件类型，如：ORDER_REFUND_SUCCEEDED 等
	EventType string `gorm:"size:64;index;not null;comment:事件类型" json:"event_type"`

	// 事件负载，JSON 字符串
	Payload string `gorm:"type:text;not null;comment:事件负载(JSON)" json:"payload"`

	// 状态：PENDING / SENT / FAILED
	Status string `gorm:"size:32;index;not null;default:'PENDING';comment:发送状态" json:"status"`

	// 可选：最后一次错误信息，便于排查
	LastError string `gorm:"type:text;comment:最后一次发送错误" json:"last_error"`

	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

func (e *UserEventOutbox) TableName() string {
	return "user_event_outbox"
}

// BeforeCreate 钩子：生成雪花主键
func (e *UserEventOutbox) BeforeCreate(tx *gorm.DB) error {
	if e.ID == 0 {
		e.ID = uint64(snowflake.GenID())
	}
	return nil
}
