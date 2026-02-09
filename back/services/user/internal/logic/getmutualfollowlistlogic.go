package logic

import (
	"context"
	"sync"

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

	// 1. 并发查询用户关注的人
	var followingIds []uint64
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		var followingRelations []model.FollowRelation
		if err := l.svcCtx.DB().Where("follower_id = ?", in.OperatorId).Find(&followingRelations).Error; err != nil {
			l.Error("获取关注关系失败: %v", err)
			return
		}
		for _, relation := range followingRelations {
			followingIds = append(followingIds, relation.FollowingID)
		}
	}()

	wg.Wait()

	// 2. 如果用户没有关注任何人，直接返回空列表
	if len(followingIds) == 0 {
		return &user.GetMutualFollowListResponse{
			Users:    []*user.UserFollowInfo{},
			Total:    0,
			Page:     int32(page),
			PageSize: int32(pageSize),
		}, nil
	}

	// 3. 并发查询互相关注关系和总数
	var mutualFollowRelations []model.FollowRelation
	var total int64

	wg.Add(2)

	// 查询互相关注关系，按关注时间倒序排序
	go func() {
		defer wg.Done()
		if err := l.svcCtx.DB().Where("follower_id IN ? AND following_id = ?", followingIds, in.OperatorId).Offset(int(offset)).Limit(int(pageSize)).Order("followed_at DESC").Find(&mutualFollowRelations).Error; err != nil {
			l.Error("获取互相关注关系失败: %v", err)
		}
	}()

	// 查询总记录数
	go func() {
		defer wg.Done()
		if err := l.svcCtx.DB().Model(&model.FollowRelation{}).Where("follower_id IN ? AND following_id = ?", followingIds, in.OperatorId).Count(&total).Error; err != nil {
			l.Error("获取互相关注总数失败: %v", err)
		}
	}()

	wg.Wait()

	// 4. 提取互相关注者ID列表
	var mutualIds []uint64
	for _, relation := range mutualFollowRelations {
		mutualIds = append(mutualIds, relation.FollowerID)
	}

	// 5. 如果没有互相关注的人，直接返回空列表
	if len(mutualIds) == 0 {
		return &user.GetMutualFollowListResponse{
			Users:    []*user.UserFollowInfo{},
			Total:    int32(total),
			Page:     int32(page),
			PageSize: int32(pageSize),
		}, nil
	}

	// 6. 关联查询用户信息（减少N+1查询）
	type MutualWithUser struct {
		FollowRelation model.FollowRelation
		User           model.User
	}

	var mutualUsers []MutualWithUser
	result := l.svcCtx.DB().Table("follow_relations").Select("follow_relations.*, users.*").Joins("JOIN users ON users.id = follow_relations.follower_id").Where("follow_relations.follower_id IN ? AND follow_relations.following_id = ?", mutualIds, in.OperatorId).Scan(&mutualUsers)
	if result.Error != nil {
		l.Error("获取互相关注用户信息失败: %v", result.Error)
		return nil, result.Error
	}

	// 7. 构建用户信息列表
	userInfos := make([]*user.UserFollowInfo, 0, len(mutualUsers))
	for _, u := range mutualUsers {
		userInfos = append(userInfos, &user.UserFollowInfo{
			UserId:     u.User.ID,
			Nickname:   u.User.Nickname,
			AvatarUrl:  u.User.AvatarURL,
			Role:       int32(u.User.Role),
			IsMutual:   true,                               // 互相关注列表中的用户都是互相关注的
			FollowedAt: u.FollowRelation.FollowedAt.Unix(), // 添加关注时间戳
		})
	}

	return &user.GetMutualFollowListResponse{
		Users:    userInfos,
		Total:    int32(total),
		Page:     int32(page),
		PageSize: int32(pageSize),
	}, nil
}
