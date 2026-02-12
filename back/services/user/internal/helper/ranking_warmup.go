package helper

import (
	"context"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"SLGaming/back/services/user/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

// 全局预热状态（使用 atomic 保证并发安全）
var (
	warmupRunning   int32 = 0 // 0=false, 1=true
	warmupDone      int32 = 0 // 0=false, 1=true
	warmupStartTime time.Time
	warmupMu        sync.RWMutex
	warmupError     error
)

// IsWarmupRunning 检查是否正在预热
func IsWarmupRunning() bool {
	return atomic.LoadInt32(&warmupRunning) == 1
}

// IsWarmupDone 检查预热是否完成
func IsWarmupDone() bool {
	return atomic.LoadInt32(&warmupDone) == 1
}

// GetWarmupStatus 获取预热状态详情
func GetWarmupStatus() (running bool, done bool, duration time.Duration, err error) {
	running = IsWarmupRunning()
	done = IsWarmupDone()

	warmupMu.RLock()
	err = warmupError
	warmupMu.RUnlock()

	if running {
		duration = time.Since(warmupStartTime)
	}

	return
}

// WarmupRankingFromMySQLAsync 异步从MySQL加载排行榜数据到Redis
func WarmupRankingFromMySQLAsync(svcCtx *svc.ServiceContext, logger logx.Logger) {
	if svcCtx.Redis == nil {
		logger.Info("redis not configured, skip ranking warmup")
		return
	}

	// 检查是否已经在预热中
	if !atomic.CompareAndSwapInt32(&warmupRunning, 0, 1) {
		logger.Info("ranking warmup already running, skip")
		return
	}

	warmupStartTime = time.Now()
	logger.Info("starting ranking warmup from mysql (async)...")

	// 启动 goroutine 异步执行预热
	go func() {
		defer func() {
			atomic.StoreInt32(&warmupRunning, 0)
			atomic.StoreInt32(&warmupDone, 1)

			duration := time.Since(warmupStartTime)
			if IsWarmupDone() && warmupError == nil {
				logger.Infof("ranking warmup completed in %v", duration)
			} else {
				logger.Errorf("ranking warmup failed after %v: %v", duration, warmupError)
			}
		}()

		ctx := context.Background()

		// 预热评分榜
		if err := warmupRanking(ctx, svcCtx, logger, "ranking:rating", "rating"); err != nil {
			warmupMu.Lock()
			warmupError = err
			warmupMu.Unlock()
			LogError(logger, "warmup", "warmup rating ranking failed", err, nil)
			return
		}

		// 预热接单榜
		if err := warmupRanking(ctx, svcCtx, logger, "ranking:orders", "total_orders"); err != nil {
			warmupMu.Lock()
			warmupError = err
			warmupMu.Unlock()
			LogError(logger, "warmup", "warmup orders ranking failed", err, nil)
			return
		}
	}()
}

// warmupRanking 预热单个排行榜
func warmupRanking(ctx context.Context, svcCtx *svc.ServiceContext, logger logx.Logger, redisKey, orderBy string) error {
	type RankingUser struct {
		UserID      uint64  `gorm:"column:user_id"`
		Rating      float64 `gorm:"column:rating"`
		TotalOrders int64   `gorm:"column:total_orders"`
	}

	db := svcCtx.DB().WithContext(ctx)
	var users []RankingUser

	// 从MySQL查询前100名
	err := db.Table("companion_profiles").
		Select("user_id, rating, total_orders").
		Where("total_orders > 0").
		Order(orderBy + " DESC").
		Limit(100).
		Find(&users).Error

	if err != nil {
		return err
	}

	if len(users) == 0 {
		logger.Infof("no data to warmup for %s", redisKey)
		return nil
	}

	// 清空现有ZSet
	svcCtx.Redis.Del(redisKey)

	// 批量添加到ZSet（使用 Zadds 一次性添加100个）
	pairs := make([]redis.Pair, 0, len(users))
	for _, u := range users {
		userIDStr := strconv.FormatUint(u.UserID, 10)
		var score int64

		if redisKey == "ranking:rating" {
			score = int64(u.Rating * 10000)
		} else {
			score = u.TotalOrders
		}

		pairs = append(pairs, redis.Pair{
			Score: score,
			Key:   userIDStr,
		})
	}

	// 一次性批量添加
	if _, err := svcCtx.Redis.Zadds(redisKey, pairs...); err != nil {
		LogError(logger, "warmup", "zadds failed", err, map[string]interface{}{
			"key":        redisKey,
			"user_count": len(users),
		})
		return err
	}

	LogInfo(logger, "warmup", "ranking warmed up", map[string]interface{}{
		"key":   redisKey,
		"count": len(users),
	})

	return nil
}
