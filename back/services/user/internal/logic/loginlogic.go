package logic

import (
	"context"
	"errors"
	"strings"

	"SLGaming/back/services/user/internal/helper"
	"SLGaming/back/services/user/internal/model"
	"SLGaming/back/services/user/internal/svc"
	"SLGaming/back/services/user/user"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type LoginLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *LoginLogic) Login(in *user.LoginRequest) (*user.LoginResponse, error) {
	phone := strings.TrimSpace(in.GetPhone())
	password := strings.TrimSpace(in.GetPassword())

	if phone == "" || password == "" {
		helper.LogError(l.Logger, helper.OpLogin, "missing required fields", nil, map[string]interface{}{
			"phone": phone,
		})
		return nil, status.Error(codes.InvalidArgument, "phone and password are required")
	}

	db := l.svcCtx.DB().WithContext(l.ctx)
	var u model.User
	if err := db.Where("phone = ?", phone).First(&u).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			helper.LogWarning(l.Logger, helper.OpLogin, "user not found", map[string]interface{}{
				"phone": phone,
			})
			return nil, status.Error(codes.NotFound, "user not found")
		}
		helper.LogError(l.Logger, helper.OpLogin, "get user failed", err, map[string]interface{}{
			"phone": phone,
		})
		return nil, status.Error(codes.Internal, err.Error())
	}

	if err := helper.VerifyPassword(u.Password, password); err != nil {
		helper.LogWarning(l.Logger, helper.OpLogin, "invalid credentials", map[string]interface{}{
			"phone": phone,
			"user_id": u.ID,
		})
		return nil, status.Error(codes.PermissionDenied, "invalid credentials")
	}

	// 记录成功日志
	helper.LogSuccess(l.Logger, helper.OpLogin, map[string]interface{}{
		"user_id": u.ID,
		"uid":     u.UID,
		"phone":   phone,
	})

	return &user.LoginResponse{
		Id:  u.ID,
		Uid: u.UID,
	}, nil
}
