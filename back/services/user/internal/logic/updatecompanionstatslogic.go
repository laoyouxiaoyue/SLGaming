package logic

import (
	"context"
	"strconv"

	"SLGaming/back/services/user/internal/helper"
	"SLGaming/back/services/user/internal/model"
	"SLGaming/back/services/user/internal/svc"
	"SLGaming/back/services/user/user"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type UpdateCompanionStatsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateCompanionStatsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateCompanionStatsLogic {
	return &UpdateCompanionStatsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateCompanionStatsLogic) UpdateCompanionStats(in *user.UpdateCompanionStatsRequest) (*user.UpdateCompanionStatsResponse, error) {
	if in.GetUserId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	if in.GetNewRating() < 0 || in.GetNewRating() > 5 {
		return nil, status.Error(codes.InvalidArgument, "new_rating must be between 0 and 5")
	}
	if in.GetDeltaOrders() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "delta_orders must be positive")
	}

	db := l.svcCtx.DB().WithContext(l.ctx)

	var p model.CompanionProfile
	if err := db.Where("user_id = ?", in.GetUserId()).First(&p).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			helper.LogWarning(l.Logger, helper.OpUpdateCompanionStats, "companion profile not found", map[string]interface{}{
				"user_id": in.GetUserId(),
			})
			return nil, status.Error(codes.NotFound, "companion profile not found")
		}
		helper.LogError(l.Logger, helper.OpUpdateCompanionStats, "get companion profile failed", err, map[string]interface{}{
			"user_id": in.GetUserId(),
		})
		return nil, status.Error(codes.Internal, "get companion profile failed")
	}

	oldOrders := p.TotalOrders
	p.TotalOrders += in.GetDeltaOrders()

	// 重新计算加权平均评分： (旧评分*旧单数 + 本次评分*增量) / 新总单数
	if p.TotalOrders > 0 {
		p.Rating = (p.Rating*float64(oldOrders) + in.GetNewRating()*float64(in.GetDeltaOrders())) / float64(p.TotalOrders)
	}

	if err := db.Save(&p).Error; err != nil {
		helper.LogError(l.Logger, helper.OpUpdateCompanionStats, "update companion stats failed", err, map[string]interface{}{
			"user_id": p.UserID,
		})
		return nil, status.Error(codes.Internal, "update companion stats failed")
	}

	// 更新 Redis 排名 ZSet（只维护前100名）
	if l.svcCtx.Redis != nil {
		l.updateRankingZSet(p.UserID, p.Rating, p.TotalOrders)
	}

	// 记录成功日志
	helper.LogSuccess(l.Logger, helper.OpUpdateCompanionStats, map[string]interface{}{
		"user_id":      p.UserID,
		"new_rating":   p.Rating,
		"total_orders": p.TotalOrders,
	})

	return &user.UpdateCompanionStatsResponse{
		Profile: helper.ToCompanionInfo(&p),
	}, nil
}

// updateRankingZSet 更新排行榜ZSet（只维护前100名）
func (l *UpdateCompanionStatsLogic) updateRankingZSet(userID uint64, rating float64, totalOrders int64) {
	userIDStr := strconv.FormatUint(userID, 10)

	// 更新评分排名
	ratingScore := int64(rating * 10000)
	l.updateZSet("ranking:rating", ratingScore, userIDStr)

	// 更新接单数排名
	l.updateZSet("ranking:orders", totalOrders, userIDStr)
}

// updateZSet 更新单个ZSet（只保留前100名）
func (l *UpdateCompanionStatsLogic) updateZSet(key string, score int64, member string) {
	// 1. 先获取当前第100名的分数（门槛）
	lastMembers, err := l.svcCtx.Redis.ZrevrangeWithScores(key, 99, 99)
	if err != nil {
		helper.LogError(l.Logger, helper.OpUpdateCompanionStats, "get ranking threshold failed", err, map[string]interface{}{
			"key": key,
		})
		return
	}

	// 2. 判断是否进入前100
	shouldAdd := false
	if len(lastMembers) == 0 {
		// ZSet为空，直接加入
		shouldAdd = true
	} else if len(lastMembers) < 100 {
		// 不满100人，直接加入
		shouldAdd = true
	} else {
		// 已满100人，比较分数
		if score > lastMembers[0].Score {
			shouldAdd = true
		}
	}

	if shouldAdd {
		// 3. 加入ZSet
		_, err := l.svcCtx.Redis.Zadd(key, score, member)
		if err != nil {
			helper.LogError(l.Logger, helper.OpUpdateCompanionStats, "zadd failed", err, map[string]interface{}{
				"key":    key,
				"score":  score,
				"member": member,
			})
			return
		}

		// 4. 如果超过100人，删除第101名及以后
		count, _ := l.svcCtx.Redis.Zcard(key)
		if count > 100 {
			l.svcCtx.Redis.Zremrangebyrank(key, 0, -(101))
			helper.LogInfo(l.Logger, helper.OpUpdateCompanionStats, "trimmed ranking to top 100", map[string]interface{}{
				"key": key,
			})
		}

		helper.LogInfo(l.Logger, helper.OpUpdateCompanionStats, "entered top 100", map[string]interface{}{
			"key":    key,
			"member": member,
			"score":  score,
		})
	}
}
