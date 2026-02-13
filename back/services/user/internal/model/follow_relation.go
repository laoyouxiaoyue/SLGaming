package model

import (
	"SLGaming/back/pkg/snowflake"
	"time"

	"gorm.io/gorm"
)

// FollowRelation 关注关系表
type FollowRelation struct {
	BaseModel

	// 关注者ID
	FollowerID uint64 `gorm:"not null;index:idx_follower_following,unique;index:idx_follower_followed_at" json:"follower_id"`

	// 被关注者ID
	FollowingID uint64 `gorm:"not null;index:idx_follower_following,unique;index:idx_following_followed_at" json:"following_id"`

	// 关注时间
	FollowedAt time.Time `gorm:"autoCreateTime;index:idx_follower_followed_at;index:idx_following_followed_at" json:"followed_at"`
}

func (FollowRelation) TableName() string {
	return "follow_relations"
}

// BeforeCreate 创建前钩子
func (fr *FollowRelation) BeforeCreate(tx *gorm.DB) error {
	// 生成雪花主键
	if fr.ID == 0 {
		fr.ID = uint64(snowflake.GenID())
	}
	return nil
}
