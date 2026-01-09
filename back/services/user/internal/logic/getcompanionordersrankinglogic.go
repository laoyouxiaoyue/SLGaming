package logic

import (
	"context"
	"strconv"

	"SLGaming/back/services/user/internal/model"
	"SLGaming/back/services/user/internal/svc"
	"SLGaming/back/services/user/user"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GetCompanionOrdersRankingLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetCompanionOrdersRankingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCompanionOrdersRankingLogic {
	return &GetCompanionOrdersRankingLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetCompanionOrdersRankingLogic) GetCompanionOrdersRanking(in *user.GetCompanionOrdersRankingRequest) (*user.GetCompanionOrdersRankingResponse, error) {
	if l.svcCtx.Redis == nil {
		return nil, status.Error(codes.FailedPrecondition, "redis not configured")
	}

	// 分页参数
	page := int(in.GetPage())
	if page < 1 {
		page = 1
	}
	pageSize := int(in.GetPageSize())
	if pageSize < 1 {
		pageSize = 10 // 默认每页10条
	}
	if pageSize > 100 {
		pageSize = 100 // 最大100条
	}

	// 计算分页范围
	start := (page - 1) * pageSize
	end := start + pageSize - 1

	// 从 Redis ZSet 中查询排名（降序，从高到低）
	key := "ranking:orders"
	members, err := l.svcCtx.Redis.ZrevrangeWithScores(key, int64(start), int64(end))
	if err != nil {
		l.Errorf("query orders ranking from redis failed: %v", err)
		return nil, status.Error(codes.Internal, "query ranking failed")
	}

	// 获取总数
	total, err := l.svcCtx.Redis.Zcard(key)
	if err != nil {
		l.Errorf("get orders ranking total from redis failed: %v", err)
		return nil, status.Error(codes.Internal, "query ranking total failed")
	}

	// 如果没有数据，直接返回空列表
	if len(members) == 0 {
		return &user.GetCompanionOrdersRankingResponse{
			Rankings: []*user.CompanionRankingItem{},
			Total:    int32(total),
			Page:     int32(page),
			PageSize: int32(pageSize),
		}, nil
	}

	// 收集所有 user_id
	userIDs := make([]uint64, 0, len(members))
	for _, member := range members {
		userID, err := strconv.ParseUint(member.Key, 10, 64)
		if err != nil {
			l.Errorf("parse user_id from redis member failed: key=%s, err=%v", member.Key, err)
			continue
		}
		userIDs = append(userIDs, userID)
	}

	// 从数据库批量查询用户和陪玩信息
	db := l.svcCtx.DB().WithContext(l.ctx)
	var users []model.User
	if err := db.Where("id IN ?", userIDs).Find(&users).Error; err != nil {
		l.Errorf("query users failed: %v", err)
		return nil, status.Error(codes.Internal, "query users failed")
	}

	var profiles []model.CompanionProfile
	if err := db.Where("user_id IN ?", userIDs).Find(&profiles).Error; err != nil {
		l.Errorf("query companion profiles failed: %v", err)
		return nil, status.Error(codes.Internal, "query companion profiles failed")
	}

	// 构建 user_id -> 用户信息的映射
	userMap := make(map[uint64]*model.User)
	for i := range users {
		userMap[users[i].ID] = &users[i]
	}

	// 构建 user_id -> 陪玩信息的映射
	profileMap := make(map[uint64]*model.CompanionProfile)
	for i := range profiles {
		profileMap[profiles[i].UserID] = &profiles[i]
	}

	// 组装排名列表（按照 Redis 返回的顺序）
	rankings := make([]*user.CompanionRankingItem, 0, len(members))
	for i, member := range members {
		userID, _ := strconv.ParseUint(member.Key, 10, 64)
		rank := int32(start + i + 1)

		u := userMap[userID]
		p := profileMap[userID]

		// 如果用户或陪玩信息不存在，跳过
		if u == nil || p == nil {
			l.Infof("user or profile not found: user_id=%d", userID)
			continue
		}

		// score 就是 total_orders
		totalOrders := int64(member.Score)

		item := &user.CompanionRankingItem{
			UserId:      userID,
			Nickname:    u.Nickname,
			AvatarUrl:   u.AvatarURL,
			Rating:      p.Rating,
			TotalOrders: totalOrders,
			Rank:        rank,
			IsVerified:  p.IsVerified,
		}

		rankings = append(rankings, item)
	}

	return &user.GetCompanionOrdersRankingResponse{
		Rankings: rankings,
		Total:    int32(total),
		Page:     int32(page),
		PageSize: int32(pageSize),
	}, nil
}
