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

type ChangePasswordLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewChangePasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ChangePasswordLogic {
	return &ChangePasswordLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ChangePasswordLogic) ChangePassword(in *user.ChangePasswordRequest) (*user.ChangePasswordResponse, error) {
	userID := in.GetUserId()
	oldPhone := strings.TrimSpace(in.GetOldPhone())
	newPassword := strings.TrimSpace(in.GetNewPassword())
	if userID == 0 || oldPhone == "" || newPassword == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id, old_phone and new_password are required")
	}

	db := l.svcCtx.DB().WithContext(l.ctx)
	var u model.User
	if err := db.Where("id = ?", userID).First(&u).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	if strings.TrimSpace(u.Phone) != oldPhone {
		return nil, status.Error(codes.InvalidArgument, "old_phone mismatch")
	}

	hashed, err := helper.HashPassword(newPassword)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if err := db.Model(&u).Update("password", hashed).Error; err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &user.ChangePasswordResponse{Success: true}, nil
}
