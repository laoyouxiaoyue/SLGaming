package model

import (
	"time"
)

// ProcessedMessage 消息处理记录，用于幂等性检查
type ProcessedMessage struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement"`
	MessageID string    `gorm:"uniqueIndex;size:64;not null;comment:消息唯一标识"`
	EventType string    `gorm:"size:32;not null;comment:事件类型"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

// TableName 返回表名
func (ProcessedMessage) TableName() string {
	return "processed_messages"
}
