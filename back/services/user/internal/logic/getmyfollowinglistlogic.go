package logic

import (
	"context"
	"time"

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

// FollowingWithUser JOIN 查询结果结构
type FollowingWithUser struct {
	FollowingID uint64    `gorm:"column:following_id"`
	FollowedAt  time.Time `gorm:"column:followed_at"`
	Nickname    string    `gorm:"column:nickname"`
	AvatarURL   string    `gorm:"column:avatar_url"`
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
		if err := db.Model(&model.User{}).
			Where("nickname LIKE ?", "%"+in.Keyword+"%").
			Pluck("id", &matchedUserIds).Error; err != nil {
			l.Errorf("search users by keyword failed: %v", err)
			return nil, status.Error(codes.Internal, "search users failed")
		}
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

	// 6. 获取关注总数（优先从 Redis 缓存获取）
	var total int64
	if l.svcCtx.UserCache != nil {
		count, err := l.svcCtx.UserCache.GetFollowingCount(int64(followerID))
		if err == nil {
			total = count
		} else {
			var currentUser model.User
			if err := db.Select("following_count").First(&currentUser, followerID).Error; err != nil {
				l.Errorf("get user following count failed: %v", err)
				return nil, status.Error(codes.Internal, "get user info failed")
			}
			total = currentUser.FollowingCount
		}
	} else {
		var currentUser model.User
		if err := db.Select("following_count").First(&currentUser, followerID).Error; err != nil {
			l.Errorf("get user following count failed: %v", err)
			return nil, status.Error(codes.Internal, "get user info failed")
		}
		total = currentUser.FollowingCount
	}

	// 7. JOIN 查询关注列表和用户信息（合并为 1 次查询）
	var followingWithUsers []FollowingWithUser
	query := db.Table("follow_relations").
		Select("follow_relations.following_id, follow_relations.followed_at, users.nickname, users.avatar_url").
		Joins("JOIN users ON users.id = follow_relations.following_id").
		Where("follow_relations.follower_id = ?", followerID)
	if hasKeywordFilter {
		query = query.Where("follow_relations.following_id IN ?", matchedUserIds)
	}
	if err := query.Offset(int(offset)).Limit(int(pageSize)).Order("follow_relations.followed_at DESC").Scan(&followingWithUsers).Error; err != nil {
		l.Errorf("get following list failed: %v", err)
		return nil, status.Error(codes.Internal, "get following list failed")
	}

	// 8. 提取关注者 ID 列表
	var followingIds []uint64
	for _, f := range followingWithUsers {
		followingIds = append(followingIds, f.FollowingID)
	}

	// 9. 查询互相关注关系（只取 ID，减少数据传输）
	var mutualFollowMap = make(map[uint64]bool)
	if len(followingIds) > 0 {
		var mutualIds []uint64
		if err := db.Model(&model.FollowRelation{}).
			Where("follower_id IN ? AND following_id = ?", followingIds, in.OperatorId).
			Pluck("follower_id", &mutualIds).Error; err != nil {
			l.Errorf("get mutual follow relations failed: %v", err)
		}
		for _, id := range mutualIds {
			mutualFollowMap[id] = true
		}
	}

	// 10. 构建响应
	var userInfos []*user.UserFollowInfo
	for _, f := range followingWithUsers {
		userInfo := &user.UserFollowInfo{
			UserId:     f.FollowingID,
			Nickname:   f.Nickname,
			AvatarUrl:  f.AvatarURL,
			IsMutual:   mutualFollowMap[f.FollowingID],
			FollowedAt: f.FollowedAt.Unix(),
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
