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
	statusFilter := int(in.GetStatus())
	if statusFilter < 0 {
		statusFilter = model.CompanionStatusOnline // 默认只返回在线
	}
	query = query.Where("companion_profiles.status = ?", statusFilter)

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
	if len(in.GetGameSkills()) > 0 {
		// 这里简化处理，实际可能需要更复杂的 JSON 查询
		// 如果 game_skills 是 JSON 数组，需要根据数据库类型使用不同的查询方式
		// MySQL 可以使用 JSON_CONTAINS 或 JSON_SEARCH
		for _, skill := range in.GetGameSkills() {
			if skill != "" {
				// 使用 LIKE 简单匹配（实际项目中建议使用 JSON 函数）
				query = query.Where("companion_profiles.game_skills LIKE ?", "%"+skill+"%")
			}
		}
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

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

	// 转换为响应格式
	companions := make([]*user.CompanionInfo, 0, len(profiles))
	for i := range profiles {
		companions = append(companions, helper.ToCompanionInfo(&profiles[i]))
	}

	return &user.GetCompanionListResponse{
		Companions: companions,
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
