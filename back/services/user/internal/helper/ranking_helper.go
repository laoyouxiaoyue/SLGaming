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
		ps = 100 // 最大100条
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
// 用于将 Redis 的 score 转换为排名项的具体字段
type RankingItemBuilder interface {
	// BuildRating 构建评分（从 Redis score 计算）
	BuildRating(score int64) float64
	// GetRatingFromProfile 从 profile 获取评分（如果 score 不包含评分信息）
	GetRatingFromProfile(profile *model.CompanionProfile) float64
}

// RatingRankingBuilder 评分排名构建器
type RatingRankingBuilder struct{}

func (b *RatingRankingBuilder) BuildRating(score int64) float64 {
	// 从 score 还原 rating（之前乘以了 10000）
	return float64(score) / 10000.0
}

func (b *RatingRankingBuilder) GetRatingFromProfile(profile *model.CompanionProfile) float64 {
	// 评分排名中，rating 从 score 计算，不使用 profile
	return 0
}

// OrdersRankingBuilder 接单数排名构建器
type OrdersRankingBuilder struct{}

func (b *OrdersRankingBuilder) BuildRating(score int64) float64 {
	// 接单数排名中，score 就是 total_orders，rating 从 profile 获取
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

// QueryRankingWithProfiles 查询排名并关联用户和陪玩信息（使用 JOIN 优化性能）
func QueryRankingWithProfiles(
	ctx context.Context,
	svcCtx *svc.ServiceContext,
	logger logx.Logger,
	redisKey string,
	pagination PaginationParams,
	builder RankingItemBuilder,
) (*RankingQueryResult, error) {
	// 从 Redis ZSet 中查询排名（降序，从高到低）
	members, err := svcCtx.Redis.ZrevrangeWithScores(redisKey, int64(pagination.Start), int64(pagination.End))
	if err != nil {
		LogError(logger, OpGetCompanionRatingRanking, "query ranking from redis failed", err, map[string]interface{}{
			"redis_key": redisKey,
		})
		return nil, status.Error(codes.Internal, "query ranking failed")
	}

	// 获取总数
	total, err := svcCtx.Redis.Zcard(redisKey)
	if err != nil {
		LogError(logger, OpGetCompanionRatingRanking, "get ranking total from redis failed", err, map[string]interface{}{
			"redis_key": redisKey,
		})
		return nil, status.Error(codes.Internal, "query ranking total failed")
	}

	// 如果没有数据，直接返回空列表
	if len(members) == 0 {
		return &RankingQueryResult{
			Rankings: []*user.CompanionRankingItem{},
			Total:    int32(total),
			Page:     int32(pagination.Page),
			PageSize: int32(pagination.PageSize),
		}, nil
	}

	// 收集所有 user_id
	userIDs := make([]uint64, 0, len(members))
	for _, member := range members {
		userID, err := strconv.ParseUint(member.Key, 10, 64)
		if err != nil {
			LogError(logger, OpGetCompanionRatingRanking, "parse user_id from redis member failed", err, map[string]interface{}{
				"member_key": member.Key,
			})
			continue
		}
		userIDs = append(userIDs, userID)
	}

	if len(userIDs) == 0 {
		return &RankingQueryResult{
			Rankings: []*user.CompanionRankingItem{},
			Total:    int32(total),
			Page:     int32(pagination.Page),
			PageSize: int32(pagination.PageSize),
		}, nil
	}

	// 使用 JOIN 优化：一次查询获取用户和陪玩信息
	db := svcCtx.DB().WithContext(ctx)
	
	// 分别查询用户和陪玩信息（避免 JSON 标签冲突）
	var users []model.User
	if err := db.Where("id IN ?", userIDs).Find(&users).Error; err != nil {
		LogError(logger, OpGetCompanionRatingRanking, "query users failed", err, map[string]interface{}{
			"user_count": len(userIDs),
		})
		return nil, status.Error(codes.Internal, "query users failed")
	}

	var profiles []model.CompanionProfile
	if err := db.Where("user_id IN ?", userIDs).Find(&profiles).Error; err != nil {
		LogError(logger, OpGetCompanionRatingRanking, "query companion profiles failed", err, map[string]interface{}{
			"user_count": len(userIDs),
		})
		return nil, status.Error(codes.Internal, "query companion profiles failed")
	}

	// 构建 user_id -> 用户信息的映射
	userMap := make(map[uint64]*model.User)
	for i := range users {
		userMap[users[i].ID] = &users[i]
	}

	// 构建 user_id -> 陪玩信息的映射
	profileMap := make(map[uint64]*model.CompanionProfile)
	for i := range profiles {
		profileMap[profiles[i].UserID] = &profiles[i]
	}

	// 组装排名列表（按照 Redis 返回的顺序）
	rankings := make([]*user.CompanionRankingItem, 0, len(members))
	for i, member := range members {
		userID, err := strconv.ParseUint(member.Key, 10, 64)
		if err != nil {
			LogError(logger, OpGetCompanionRatingRanking, "parse user_id from redis member failed", err, map[string]interface{}{
				"member_key": member.Key,
			})
			continue
		}
		rank := int32(pagination.Start + i + 1)

		u := userMap[userID]
		if u == nil {
			logger.Infof("user not found: user_id=%d", userID)
			continue
		}

		p := profileMap[userID]
		// 如果陪玩信息不存在，跳过（排名只显示陪玩）
		if p == nil {
			logger.Infof("companion profile not found: user_id=%d", userID)
			continue
		}

		// 构建评分：优先从 score 计算，如果 builder 返回 0，则从 profile 获取
		rating := builder.BuildRating(member.Score)
		if rating == 0 {
			rating = builder.GetRatingFromProfile(p)
		}

		// score 作为 total_orders（对于接单数排名）或用于计算 rating（对于评分排名）
		totalOrders := int64(member.Score)
		// 对于评分排名，total_orders 从 profile 获取
		if _, ok := builder.(*RatingRankingBuilder); ok {
			totalOrders = p.TotalOrders
		}

		item := &user.CompanionRankingItem{
			UserId:      userID,
			Nickname:    u.Nickname,
			AvatarUrl:   u.AvatarURL,
			Rating:      rating,
			TotalOrders: totalOrders,
			Rank:        rank,
			IsVerified:  p.IsVerified,
		}

		rankings = append(rankings, item)
	}

	return &RankingQueryResult{
		Rankings: rankings,
		Total:    int32(total),
		Page:     int32(pagination.Page),
		PageSize: int32(pagination.PageSize),
	}, nil
}
