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

	// 步骤1：布隆过滤器快速检查手机号是否存在
	// 如果布隆过滤器说"不存在"，那手机号一定不存在，直接返回（省去数据库查询）
	if l.svcCtx.BloomFilter != nil {
		exists, err := l.svcCtx.BloomFilter.Phone.MightContain(l.ctx, phone)
		if err != nil {
			l.Logger.Errorf("bloom filter check phone failed: %v", err)
			// 布隆过滤器查询失败，降级到数据库查询
		} else if !exists {
			// 手机号肯定不存在，直接返回
			l.Logger.Info("[ForgetPassword] user not found (bloom filter), phone: " + phone)
			return nil, status.Error(codes.NotFound, "user not found")
		}
		// 如果存在，需要查数据库确认（布隆过滤器有假阳性）
	}

	db := l.svcCtx.DB().WithContext(l.ctx)
	var u model.User
	if err := db.Where("phone = ?", phone).First(&u).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	hashed, err := helper.HashPassword(newPassword)
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
