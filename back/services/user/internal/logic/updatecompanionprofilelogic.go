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

type UpdateCompanionProfileLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateCompanionProfileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateCompanionProfileLogic {
	return &UpdateCompanionProfileLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateCompanionProfileLogic) UpdateCompanionProfile(in *user.UpdateCompanionProfileRequest) (*user.UpdateCompanionProfileResponse, error) {
	userID := in.GetUserId()
	if userID == 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
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

	// 查询或创建陪玩信息
	var profile model.CompanionProfile
	if err := db.Where("user_id = ?", userID).First(&profile).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 如果不存在，创建新的
			profile = model.CompanionProfile{
				UserID:       userID,
				GameSkills:   "[]",
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

	// 更新字段
	updates := map[string]any{}

	if gameSkills := strings.TrimSpace(in.GetGameSkills()); gameSkills != "" {
		updates["game_skills"] = gameSkills
	}

	if in.GetPricePerHour() > 0 {
		updates["price_per_hour"] = in.GetPricePerHour()
	}

	if in.GetStatus() >= 0 {
		statusVal := int(in.GetStatus())
		if statusVal != model.CompanionStatusOffline && statusVal != model.CompanionStatusOnline && statusVal != model.CompanionStatusBusy {
			return nil, status.Error(codes.InvalidArgument, "invalid status")
		}
		updates["status"] = statusVal
	}

	if len(updates) > 0 {
		// 使用明确的 WHERE 条件更新，避免 GORM 报错
		if err := db.Model(&model.CompanionProfile{}).Where("user_id = ?", userID).Updates(updates).Error; err != nil {
			l.Errorf("update companion profile failed: user_id=%d, error=%v", userID, err)
			return nil, status.Error(codes.Internal, err.Error())
		}
		// 重新查询获取最新数据
		if err := db.Where("user_id = ?", userID).First(&profile).Error; err != nil {
			l.Errorf("query updated companion profile failed: user_id=%d, error=%v", userID, err)
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &user.UpdateCompanionProfileResponse{
		Profile: helper.ToCompanionInfo(&profile),
	}, nil
}
