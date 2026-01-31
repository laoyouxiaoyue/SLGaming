// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"context"

	"SLGaming/back/services/gateway/internal/jwt"
	"SLGaming/back/services/gateway/internal/svc"
	"SLGaming/back/services/gateway/internal/types"
	"SLGaming/back/services/gateway/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type RefreshTokenLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRefreshTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RefreshTokenLogic {
	return &RefreshTokenLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}
func (l *RefreshTokenLogic) RefreshToken(req *types.RefreshTokenRequest) (resp *types.RefreshTokenResponse, err error) {
	if req.RefreshToken == "" {
		return &types.RefreshTokenResponse{
			BaseResp: types.BaseResp{
				Code: 400,
				Msg:  "Refresh Token 不能为空",
			},
		}, nil
	}

	// 验证 Refresh Token
	claims, err := l.svcCtx.JWT.VerifyToken(req.RefreshToken)
	if err != nil {
		if err == jwt.ErrExpiredToken {
			return &types.RefreshTokenResponse{
				BaseResp: types.BaseResp{
					Code: 401,
					Msg:  "Refresh Token 已过期",
				},
			}, nil
		}
		code, msg := utils.HandleError(err, l.Logger, "VerifyToken")
		return &types.RefreshTokenResponse{
			BaseResp: types.BaseResp{
				Code: code,
				Msg:  msg,
			},
		}, nil
	}

	// 验证 Refresh Token 是否在黑名单中（未被撤销）
	if l.svcCtx.TokenStore != nil {
		// 从 claims 中获取 token 过期时间
		tokenExpiration := claims.ExpiresAt.Time
		valid, err := l.svcCtx.TokenStore.VerifyRefreshToken(l.ctx, claims.UserID, req.RefreshToken, tokenExpiration)
		if err != nil {
			l.Errorf("verify refresh token failed: %v", err)
			return &types.RefreshTokenResponse{
				BaseResp: types.BaseResp{
					Code: 500,
					Msg:  "验证 Refresh Token 失败",
				},
			}, nil
		}
		if !valid {
			return &types.RefreshTokenResponse{
				BaseResp: types.BaseResp{
					Code: 401,
					Msg:  "Refresh Token 已被撤销",
				},
			}, nil
		}
	}

	role := claims.Role

	// 只生成新的 Access Token，Refresh Token 保持不变
	accessToken, err := l.svcCtx.JWT.GenerateAccessToken(claims.UserID, role)
	if err != nil {
		code, msg := utils.HandleError(err, l.Logger, "GenerateAccessToken")
		return &types.RefreshTokenResponse{
			BaseResp: types.BaseResp{
				Code: code,
				Msg:  msg,
			},
		}, nil
	}

	// Refresh Token 不刷新，保持原样
	// 计算 Access Token 的过期时间
	expiresIn := int64(l.svcCtx.JWT.GetAccessTokenDuration().Seconds())

	return &types.RefreshTokenResponse{
		BaseResp: types.BaseResp{
			Code: 0,
			Msg:  "success",
		},
		Data: types.LoginData{
			AccessToken:  accessToken,
			RefreshToken: req.RefreshToken, // 保持原 Refresh Token
			ExpiresIn:    expiresIn,
		},
	}, nil
}
