package logic

import (
	"context"
	"errors"

	"SLGaming/back/pkg/snowflake"
	"SLGaming/back/services/user/internal/helper"
	"SLGaming/back/services/user/internal/model"
	"SLGaming/back/services/user/internal/svc"
	"SLGaming/back/services/user/user"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type GetCompanionProfileLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetCompanionProfileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCompanionProfileLogic {
	return &GetCompanionProfileLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetCompanionProfileLogic) GetCompanionProfile(in *user.GetCompanionProfileRequest) (*user.GetCompanionProfileResponse, error) {
	userID := in.GetUserId()
	if userID == 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	// 布隆过滤器快速检查用户是否存在
	// 如果布隆过滤器说"不存在"，那用户一定不存在，直接返回（省去数据库查询）
	if l.svcCtx.BloomFilter != nil {
		exists, err := l.svcCtx.BloomFilter.UserID.MightContain(l.ctx, int64(userID))
		if err != nil {
			l.Errorf("bloom filter check user failed: %v", err)
			// 布隆过滤器查询失败，降级到数据库查询
		} else if !exists {
			// 用户一定不存在，直接返回
			return nil, status.Error(codes.NotFound, "user not found")
		}
		// 如果存在，需要查数据库确认（布隆过滤器有假阳性）
	}

	db := l.svcCtx.DB().WithContext(l.ctx)

	// 先检查用户是否存在且是陪玩角色
	var u model.User
	if err := db.Where("id = ?", userID).First(&u).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	if !u.IsCompanion() {
		return nil, status.Error(codes.FailedPrecondition, "user is not a companion")
	}

	// 查询陪玩信息
	var profile model.CompanionProfile
	if err := db.Where("user_id = ?", userID).First(&profile).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 如果陪玩信息不存在，创建一个默认的
			profile = model.CompanionProfile{
				BaseModel: model.BaseModel{
					ID: uint64(snowflake.GenID()),
				},
				UserID:       userID,
				GameSkills:   "",
				PricePerHour: 0,
				Status:       model.CompanionStatusOffline,
				Rating:       0,
				TotalOrders:  0,
				IsVerified:   false,
			}
			if err := db.Create(&profile).Error; err != nil {
				return nil, status.Error(codes.Internal, err.Error())
			}
		} else {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &user.GetCompanionProfileResponse{
		Profile: helper.ToCompanionInfoWithUser(&profile, &u),
	}, nil
}
