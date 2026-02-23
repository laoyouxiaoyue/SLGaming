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

// FollowerWithUser JOIN 查询结果结构（扁平化，避免 GORM 嵌套映射问题）
type FollowerWithUser struct {
	FollowerID uint64    `gorm:"column:follower_id"`
	FollowedAt time.Time `gorm:"column:followed_at"`
	Nickname   string    `gorm:"column:nickname"`
	AvatarURL  string    `gorm:"column:avatar_url"`
}

func (l *GetMyFollowersListLogic) GetMyFollowersList(in *user.GetMyFollowersListRequest) (*user.GetMyFollowersListResponse, error) {
	// 1. 验证参数
	if in.OperatorId == 0 {
		return nil, status.Error(codes.InvalidArgument, "operator_id is required")
	}

	const maxFollowersLimit = 1000

	// 2. 规范化分页参数
	page := max(int64(in.Page), 1)
	pageSize := clamp(int64(in.PageSize), 1, 100)
	offset := (page - 1) * pageSize

	// 3. 检查是否超出最大限制
	if offset >= maxFollowersLimit {
		return &user.GetMyFollowersListResponse{
			Users:    []*user.UserFollowInfo{},
			Total:    int32(maxFollowersLimit),
			Page:     int32(page),
			PageSize: int32(pageSize),
		}, nil
	}

	// 4. 调整 pageSize，确保不超过限制
	if offset+pageSize > maxFollowersLimit {
		pageSize = maxFollowersLimit - offset
	}

	db := l.svcCtx.DB().WithContext(l.ctx)

	// 5. JOIN 查询粉丝列表和用户信息（合并为 1 次查询）
	var followers []FollowerWithUser
	if err := db.Table("follow_relations").
		Select("follow_relations.follower_id, follow_relations.followed_at, users.nickname, users.avatar_url").
		Joins("JOIN users ON users.id = follow_relations.follower_id").
		Where("follow_relations.following_id = ?", in.OperatorId).
		Offset(int(offset)).
		Limit(int(pageSize)).
		Order("follow_relations.followed_at DESC").
		Scan(&followers).Error; err != nil {
		l.Errorf("get followers list failed: %v", err)
		return nil, status.Error(codes.Internal, "get followers list failed")
	}

	// 6. 提取粉丝 ID 列表
	var followerIds []uint64
	for _, f := range followers {
		followerIds = append(followerIds, f.FollowerID)
	}

	// 7. 批量查询互相关注关系（只取 ID，减少数据传输）
	mutualFollowMap := make(map[uint64]bool)
	if len(followerIds) > 0 {
		var mutualIds []uint64
		if err := db.Model(&model.FollowRelation{}).
			Where("follower_id = ? AND following_id IN ?", in.OperatorId, followerIds).
			Pluck("following_id", &mutualIds).Error; err != nil {
			l.Errorf("get mutual follow relations failed: %v", err)
		}
		for _, id := range mutualIds {
			mutualFollowMap[id] = true
		}
	}

	// 8. 构建响应
	userInfos := make([]*user.UserFollowInfo, 0, len(followers))
	for _, f := range followers {
		userInfos = append(userInfos, &user.UserFollowInfo{
			UserId:     f.FollowerID,
			Nickname:   f.Nickname,
			AvatarUrl:  f.AvatarURL,
			IsMutual:   mutualFollowMap[f.FollowerID],
			FollowedAt: f.FollowedAt.Unix(),
		})
	}

	// 9. 获取总记录数（优先从 Redis 缓存获取，避免慢 COUNT）
	var total int64
	if l.svcCtx.UserCache != nil {
		count, err := l.svcCtx.UserCache.GetFollowerCount(int64(in.OperatorId))
		if err == nil {
			total = count
		} else {
			db.Model(&model.FollowRelation{}).Where("following_id = ?", in.OperatorId).Count(&total)
		}
	} else {
		db.Model(&model.FollowRelation{}).Where("following_id = ?", in.OperatorId).Count(&total)
	}

	// 10. 限制返回的总数不超过最大值
	if total > maxFollowersLimit {
		total = maxFollowersLimit
	}

	l.Infof("get followers list: operator=%d, page=%d, page_size=%d, total=%d, result=%d",
		in.OperatorId, page, pageSize, total, len(userInfos))

	return &user.GetMyFollowersListResponse{
		Users:    userInfos,
		Total:    int32(total),
		Page:     int32(page),
		PageSize: int32(pageSize),
	}, nil
}
