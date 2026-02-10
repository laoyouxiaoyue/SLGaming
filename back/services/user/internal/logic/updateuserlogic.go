package logic

import (
	"context"
	"errors"
	"strings"

	"SLGaming/back/services/user/internal/bloom"
	"SLGaming/back/services/user/internal/helper"
	"SLGaming/back/services/user/internal/model"
	"SLGaming/back/services/user/internal/svc"
	"SLGaming/back/services/user/user"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type UpdateUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateUserLogic {
	return &UpdateUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateUserLogic) UpdateUser(in *user.UpdateUserRequest) (*user.UpdateUserResponse, error) {
	if in.GetId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	db := l.svcCtx.DB().WithContext(l.ctx)
	var u model.User
	if err := db.Where("id = ?", in.GetId()).First(&u).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	updates := map[string]any{}

	if nickname := strings.TrimSpace(in.GetNickname()); nickname != "" && nickname != u.Nickname {
		updates["nickname"] = nickname
	}

	if phone := strings.TrimSpace(in.GetPhone()); phone != "" && phone != u.Phone {
		var count int64
		if err := db.Model(&model.User{}).Where("phone = ? AND id <> ?", phone, u.ID).Count(&count).Error; err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		if count > 0 {
			return nil, status.Error(codes.AlreadyExists, "phone already used")
		}
		updates["phone"] = phone
	}

	if password := strings.TrimSpace(in.GetPassword()); password != "" {
		hashed, err := helper.HashPassword(password)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		updates["password"] = hashed
	}

	// 更新角色
	if in.GetRole() != 0 {
		role := int(in.GetRole())
		if role != model.RoleBoss && role != model.RoleCompanion && role != model.RoleAdmin {
			return nil, status.Error(codes.InvalidArgument, "invalid role")
		}
		updates["role"] = role
	}

	// 更新头像
	if avatarURL := strings.TrimSpace(in.GetAvatarUrl()); avatarURL != "" {
		updates["avatar_url"] = avatarURL
	}

	// 更新个人简介
	if bio := strings.TrimSpace(in.GetBio()); bio != "" {
		updates["bio"] = bio
	}

	if len(updates) > 0 {
		if err := db.Model(&u).Updates(updates).Error; err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		if err := db.Where("id = ?", u.ID).First(&u).Error; err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		// 清除用户缓存，确保信息立即更新
		if l.svcCtx.Redis != nil {
			cacheKey := bloom.GetUserCacheKey(int64(u.ID))
			if _, err := l.svcCtx.Redis.Del(cacheKey); err != nil {
				l.Logger.Errorf("delete user cache failed: %v", err)
			} else {
				l.Logger.Infof("user cache deleted successfully: %s", cacheKey)
			}
		}
	}

	return &user.UpdateUserResponse{
		User: helper.ToUserInfo(&u),
	}, nil
}
