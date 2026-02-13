package logic

import (
	"context"
	"errors"
	"strings"
	"time"

	"SLGaming/back/services/user/internal/metrics"
	"SLGaming/back/services/user/internal/model"
	"SLGaming/back/services/user/internal/svc"
	"SLGaming/back/services/user/user"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type LoginByCodeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLoginByCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginByCodeLogic {
	return &LoginByCodeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *LoginByCodeLogic) LoginByCode(in *user.LoginByCodeRequest) (*user.LoginByCodeResponse, error) {
	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime).Seconds()
		metrics.UserLoginDuration.WithLabelValues("code").Observe(duration)
	}()

	phone := strings.TrimSpace(in.GetPhone())
	if phone == "" {
		return nil, status.Error(codes.InvalidArgument, "phone is required")
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
			l.Logger.Info("[LoginByCode] user not found (bloom filter), phone: " + phone)
			return nil, status.Error(codes.NotFound, "user not found")
		}
		// 如果存在，需要查数据库确认（布隆过滤器有假阳性）
	}

	db := l.svcCtx.DB().WithContext(l.ctx)
	var u model.User
	if err := db.Where("phone = ?", phone).First(&u).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			metrics.UserLoginTotal.WithLabelValues("not_found", "code").Inc()
			return nil, status.Error(codes.NotFound, "user not found")
		}
		metrics.UserLoginTotal.WithLabelValues("error", "code").Inc()
		return nil, status.Error(codes.Internal, err.Error())
	}

	metrics.UserLoginTotal.WithLabelValues("success", "code").Inc()

	return &user.LoginByCodeResponse{
		Id:  u.ID,
		Uid: u.UID,
	}, nil
}
