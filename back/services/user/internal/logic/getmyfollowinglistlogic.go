package logic

import (
	"context"
	"sync"

	"SLGaming/back/services/user/internal/model"
	"SLGaming/back/services/user/internal/svc"
	"SLGaming/back/services/user/user"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GetMyFollowingListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetMyFollowingListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMyFollowingListLogic {
	return &GetMyFollowingListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetMyFollowingListLogic) GetMyFollowingList(in *user.GetMyFollowingListRequest) (*user.GetMyFollowingListResponse, error) {
	// 1. 验证参数
	if in.OperatorId == 0 {
		return nil, status.Error(codes.InvalidArgument, "operator_id is required")
	}

	// 2. 规范化分页参数
	page := max(int64(in.Page), 1)
	pageSize := clamp(int64(in.PageSize), 1, 100)
	offset := (page - 1) * pageSize

	// 3. 查询关注关系
	db := l.svcCtx.DB().WithContext(l.ctx)

	// 4. 处理搜索关键词
	var matchedUserIds []uint64
	if in.Keyword != "" {
		// 先搜索昵称匹配的用户
		if err := db.Model(&model.User{}).
			Where("nickname LIKE ?", "%"+in.Keyword+"%").
			Pluck("id", &matchedUserIds).Error; err != nil {
			l.Errorf("search users by keyword failed: %v", err)
			return nil, status.Error(codes.Internal, "search users failed")
		}
		// 如果没有匹配的用户，直接返回空结果
		if len(matchedUserIds) == 0 {
			return &user.GetMyFollowingListResponse{
				Users:    []*user.UserFollowInfo{},
				Total:    0,
				Page:     int32(page),
				PageSize: int32(pageSize),
			}, nil
		}
	}

	// 5. 定义查询条件参数
	followerID := in.OperatorId
	hasKeywordFilter := len(matchedUserIds) > 0

	// 6. 从用户表获取关注总数（O(1)查询，避免COUNT性能问题）
	var currentUser model.User
	if err := db.First(&currentUser, followerID).Error; err != nil {
		l.Errorf("get user following count failed: %v", err)
		return nil, status.Error(codes.Internal, "get user info failed")
	}
	total := currentUser.FollowingCount

	// 7. 查询关注列表，按关注时间倒序排序
	var followRelations []model.FollowRelation
	listQuery := db.Model(&model.FollowRelation{}).Where("follower_id = ?", followerID)
	if hasKeywordFilter {
		listQuery = listQuery.Where("following_id IN ?", matchedUserIds)
	}
	if err := listQuery.Offset(int(offset)).Limit(int(pageSize)).Order("followed_at DESC").Find(&followRelations).Error; err != nil {
		l.Errorf("get following list failed: %v", err)
		return nil, status.Error(codes.Internal, "get following list failed")
	}

	// 8. 提取关注者ID列表
	var followingIds []uint64
	for _, relation := range followRelations {
		followingIds = append(followingIds, relation.FollowingID)
	}

	// 9. 批量查询用户信息和互相关注关系（使用新的WaitGroup）
	var users []model.User
	var wg sync.WaitGroup

	if len(followingIds) > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := db.Where("id IN ?", followingIds).Find(&users).Error; err != nil {
				l.Errorf("get users failed: %v", err)
			}
		}()
	}

	// 10. 批量查询互相关注关系
	var mutualFollowMap = make(map[uint64]bool)
	if len(followingIds) > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var mutualRelations []model.FollowRelation
			if err := db.Where("follower_id IN ? AND following_id = ?", followingIds, in.OperatorId).Find(&mutualRelations).Error; err != nil {
				l.Errorf("get mutual follow relations failed: %v", err)
				return
			}
			for _, r := range mutualRelations {
				mutualFollowMap[r.FollowerID] = true
			}
		}()
	}

	wg.Wait()

	// 11. 构建响应
	userMap := make(map[uint64]*model.User)
	for i := range users {
		userMap[users[i].ID] = &users[i]
	}

	var userInfos []*user.UserFollowInfo
	for _, relation := range followRelations {
		u := userMap[relation.FollowingID]
		if u == nil {
			continue
		}

		userInfo := &user.UserFollowInfo{
			UserId:     u.ID,
			Nickname:   u.Nickname,
			AvatarUrl:  u.AvatarURL,
			IsMutual:   mutualFollowMap[relation.FollowingID],
			FollowedAt: relation.FollowedAt.Unix(),
		}

		userInfos = append(userInfos, userInfo)
	}

	l.Infof("get following list: operator=%d, keyword=%s, page=%d, page_size=%d, total=%d, result=%d",
		in.OperatorId, in.Keyword, page, pageSize, total, len(userInfos))

	return &user.GetMyFollowingListResponse{
		Users:    userInfos,
		Total:    int32(total),
		Page:     int32(page),
		PageSize: int32(pageSize),
	}, nil
}
