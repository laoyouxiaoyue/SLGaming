package logic

import (
	"context"

	"SLGaming/back/services/user/internal/model"
	"SLGaming/back/services/user/internal/svc"
	"SLGaming/back/services/user/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMutualFollowListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetMutualFollowListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMutualFollowListLogic {
	return &GetMutualFollowListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetMutualFollowListLogic) GetMutualFollowList(in *user.GetMutualFollowListRequest) (*user.GetMutualFollowListResponse, error) {
	// 分页参数处理
	page := max(int64(in.Page), 1)
	pageSize := clamp(int64(in.PageSize), 1, 100)
	offset := (page - 1) * pageSize

	// 1. 查询用户关注的人（只取 ID）
	var followingIds []uint64
	if err := l.svcCtx.DB().Model(&model.FollowRelation{}).
		Where("follower_id = ?", in.OperatorId).
		Pluck("following_id", &followingIds).Error; err != nil {
		l.Error("获取关注关系失败: %v", err)
		return nil, err
	}

	// 2. 如果用户没有关注任何人，直接返回空列表
	if len(followingIds) == 0 {
		return &user.GetMutualFollowListResponse{
			Users:    []*user.UserFollowInfo{},
			Total:    0,
			Page:     int32(page),
			PageSize: int32(pageSize),
		}, nil
	}

	// 3. 查询互相关注总数
	var total int64
	if err := l.svcCtx.DB().Model(&model.FollowRelation{}).
		Where("follower_id IN ? AND following_id = ?", followingIds, in.OperatorId).
		Count(&total).Error; err != nil {
		l.Error("获取互相关注总数失败: %v", err)
	}

	// 4. JOIN 查询互相关注关系和用户信息
	type MutualWithUser struct {
		FollowerID uint64 `gorm:"column:follower_id"`
		FollowedAt int64  `gorm:"column:followed_at"`
		Nickname   string `gorm:"column:nickname"`
		AvatarURL  string `gorm:"column:avatar_url"`
	}

	var mutualUsers []MutualWithUser
	if err := l.svcCtx.DB().Table("follow_relations").
		Select("follow_relations.follower_id, follow_relations.followed_at, users.nickname, users.avatar_url").
		Joins("JOIN users ON users.id = follow_relations.follower_id").
		Where("follow_relations.follower_id IN ? AND follow_relations.following_id = ?", followingIds, in.OperatorId).
		Offset(int(offset)).Limit(int(pageSize)).
		Order("follow_relations.followed_at DESC").
		Scan(&mutualUsers).Error; err != nil {
		l.Error("获取互相关注用户信息失败: %v", err)
		return nil, err
	}

	// 5. 构建用户信息列表
	userInfos := make([]*user.UserFollowInfo, 0, len(mutualUsers))
	for _, u := range mutualUsers {
		userInfos = append(userInfos, &user.UserFollowInfo{
			UserId:     u.FollowerID,
			Nickname:   u.Nickname,
			AvatarUrl:  u.AvatarURL,
			IsMutual:   true,
			FollowedAt: u.FollowedAt,
		})
	}

	return &user.GetMutualFollowListResponse{
		Users:    userInfos,
		Total:    int32(total),
		Page:     int32(page),
		PageSize: int32(pageSize),
	}, nil
}
