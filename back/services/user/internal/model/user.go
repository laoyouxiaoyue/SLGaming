package model

import (
	"SLGaming/back/pkg/snowflake"
	"errors"
	"gorm.io/gorm"
	"math/rand"
	"time"
)

// BaseModel 基础模型
type BaseModel struct {
	// ID：系统内部唯一标识，雪花算法 (19位)，用于数据库关联
	ID        uint64         `gorm:"primaryKey;autoIncrement:false" json:"id,string"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

type User struct {
	BaseModel

	// UID：展示给用户看的 ID (靓号)
	// 特点：短(6-10位)，唯一，必须有索引用于搜索
	UID uint64 `gorm:"uniqueIndex;not null;comment:用户展示ID(靓号)" json:"uid"`

	Nickname string `gorm:"size:64;not null;default:'';comment:昵称" json:"nickname"`
	Password string `gorm:"size:128;not null;comment:加密密码" json:"password"`
	Phone    string `gorm:"size:20;uniqueIndex;not null;comment:手机号" json:"phone"`
}

func (u *User) TableName() string {
	return "users"
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	// 1. 生成系统主键 (雪花算法，必须有)
	if u.ID == 0 {
		u.ID = uint64(snowflake.GenID())
	}

	// 2. 生成展示 UID (如果已经有了就不生成)
	if u.UID == 0 {
		// 重试机制：最多尝试 5 次
		for i := 0; i < 5; i++ {
			// 生成一个随机 8 位数 (10000000 - 99999999)
			code := uint64(rand.Intn(90000000) + 10000000)

			// 检查数据库里有没有这个 UID
			// 注意：这里需要确保 UID 字段在数据库里有唯一索引 (UniqueIndex)
			var count int64
			tx.Model(&User{}).Where("uid = ?", code).Count(&count)

			if count == 0 {
				u.UID = code
				return nil // 成功找到一个没用的 ID，赋值并退出
			}
			// 如果 count > 0，说明重复了，循环继续，生成下一个随机数
		}
		// 5次都失败（概率极低），返回错误
		return errors.New("生成UID失败，请稍后重试")
	}
	return nil
}
