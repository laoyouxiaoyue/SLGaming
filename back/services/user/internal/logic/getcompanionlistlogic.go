package logic

import (
	"context"
	"encoding/json"
	"strings"

	"SLGaming/back/services/user/internal/helper"
	"SLGaming/back/services/user/internal/model"
	"SLGaming/back/services/user/internal/svc"
	"SLGaming/back/services/user/user"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GetCompanionListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetCompanionListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCompanionListLogic {
	return &GetCompanionListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetCompanionListLogic) GetCompanionList(in *user.GetCompanionListRequest) (*user.GetCompanionListResponse, error) {
	db := l.svcCtx.DB().WithContext(l.ctx)

	// 构建查询
	query := db.Model(&model.CompanionProfile{}).
		Joins("JOIN users ON companion_profiles.user_id = users.id").
		Where("users.role = ?", model.RoleCompanion).
		Where("users.deleted_at IS NULL")

	// 状态筛选（默认只返回在线）
	// 如果 status <= 0，表示未指定或无效值，使用默认值（在线）
	statusFilter := int(in.GetStatus())
	if statusFilter <= 0 {
		statusFilter = model.CompanionStatusOnline // 默认只返回在线
	}
	query = query.Where("companion_profiles.status = ?", statusFilter)
	l.Infof("[GetCompanionList] status filter: %d", statusFilter)

	// 价格筛选
	if in.GetMinPrice() > 0 {
		query = query.Where("companion_profiles.price_per_hour >= ?", in.GetMinPrice())
	}
	if in.GetMaxPrice() > 0 {
		query = query.Where("companion_profiles.price_per_hour <= ?", in.GetMaxPrice())
	}

	// 认证筛选
	if in.GetIsVerified() {
		query = query.Where("companion_profiles.is_verified = ?", true)
	}

	// 游戏技能筛选
	if gameSkill := strings.TrimSpace(in.GetGameSkill()); gameSkill != "" {
		l.Infof("[GetCompanionList] game skill filter: %s", gameSkill)
		// 匹配单个游戏技能
		query = query.Where("companion_profiles.game_skills = ?", gameSkill)
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		l.Errorf("[GetCompanionList] count failed: %v", err)
		return nil, status.Error(codes.Internal, err.Error())
	}
	l.Infof("[GetCompanionList] total count: %d", total)

	// 规范化分页参数（默认每页20条）
	pagination := helper.NormalizePaginationWithDefault(in.GetPage(), in.GetPageSize(), 20)
	offset := (pagination.Page - 1) * pagination.PageSize

	// 查询列表
	var profiles []model.CompanionProfile
	if err := query.Select("companion_profiles.*").
		Order("companion_profiles.rating DESC, companion_profiles.total_orders DESC").
		Offset(offset).
		Limit(pagination.PageSize).
		Find(&profiles).Error; err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// 收集所有用户ID，批量查询用户的 bio 和 avatar_url
	if len(profiles) > 0 {
		userIDs := make([]uint64, 0, len(profiles))
		for i := range profiles {
			userIDs = append(userIDs, profiles[i].UserID)
		}

		var users []model.User
		if err := db.Select("id, nickname, avatar_url, bio").Where("id IN ?", userIDs).Find(&users).Error; err != nil {
			l.Errorf("[GetCompanionList] query users failed: %v", err)
			return nil, status.Error(codes.Internal, err.Error())
		}

		// 创建 user_id -> User 的映射
		userMap := make(map[uint64]*model.User)
		for i := range users {
			userMap[users[i].ID] = &users[i]
		}

		// 转换为响应格式
		companions := make([]*user.CompanionInfo, 0, len(profiles))
		for i := range profiles {
			u := userMap[profiles[i].UserID]
			companions = append(companions, helper.ToCompanionInfoWithUser(&profiles[i], u))
		}

		return &user.GetCompanionListResponse{
			Companions: companions,
			Total:      int32(total),
			Page:       int32(pagination.Page),
			PageSize:   int32(pagination.PageSize),
		}, nil
	}

	// 如果没有数据，返回空列表
	return &user.GetCompanionListResponse{
		Companions: []*user.CompanionInfo{},
		Total:      int32(total),
		Page:       int32(pagination.Page),
		PageSize:   int32(pagination.PageSize),
	}, nil
}

// 辅助函数：检查游戏技能是否匹配（用于更精确的筛选）
func matchGameSkills(gameSkillsJSON string, targetSkills []string) bool {
	if len(targetSkills) == 0 {
		return true
	}

	var skills []string
	if err := json.Unmarshal([]byte(gameSkillsJSON), &skills); err != nil {
		return false
	}

	skillMap := make(map[string]bool)
	for _, skill := range skills {
		skillMap[strings.ToLower(skill)] = true
	}

	for _, target := range targetSkills {
		if skillMap[strings.ToLower(target)] {
			return true
		}
	}

	return false
}
