package helper

import (
	"context"
	"strconv"

	"SLGaming/back/services/user/internal/model"
	"SLGaming/back/services/user/internal/svc"
	"SLGaming/back/services/user/user"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// PaginationParams 分页参数
type PaginationParams struct {
	Page     int
	PageSize int
	Start    int
	End      int
}

// NormalizePagination 规范化分页参数（默认每页10条）
func NormalizePagination(page, pageSize int32) PaginationParams {
	return NormalizePaginationWithDefault(page, pageSize, 10)
}

// NormalizePaginationWithDefault 规范化分页参数（支持自定义默认值）
func NormalizePaginationWithDefault(page, pageSize int32, defaultPageSize int) PaginationParams {
	p := int(page)
	if p < 1 {
		p = 1
	}
	ps := int(pageSize)
	if ps < 1 {
		ps = defaultPageSize
	}
	if ps > 100 {
		ps = 100
	}
	start := (p - 1) * ps
	end := start + ps - 1
	return PaginationParams{
		Page:     p,
		PageSize: ps,
		Start:    start,
		End:      end,
	}
}

// RankingItemBuilder 排名项构建器接口
type RankingItemBuilder interface {
	BuildRating(score int64) float64
	GetRatingFromProfile(profile *model.CompanionProfile) float64
}

// RatingRankingBuilder 评分排名构建器
type RatingRankingBuilder struct{}

func (b *RatingRankingBuilder) BuildRating(score int64) float64 {
	return float64(score) / 10000.0
}

func (b *RatingRankingBuilder) GetRatingFromProfile(profile *model.CompanionProfile) float64 {
	return 0
}

// OrdersRankingBuilder 接单数排名构建器
type OrdersRankingBuilder struct{}

func (b *OrdersRankingBuilder) BuildRating(score int64) float64 {
	return 0
}

func (b *OrdersRankingBuilder) GetRatingFromProfile(profile *model.CompanionProfile) float64 {
	return profile.Rating
}

// RankingQueryResult 排名查询结果
type RankingQueryResult struct {
	Rankings []*user.CompanionRankingItem
	Total    int32
	Page     int32
	PageSize int32
}

// QueryRankingWithProfiles 查询排行榜（ZSet只存前100）
func QueryRankingWithProfiles(
	ctx context.Context,
	svcCtx *svc.ServiceContext,
	logger logx.Logger,
	redisKey string,
	page, pageSize int32,
	builder RankingItemBuilder,
) (*RankingQueryResult, error) {
	// 规范化分页参数
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	start := (page - 1) * pageSize
	end := start + pageSize - 1

	// 限制最多100条（ZSet只存前100）
	if start >= 100 {
		return &RankingQueryResult{
			Rankings: []*user.CompanionRankingItem{},
			Total:    100,
			Page:     page,
			PageSize: pageSize,
		}, nil
	}
	if end >= 100 {
		end = 99
	}

	// 检查预热状态
	if !IsWarmupDone() {
		if IsWarmupRunning() {
			LogInfo(logger, OpGetCompanionRatingRanking, "ranking warmup in progress, query from mysql directly", nil)
		} else {
			LogInfo(logger, OpGetCompanionRatingRanking, "ranking warmup not started or failed, query from mysql directly", nil)
		}
		// 预热未完成，直接走MySQL查询（避免查询空的Redis）
		return queryRankingFromMySQL(ctx, svcCtx, logger, redisKey, start, end, builder)
	}

	// 从Redis查询（ZSet只存前100名）
	members, err := svcCtx.Redis.ZrevrangeWithScores(redisKey, int64(start), int64(end))
	if err != nil {
		LogError(logger, OpGetCompanionRatingRanking, "query ranking from redis failed, fallback to mysql", err, map[string]interface{}{
			"redis_key": redisKey,
		})
		// Redis故障，降级到MySQL查询
		return queryRankingFromMySQL(ctx, svcCtx, logger, redisKey, start, end, builder)
	}

	// 获取总数（最多100）
	total, _ := svcCtx.Redis.Zcard(redisKey)
	if total > 100 {
		total = 100
	}

	if len(members) == 0 {
		return &RankingQueryResult{
			Rankings: []*user.CompanionRankingItem{},
			Total:    int32(total),
			Page:     page,
			PageSize: pageSize,
		}, nil
	}

	// 收集user_ids
	userIDs := make([]uint64, 0, len(members))
	for _, member := range members {
		userID, err := strconv.ParseUint(member.Key, 10, 64)
		if err != nil {
			continue
		}
		userIDs = append(userIDs, userID)
	}

	if len(userIDs) == 0 {
		return &RankingQueryResult{
			Rankings: []*user.CompanionRankingItem{},
			Total:    int32(total),
			Page:     page,
			PageSize: pageSize,
		}, nil
	}

	// 使用JOIN一次性查询（避免多次查询）
	type RankingUser struct {
		ID          uint64  `gorm:"column:id"`
		Nickname    string  `gorm:"column:nickname"`
		AvatarURL   string  `gorm:"column:avatar_url"`
		Rating      float64 `gorm:"column:rating"`
		TotalOrders int64   `gorm:"column:total_orders"`
		IsVerified  bool    `gorm:"column:is_verified"`
	}

	db := svcCtx.DB().WithContext(ctx)
	var rankingUsers []RankingUser

	queryErr := db.Table("users").
		Select("users.id, users.nickname, users.avatar_url, "+
			"companion_profiles.rating, companion_profiles.total_orders, companion_profiles.is_verified").
		Joins("INNER JOIN companion_profiles ON users.id = companion_profiles.user_id").
		Where("users.id IN ?", userIDs).
		Find(&rankingUsers).Error

	if queryErr != nil {
		LogError(logger, OpGetCompanionRatingRanking, "query ranking users failed", queryErr, nil)
		return nil, status.Error(codes.Internal, "query ranking users failed")
	}

	// 构建user_id -> user信息的map
	userMap := make(map[uint64]*RankingUser)
	for i := range rankingUsers {
		userMap[rankingUsers[i].ID] = &rankingUsers[i]
	}

	// 组装结果（保持Redis返回的顺序）
	rankings := make([]*user.CompanionRankingItem, 0, len(members))
	for i, member := range members {
		userID, _ := strconv.ParseUint(member.Key, 10, 64)
		rank := int32(int(start) + i + 1)

		u, ok := userMap[userID]
		if !ok {
			continue
		}

		rating := builder.BuildRating(member.Score)
		if rating == 0 {
			rating = u.Rating
		}

		totalOrders := int64(member.Score)
		if _, ok := builder.(*RatingRankingBuilder); ok {
			totalOrders = u.TotalOrders
		}

		item := &user.CompanionRankingItem{
			UserId:      userID,
			Nickname:    u.Nickname,
			AvatarUrl:   u.AvatarURL,
			Rating:      rating,
			TotalOrders: totalOrders,
			Rank:        rank,
			IsVerified:  u.IsVerified,
		}

		rankings = append(rankings, item)
	}

	return &RankingQueryResult{
		Rankings: rankings,
		Total:    int32(total),
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// queryRankingFromMySQL Redis故障时从MySQL查询（降级方案）
func queryRankingFromMySQL(
	ctx context.Context,
	svcCtx *svc.ServiceContext,
	logger logx.Logger,
	rankingType string,
	start, end int32,
	builder RankingItemBuilder,
) (*RankingQueryResult, error) {
	LogInfo(logger, OpGetCompanionRatingRanking, "fallback to mysql for ranking", map[string]interface{}{
		"ranking_type": rankingType,
	})

	// 确定排序字段
	orderBy := "rating"
	if rankingType == "ranking:orders" {
		orderBy = "total_orders"
	}

	db := svcCtx.DB().WithContext(ctx)

	// 从MySQL查询前100名
	type RankingUser struct {
		ID          uint64  `gorm:"column:id"`
		Nickname    string  `gorm:"column:nickname"`
		AvatarURL   string  `gorm:"column:avatar_url"`
		Rating      float64 `gorm:"column:rating"`
		TotalOrders int64   `gorm:"column:total_orders"`
		IsVerified  bool    `gorm:"column:is_verified"`
	}

	var rankingUsers []RankingUser
	err := db.Table("users").
		Select("users.id, users.nickname, users.avatar_url, " +
			"companion_profiles.rating, companion_profiles.total_orders, companion_profiles.is_verified").
		Joins("INNER JOIN companion_profiles ON users.id = companion_profiles.user_id").
		Where("companion_profiles.total_orders > 0"). // 只查有订单的陪玩
		Order(orderBy + " DESC").
		Limit(100).
		Find(&rankingUsers).Error

	if err != nil {
		LogError(logger, OpGetCompanionRatingRanking, "query ranking from mysql failed", err, nil)
		return nil, status.Error(codes.Internal, "query ranking failed")
	}

	total := int32(len(rankingUsers))
	if total > 100 {
		total = 100
	}

	// 分页
	if int(start) >= len(rankingUsers) {
		return &RankingQueryResult{
			Rankings: []*user.CompanionRankingItem{},
			Total:    total,
			Page:     int32(start/int32(10)) + 1,
			PageSize: int32(end - start + 1),
		}, nil
	}

	endIdx := int(end) + 1
	if endIdx > len(rankingUsers) {
		endIdx = len(rankingUsers)
	}

	pageUsers := rankingUsers[start:endIdx]

	// 组装结果
	rankings := make([]*user.CompanionRankingItem, 0, len(pageUsers))
	for i, u := range pageUsers {
		rank := int32(int(start) + i + 1)

		rating := u.Rating
		totalOrders := u.TotalOrders
		if rankingType == "ranking:rating" {
			// 评分榜需要特殊处理rating显示
			rating = u.Rating
		} else if rankingType == "ranking:orders" {
			// 接单榜显示评分从profile获取
			rating = u.Rating
		}

		item := &user.CompanionRankingItem{
			UserId:      u.ID,
			Nickname:    u.Nickname,
			AvatarUrl:   u.AvatarURL,
			Rating:      rating,
			TotalOrders: totalOrders,
			Rank:        rank,
			IsVerified:  u.IsVerified,
		}

		rankings = append(rankings, item)
	}

	return &RankingQueryResult{
		Rankings: rankings,
		Total:    total,
		Page:     int32(start/int32(10)) + 1,
		PageSize: int32(end - start + 1),
	}, nil
}
