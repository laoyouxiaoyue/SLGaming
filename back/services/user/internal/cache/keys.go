package cache

import "time"

const (
	// 用户相关缓存键
	UserInfoKey       = "user:info:%d"
	FollowerCountKey  = "user:follower_count:%d"
	FollowingCountKey = "user:following_count:%d"
	UserWalletKey     = "user:wallet:%d"

	// 陪玩相关缓存键
	CompanionInfoKey    = "companion:info:%d"
	CompanionListKey    = "companion:list:%s"
	CompanionRankingKey = "companion:ranking:%s"

	// 订单相关缓存键
	OrderInfoKey = "order:info:%d"
	OrderListKey = "order:list:%d:%s"

	// 游戏技能相关缓存键
	GameSkillListKey = "game:skill:list"

	// 缓存过期时间
	UserInfoExpire         = 30 * time.Minute
	CountCacheExpire       = 1 * time.Hour
	UserWalletExpire       = 15 * time.Minute
	CompanionInfoExpire    = 30 * time.Minute
	CompanionListExpire    = 5 * time.Minute
	CompanionRankingExpire = 10 * time.Minute
	OrderInfoExpire        = 2 * time.Hour
	OrderListExpire        = 15 * time.Minute
	GameSkillListExpire    = 1 * time.Hour
)
