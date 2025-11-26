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
	phone := in.GetPhone()
	code := strings.TrimSpace(in.GetCode())
	_ = code // TODO: integrate SMS verification

	var u model.User
	db := l.svcCtx.DB().WithContext(l.ctx)
	if err := db.Where("phone = ?", phone).First(&u).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	hashed, err := hashPassword(in.GetPassword())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if err := db.Model(&u).Update("password", hashed).Error; err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// TODO: integrate real SMS code verification service.

	return &user.ForgetPasswordResponse{
		AccessToken: generateToken(u.ID, u.UID),
	}, nil
}
