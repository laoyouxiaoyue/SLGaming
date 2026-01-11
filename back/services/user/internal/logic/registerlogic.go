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
	phone := strings.TrimSpace(in.GetPhone())
	password := strings.TrimSpace(in.GetPassword())
	nickname := strings.TrimSpace(in.GetNickname())

	if phone == "" {
		return nil, status.Error(codes.InvalidArgument, "phone is required")
	}
	if password == "" {
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}

	db := l.svcCtx.DB().WithContext(l.ctx)

	var existing model.User
	if err := db.Where("phone = ?", phone).First(&existing).Error; err == nil {
		helper.LogWarning(l.Logger, helper.OpRegister, "phone already registered", map[string]interface{}{
			"phone": phone,
		})
		return nil, status.Error(codes.AlreadyExists, "phone already registered")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		helper.LogError(l.Logger, helper.OpRegister, "check phone exists failed", err, map[string]interface{}{
			"phone": phone,
		})
		return nil, status.Error(codes.Internal, err.Error())
	}

	hashed, err := helper.HashPassword(password)
	if err != nil {
		helper.LogError(l.Logger, helper.OpRegister, "hash password failed", err, nil)
		return nil, status.Error(codes.Internal, err.Error())
	}

	// 处理角色，默认为老板
	role := int(in.GetRole())
	if role == 0 {
		role = model.RoleBoss
	}
	// 验证角色值
	if role != model.RoleBoss && role != model.RoleCompanion && role != model.RoleAdmin {
		return nil, status.Error(codes.InvalidArgument, "invalid role")
	}

	userModel := &model.User{
		Phone:    phone,
		Password: hashed,
		Nickname: helper.EnsureNickname(nickname, phone),
		Role:     role,
	}

	if err := db.Create(userModel).Error; err != nil {
		helper.LogError(l.Logger, helper.OpRegister, "create user failed", err, map[string]interface{}{
			"phone": phone,
		})
		return nil, status.Error(codes.Internal, err.Error())
	}

	// 记录成功日志
	helper.LogSuccess(l.Logger, helper.OpRegister, map[string]interface{}{
		"user_id": userModel.ID,
		"uid":     userModel.UID,
		"phone":   phone,
	})

	return &user.RegisterResponse{
		Id:  userModel.ID,
		Uid: userModel.UID,
	}, nil
}
