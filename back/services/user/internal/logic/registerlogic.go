package logic

import (
	"context"
	"errors"

	"SLGaming/back/services/user/internal/model"
	"SLGaming/back/services/user/internal/svc"
	"SLGaming/back/services/user/user"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type RegisterLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RegisterLogic) Register(in *user.RegisterRequest) (*user.RegisterResponse, error) {
	phone := in.GetPhone()
	db := l.svcCtx.DB().WithContext(l.ctx)
	var existing model.User
	if err := db.Where("phone = ?", phone).First(&existing).Error; err == nil {
		return nil, status.Error(codes.AlreadyExists, "phone already registered")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, status.Error(codes.Internal, err.Error())
	}

	hashed, err := hashPassword(in.GetPassword())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	newUser := &model.User{
		Nickname: ensureNickname(in.GetNickname(), phone),
		Phone:    phone,
		Password: hashed,
	}

	if err := db.Create(newUser).Error; err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &user.RegisterResponse{
		Id:  newUser.ID,
		Uid: newUser.UID,
	}, nil
}
