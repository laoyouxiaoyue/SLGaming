package middleware

import (
	"context"
	"fmt"
)

// userIDKey context 中存储用户 ID 的 key
type ctxKey string

const (
	userIDKey       ctxKey = "user_id"
	accessTokenKey  ctxKey = "access_token"
	refreshTokenKey ctxKey = "refresh_token"
)

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

// SetAccessToken 将 Access Token 设置到 context 中
func SetAccessToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, accessTokenKey, token)
}

// GetAccessToken 从 context 中获取 Access Token
func GetAccessToken(ctx context.Context) (string, error) {
	val := ctx.Value(accessTokenKey)
	if val == nil {
		return "", fmt.Errorf("access token not found in context")
	}

	token, ok := val.(string)
	if !ok {
		return "", fmt.Errorf("invalid access token type in context")
	}

	return token, nil
}

// SetRefreshToken 将 Refresh Token 设置到 context 中
func SetRefreshToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, refreshTokenKey, token)
}

// GetRefreshToken 从 context 中获取 Refresh Token
func GetRefreshToken(ctx context.Context) (string, error) {
	val := ctx.Value(refreshTokenKey)
	if val == nil {
		return "", fmt.Errorf("refresh token not found in context")
	}

	token, ok := val.(string)
	if !ok {
		return "", fmt.Errorf("invalid refresh token type in context")
	}

	return token, nil
}
