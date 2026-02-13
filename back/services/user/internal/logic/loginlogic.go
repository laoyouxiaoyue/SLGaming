package logic

import (
	"context"
	"errors"
	"strings"
	"time"

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
	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime).Seconds()
		metrics.UserLoginDuration.WithLabelValues("password").Observe(duration)
	}()

	phone := strings.TrimSpace(in.GetPhone())
	password := strings.TrimSpace(in.GetPassword())

	if phone == "" || password == "" {
		helper.LogError(l.Logger, helper.OpLogin, "missing required fields", nil, map[string]interface{}{
			"phone": phone,
		})
		return nil, status.Error(codes.InvalidArgument, "phone and password are required")
	}

	// 步骤1：布隆过滤器快速检查手机号是否存在
	// 如果布隆过滤器说"不存在"，那手机号一定不存在，直接返回（省去数据库查询）
	if l.svcCtx.BloomFilter != nil {
		exists, err := l.svcCtx.BloomFilter.Phone.MightContain(l.ctx, phone)
		if err != nil {
			l.Logger.Errorf("bloom filter check phone failed: %v", err)
			// 布隆过滤器查询失败，降级到数据库查询
		} else if !exists {
			// 手机号肯定不存在，直接返回
			helper.LogWarning(l.Logger, helper.OpLogin, "user not found (bloom filter)", map[string]interface{}{
				"phone": phone,
			})
			return nil, status.Error(codes.NotFound, "user not found")
		}
		// 如果存在，需要查数据库确认（布隆过滤器有假阳性）
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
			"phone":   phone,
			"user_id": u.ID,
		})
		metrics.UserLoginTotal.WithLabelValues("error", "password").Inc()
		return nil, status.Error(codes.PermissionDenied, "invalid credentials")
	}

	// 记录成功日志
	helper.LogSuccess(l.Logger, helper.OpLogin, map[string]interface{}{
		"user_id": u.ID,
		"uid":     u.UID,
		"phone":   phone,
	})

	metrics.UserLoginTotal.WithLabelValues("success", "password").Inc()

	return &user.LoginResponse{
		Id:  u.ID,
		Uid: u.UID,
	}, nil
}
