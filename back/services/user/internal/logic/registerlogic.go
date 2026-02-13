package logic

import (
	"context"
	"errors"
	"strings"

	"SLGaming/back/services/user/internal/helper"
	"SLGaming/back/services/user/internal/metrics"
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

	// 步骤1：布隆过滤器快速检查手机号是否可能存在
	phoneExists := true // 默认为true，表示需要查数据库确认
	if l.svcCtx.BloomFilter != nil {
		mightExist, err := l.svcCtx.BloomFilter.Phone.MightContain(l.ctx, phone)
		if err != nil {
			l.Logger.Errorf("bloom filter check phone failed: %v", err)
			// 布隆过滤器查询失败，降级到数据库查询（phoneExists保持true）
		} else if !mightExist {
			// 布隆过滤器说"不存在"，那一定不存在，直接继续注册流程
			// 省去了数据库查询，加速了注册流程
			phoneExists = false
			helper.LogInfo(l.Logger, helper.OpRegister, "phone definitely not exists (bloom filter), skip db check", map[string]interface{}{
				"phone": phone,
			})
		} else {
			// 布隆过滤器说"可能存在"，需要查数据库确认（因为布隆过滤器有假阳性）
			helper.LogInfo(l.Logger, helper.OpRegister, "phone might exist in bloom filter, checking database", map[string]interface{}{
				"phone": phone,
			})
		}
	}

	db := l.svcCtx.DB().WithContext(l.ctx)

	// 步骤2：如果布隆过滤器无法确定不存在，则查数据库确认
	if phoneExists {
		var existing model.User
		if err := db.Where("phone = ?", phone).First(&existing).Error; err == nil {
			helper.LogWarning(l.Logger, helper.OpRegister, "phone already registered", map[string]interface{}{
				"phone": phone,
			})
			metrics.UserRegisterTotal.WithLabelValues("duplicate").Inc()
			return nil, status.Error(codes.AlreadyExists, "phone already registered")
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			helper.LogError(l.Logger, helper.OpRegister, "check phone exists failed", err, map[string]interface{}{
				"phone": phone,
			})
			return nil, status.Error(codes.Internal, err.Error())
		}
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
		metrics.UserRegisterTotal.WithLabelValues("error").Inc()
		return nil, status.Error(codes.Internal, err.Error())
	}

	// 将新用户添加到布隆过滤器（ID、手机号、UID）
	if l.svcCtx.BloomFilter != nil {
		// 添加用户ID
		if err := l.svcCtx.BloomFilter.UserID.Add(l.ctx, int64(userModel.ID)); err != nil {
			helper.LogError(l.Logger, helper.OpRegister, "add user id to bloom filter failed", err, map[string]interface{}{
				"user_id": userModel.ID,
			})
		}
		// 添加手机号（关键！这样下次注册时才能快速检查）
		if err := l.svcCtx.BloomFilter.Phone.Add(l.ctx, phone); err != nil {
			helper.LogError(l.Logger, helper.OpRegister, "add phone to bloom filter failed", err, map[string]interface{}{
				"phone": phone,
			})
		}
		// 添加UID
		if err := l.svcCtx.BloomFilter.UID.Add(l.ctx, userModel.UID); err != nil {
			helper.LogError(l.Logger, helper.OpRegister, "add uid to bloom filter failed", err, map[string]interface{}{
				"uid": userModel.UID,
			})
		}
	}

	// 记录成功日志
	helper.LogSuccess(l.Logger, helper.OpRegister, map[string]interface{}{
		"user_id": userModel.ID,
		"uid":     userModel.UID,
		"phone":   phone,
	})

	metrics.UserRegisterTotal.WithLabelValues("success").Inc()

	return &user.RegisterResponse{
		Id:  userModel.ID,
		Uid: userModel.UID,
	}, nil
}
