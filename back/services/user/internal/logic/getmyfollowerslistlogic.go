package logic

import (
	"context"

	"SLGaming/back/services/user/internal/model"
	"SLGaming/back/services/user/internal/svc"
	"SLGaming/back/services/user/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMyFollowersListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetMyFollowersListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMyFollowersListLogic {
	return &GetMyFollowersListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetMyFollowersListLogic) GetMyFollowersList(in *user.GetMyFollowersListRequest) (*user.GetMyFollowersListResponse, error) {
	const maxFollowersLimit = 1000 // 最多只能查看前1000名粉丝

	// 分页参数处理
	page := max(int64(in.Page), 1)
	pageSize := clamp(int64(in.PageSize), 1, 100)

	// 检查是否超出最大限制
	offset := (page - 1) * pageSize
	if offset >= maxFollowersLimit {
		// 超出限制，返回空列表
		return &user.GetMyFollowersListResponse{
			Users:    []*user.UserFollowInfo{},
			Total:    int32(maxFollowersLimit),
			Page:     int32(page),
			PageSize: int32(pageSize),
		}, nil
	}

	// 调整 pageSize，确保不超过限制
	if offset+pageSize > maxFollowersLimit {
		pageSize = maxFollowersLimit - offset
	}

	// 1. 查询关注关系并关联用户信息（减少N+1查询）
	type FollowerWithUser struct {
		FollowRelation model.FollowRelation
		User           model.User
	}

	var followers []FollowerWithUser
	result := l.svcCtx.DB().Table("follow_relations").Select("follow_relations.*, users.*").Joins("JOIN users ON users.id = follow_relations.follower_id").Where("follow_relations.following_id = ?", in.OperatorId).Offset(int(offset)).Limit(int(pageSize)).Order("follow_relations.followed_at DESC").Scan(&followers)
	if result.Error != nil {
		l.Error("获取粉丝列表失败: %v", result.Error)
		return nil, result.Error
	}

	// 2. 提取粉丝ID列表用于批量检查互相关注
	var followerIds []uint64
	for _, f := range followers {
		followerIds = append(followerIds, f.FollowRelation.FollowerID)
	}

	// 3. 批量查询互相关注关系（避免为每个粉丝单独查询）
	var mutualFollowMap = make(map[uint64]bool)
	if len(followerIds) > 0 {
		var mutualRelations []model.FollowRelation
		l.svcCtx.DB().Where("follower_id = ? AND following_id IN ?", in.OperatorId, followerIds).Find(&mutualRelations)
		for _, r := range mutualRelations {
			mutualFollowMap[r.FollowingID] = true
		}
	}

	// 4. 构建用户信息列表
	userInfos := make([]*user.UserFollowInfo, 0, len(followers))
	for _, f := range followers {
		userInfos = append(userInfos, &user.UserFollowInfo{
			UserId:     f.User.ID,
			Nickname:   f.User.Nickname,
			AvatarUrl:  f.User.AvatarURL,
			IsMutual:   mutualFollowMap[f.User.ID],
			FollowedAt: f.FollowRelation.FollowedAt.Unix(),
		})
	}

	// 5. 获取总记录数（优先从 Redis 缓存获取，避免慢 COUNT）
	var total int64
	if l.svcCtx.UserCache != nil {
		count, err := l.svcCtx.UserCache.GetFollowerCount(int64(in.OperatorId))
		if err == nil {
			total = count
		} else {
			// Redis 未命中，从数据库查询
			l.svcCtx.DB().Model(&model.FollowRelation{}).Where("following_id = ?", in.OperatorId).Count(&total)
		}
	} else {
		l.svcCtx.DB().Model(&model.FollowRelation{}).Where("following_id = ?", in.OperatorId).Count(&total)
	}

	// 限制返回的总数不超过最大值
	if total > maxFollowersLimit {
		total = maxFollowersLimit
	}

	return &user.GetMyFollowersListResponse{
		Users:    userInfos,
		Total:    int32(total),
		Page:     int32(page),
		PageSize: int32(pageSize),
	}, nil
}
