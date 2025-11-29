package logic

import (
	"context"
	"errors"
	"strings"

	"SLGaming/back/services/user/internal/model"
	"SLGaming/back/services/user/internal/svc"
	"SLGaming/back/services/user/user"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type ForgetPasswordLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewForgetPasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ForgetPasswordLogic {
	return &ForgetPasswordLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ForgetPasswordLogic) ForgetPassword(in *user.ForgetPasswordRequest) (*user.ForgetPasswordResponse, error) {
	phone := strings.TrimSpace(in.GetPhone())
	newPassword := strings.TrimSpace(in.GetPassword())

	if phone == "" || newPassword == "" {
		return nil, status.Error(codes.InvalidArgument, "phone and password are required")
	}

	db := l.svcCtx.DB().WithContext(l.ctx)
	var u model.User
	if err := db.Where("phone = ?", phone).First(&u).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	hashed, err := hashPassword(newPassword)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if err := db.Model(&u).Update("password", hashed).Error; err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &user.ForgetPasswordResponse{
		Id:  u.ID,
		Uid: u.UID,
	}, nil
}
