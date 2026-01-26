package model

// GameSkill 游戏技能表
type GameSkill struct {
	BaseModel

	// 技能名称（唯一）
	Name string `gorm:"size:64;uniqueIndex;not null;comment:技能名称" json:"name"`

	// 技能描述（可选）
	Description string `gorm:"type:text;comment:技能描述" json:"description"`
}

func (g *GameSkill) TableName() string {
	return "game_skills"
}
