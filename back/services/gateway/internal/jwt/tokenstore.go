package jwt

import (
	"context"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

// TokenStore Token 存储接口（黑名单模式：只存储已撤销的 token）
type TokenStore interface {
	// StoreRefreshToken 不再存储有效 token（黑名单模式不需要）
	// 保留此方法以兼容接口，实际为空操作
	StoreRefreshToken(ctx context.Context, userID uint64, token string, expiration time.Duration) error
	// VerifyRefreshToken 验证 Refresh Token 是否在黑名单中（已被撤销）
	// 返回 true 表示 token 有效（不在黑名单），false 表示已被撤销
	VerifyRefreshToken(ctx context.Context, userID uint64, token string, tokenExpiration time.Time) (bool, error)
	// VerifyAccessToken 验证 Access Token 是否在黑名单中（已被撤销）
	// 返回 true 表示 token 有效（不在黑名单），false 表示已被撤销
	VerifyAccessToken(ctx context.Context, userID uint64, token string, tokenExpiration time.Time) (bool, error)
	// RevokeRefreshToken 撤销 Refresh Token（加入黑名单）
	RevokeRefreshToken(ctx context.Context, userID uint64, token string, remainingTTL time.Duration) error
	// RevokeAccessToken 撤销 Access Token（加入黑名单）
	RevokeAccessToken(ctx context.Context, userID uint64, token string, remainingTTL time.Duration) error
	// RevokeAllUserTokens 撤销用户的所有 Token（设置用户级别黑名单，包括 Access Token 和 Refresh Token）
	RevokeAllUserTokens(ctx context.Context, userID uint64, refreshTokenDuration time.Duration) error
}

// RedisTokenStore Redis Token 存储
type RedisTokenStore struct {
	redis *redis.Redis
}

// NewRedisTokenStore 创建 Redis Token 存储
func NewRedisTokenStore(r *redis.Redis) *RedisTokenStore {
	return &RedisTokenStore{
		redis: r,
	}
}

// StoreRefreshToken 不再存储有效 token（黑名单模式不需要）
// 保留此方法以兼容接口，实际为空操作
func (r *RedisTokenStore) StoreRefreshToken(ctx context.Context, userID uint64, token string, expiration time.Duration) error {
	// 黑名单模式：不存储有效 token，只存储被撤销的 token
	return nil
}

// VerifyRefreshToken 验证 Refresh Token 是否在黑名单中（已被撤销）
// tokenExpiration: token 的过期时间，用于计算黑名单的 TTL
// 返回 true 表示 token 有效（不在黑名单），false 表示已被撤销
func (r *RedisTokenStore) VerifyRefreshToken(ctx context.Context, userID uint64, token string, tokenExpiration time.Time) (bool, error) {
	// 1. 先检查用户级别的黑名单
	userBlacklistKey := r.getUserBlacklistKey(userID)
	exists, err := r.redis.Exists(userBlacklistKey)
	if err != nil {
		return false, err
	}
	if exists {
		// 用户所有 token 都被撤销
		return false, nil
	}

	// 2. 检查具体 token 的黑名单
	tokenBlacklistKey := r.getRefreshTokenBlacklistKey(userID, token)
	exists, err = r.redis.Exists(tokenBlacklistKey)
	if err != nil {
		return false, err
	}
	if exists {
		// token 已被撤销
		return false, nil
	}

	// token 不在黑名单中，有效
	return true, nil
}

// VerifyAccessToken 验证 Access Token 是否在黑名单中（已被撤销）
// tokenExpiration: token 的过期时间，用于计算黑名单的 TTL
// 返回 true 表示 token 有效（不在黑名单），false 表示已被撤销
func (r *RedisTokenStore) VerifyAccessToken(ctx context.Context, userID uint64, token string, tokenExpiration time.Time) (bool, error) {
	// 1. 先检查用户级别的黑名单
	userBlacklistKey := r.getUserBlacklistKey(userID)
	exists, err := r.redis.Exists(userBlacklistKey)
	if err != nil {
		return false, err
	}
	if exists {
		// 用户所有 token 都被撤销
		return false, nil
	}

	// 2. 检查具体 token 的黑名单
	tokenBlacklistKey := r.getAccessTokenBlacklistKey(userID, token)
	exists, err = r.redis.Exists(tokenBlacklistKey)
	if err != nil {
		return false, err
	}
	if exists {
		// token 已被撤销
		return false, nil
	}

	// token 不在黑名单中，有效
	return true, nil
}

// RevokeRefreshToken 撤销 Refresh Token（加入黑名单）
// remainingTTL: token 的剩余有效期，用于设置黑名单的过期时间
func (r *RedisTokenStore) RevokeRefreshToken(ctx context.Context, userID uint64, token string, remainingTTL time.Duration) error {
	tokenBlacklistKey := r.getRefreshTokenBlacklistKey(userID, token)
	// 将 token 加入黑名单，过期时间设置为 token 的剩余有效期
	ttl := int(remainingTTL.Seconds())
	if ttl <= 0 {
		// 如果已经过期，设置一个很短的过期时间（1分钟）
		ttl = 60
	}
	err := r.redis.Setex(tokenBlacklistKey, "1", ttl)
	if err == nil {
		logx.Infof("Revoked refresh token for user %d, TTL: %d seconds", userID, ttl)
	}
	return err
}

// RevokeAccessToken 撤销 Access Token（加入黑名单）
// remainingTTL: token 的剩余有效期，用于设置黑名单的过期时间
func (r *RedisTokenStore) RevokeAccessToken(ctx context.Context, userID uint64, token string, remainingTTL time.Duration) error {
	tokenBlacklistKey := r.getAccessTokenBlacklistKey(userID, token)
	// 将 token 加入黑名单，过期时间设置为 token 的剩余有效期
	ttl := int(remainingTTL.Seconds())
	if ttl <= 0 {
		// 如果已经过期，设置一个很短的过期时间（1分钟）
		ttl = 60
	}
	err := r.redis.Setex(tokenBlacklistKey, "1", ttl)
	if err == nil {
		logx.Infof("Revoked access token for user %d, TTL: %d seconds", userID, ttl)
	}
	return err
}

// RevokeAllUserTokens 撤销用户的所有 Token（设置用户级别黑名单，包括 Access Token 和 Refresh Token）
// refreshTokenDuration: Refresh Token 的完整有效期，用于设置用户黑名单的过期时间
func (r *RedisTokenStore) RevokeAllUserTokens(ctx context.Context, userID uint64, refreshTokenDuration time.Duration) error {
	userBlacklistKey := r.getUserBlacklistKey(userID)
	// 设置用户级别的黑名单标记，过期时间为 Refresh Token 的完整有效期
	// 这样在该时间段内，用户的所有 token（包括 Access Token 和 Refresh Token）都会被拒绝
	ttl := int(refreshTokenDuration.Seconds())
	if ttl <= 0 {
		ttl = int(7 * 24 * time.Hour.Seconds()) // 默认 7 天
	}
	err := r.redis.Setex(userBlacklistKey, "1", ttl)
	if err == nil {
		logx.Infof("Revoked all tokens for user %d, TTL: %d seconds", userID, ttl)
	}
	return err
}

// getRefreshTokenBlacklistKey 获取 Refresh Token 黑名单的 Redis key
func (r *RedisTokenStore) getRefreshTokenBlacklistKey(userID uint64, token string) string {
	return fmt.Sprintf("blacklist:refresh_token:%d:%s", userID, token)
}

// getAccessTokenBlacklistKey 获取 Access Token 黑名单的 Redis key
func (r *RedisTokenStore) getAccessTokenBlacklistKey(userID uint64, token string) string {
	return fmt.Sprintf("blacklist:access_token:%d:%s", userID, token)
}

// getUserBlacklistKey 获取用户级别黑名单的 Redis key
func (r *RedisTokenStore) getUserBlacklistKey(userID uint64) string {
	return fmt.Sprintf("blacklist:user:%d", userID)
}
