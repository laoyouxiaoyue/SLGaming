package middleware

import (
	"context"
	"fmt"
)

// userIDKey context 中存储用户 ID 的 key
type ctxKey string

const userIDKey ctxKey = "user_id"

// SetUserID 将用户 ID 设置到 context 中
func SetUserID(ctx context.Context, userID uint64) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// GetUserID 从 context 中获取用户 ID
func GetUserID(ctx context.Context) (uint64, error) {
	val := ctx.Value(userIDKey)
	if val == nil {
		return 0, fmt.Errorf("user id not found in context")
	}

	userID, ok := val.(uint64)
	if !ok {
		return 0, fmt.Errorf("invalid user id type in context")
	}

	return userID, nil
}

// MustGetUserID 从 context 中获取用户 ID，如果不存在则 panic
func MustGetUserID(ctx context.Context) uint64 {
	userID, err := GetUserID(ctx)
	if err != nil {
		panic(fmt.Sprintf("must get user id from context: %v", err))
	}
	return userID
}
