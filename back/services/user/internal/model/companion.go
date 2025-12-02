package model

// CompanionProfile 陪玩信息表
// 只有角色为陪玩（RoleCompanion）的用户才会有此记录
type CompanionProfile struct {
	BaseModel

	// 关联的用户 ID（users.id），一对一关系
	UserID uint64 `gorm:"uniqueIndex;not null;comment:用户ID" json:"user_id,string"`

	// 游戏技能列表（JSON 格式，如：["王者荣耀","和平精英","英雄联盟"]）
	GameSkills string `gorm:"type:text;comment:游戏技能列表(JSON)" json:"game_skills"`

	// 每小时价格（单位：帅币）
	PricePerHour int64 `gorm:"not null;default:0;comment:每小时价格(帅币)" json:"price_per_hour"`

	// 陪玩状态：0=离线, 1=在线, 2=忙碌
	Status int `gorm:"not null;default:0;index;comment:状态(0=离线,1=在线,2=忙碌)" json:"status"`

	// 评分（0-5分，保留2位小数）
	Rating float64 `gorm:"type:decimal(3,2);not null;default:0;comment:评分(0-5)" json:"rating"`

	// 总接单数
	TotalOrders int64 `gorm:"not null;default:0;comment:总接单数" json:"total_orders"`

	// 是否认证（平台认证的陪玩）
	IsVerified bool `gorm:"not null;default:false;index;comment:是否认证" json:"is_verified"`
}

func (c *CompanionProfile) TableName() string {
	return "companion_profiles"
}

// 陪玩状态常量
const (
	CompanionStatusOffline = 0 // 离线
	CompanionStatusOnline  = 1 // 在线
	CompanionStatusBusy    = 2 // 忙碌
)

// IsOnline 判断是否在线
func (c *CompanionProfile) IsOnline() bool {
	return c.Status == CompanionStatusOnline
}

// IsAvailable 判断是否可接单（在线且不忙碌）
func (c *CompanionProfile) IsAvailable() bool {
	return c.Status == CompanionStatusOnline
}
