package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	"SLGaming/back/services/code/codeclient"
	"SLGaming/back/services/gateway/internal/svc"
	"SLGaming/back/services/gateway/internal/types"
	"SLGaming/back/services/user/userclient"

	"github.com/golang-jwt/jwt/v4"
)

var (
	errUserRPCUnavailable = errors.New("user rpc client is not initialized")
	errCodeRPCUnavailable = errors.New("code rpc client is not initialized")
)

type jwtClaims struct {
	UserID uint64 `json:"user_id"`
	jwt.RegisteredClaims
}

func getRPC(key string, svcCtx *svc.ServiceContext) (interface{}, error) {
	if svcCtx == nil {
		return nil, errors.New("service context is nil")
	}

	clients := map[string]interface{}{
		"user": svcCtx.UserRPC,
		"code": svcCtx.CodeRPC,
	}

	client, ok := clients[key]
	if !ok {
		return nil, fmt.Errorf("rpc client not registered for key: %s", key)
	}
	if client == nil {
		switch key {
		case "user":
			return nil, errUserRPCUnavailable
		case "code":
			return nil, errCodeRPCUnavailable
		default:
			return nil, fmt.Errorf("rpc client %s is nil", key)
		}
	}
	return client, nil
}

func successResp() types.BaseResp {
	return types.BaseResp{
		Code: 0,
		Msg:  "OK",
	}
}

func toUserInfo(info *userclient.UserInfo) types.UserInfo {
	if info == nil {
		return types.UserInfo{}
	}
	return types.UserInfo{
		Id:       info.GetId(),
		Uid:      info.GetUid(),
		Nickname: info.GetNickname(),
		Phone:    info.GetPhone(),
	}
}

func verifyCode(ctx context.Context, svcCtx *svc.ServiceContext, phone, code, purpose string) error {
	if code == "" {
		return errors.New("verification code is required")
	}
	rpc, err := getRPC("code", svcCtx)
	if err != nil {
		return err
	}
	codeRPC := rpc.(codeclient.Code)
	resp, err := codeRPC.VerifyCode(ctx, &codeclient.VerifyCodeRequest{
		Phone:   phone,
		Purpose: purpose,
		Code:    code,
	})
	if err != nil {
		return err
	}
	if !resp.Passed {
		return fmt.Errorf("verification code is invalid or expired")
	}
	return nil
}

func generateAccessToken(ctx context.Context, svcCtx *svc.ServiceContext, id uint64) (string, error) {
	cfg := svcCtx.Config.JWT
	secret := cfg.Secret
	if secret == "" {
		secret = "slgaming-gateway-secret"
	}

	ttl := cfg.TTL
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}

	now := time.Now()
	claims := jwtClaims{
		UserID: id,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			Subject:   fmt.Sprintf("%d", id),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
