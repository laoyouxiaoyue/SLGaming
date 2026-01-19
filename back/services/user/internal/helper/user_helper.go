package helper

import (
	"strings"

	"SLGaming/back/services/user/internal/model"
	"SLGaming/back/services/user/user"

	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	if password == "" {
		password = "123456"
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

func VerifyPassword(hashed, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password))
}

func EnsureNickname(nickname, phone string) string {
	nickname = strings.TrimSpace(nickname)
	if nickname != "" {
		return nickname
	}
	if len(phone) >= 4 {
		return "用户" + phone[len(phone)-4:]
	}
	return "新用户"
}

func ToUserInfo(u *model.User) *user.UserInfo {
	if u == nil {
		return nil
	}
	return &user.UserInfo{
		Id:        u.ID,
		Uid:       u.UID,
		Nickname:  u.Nickname,
		Phone:     u.Phone,
		Role:      int32(u.Role),
		AvatarUrl: u.AvatarURL,
		Bio:       u.Bio,
	}
}

func ToCompanionInfo(p *model.CompanionProfile) *user.CompanionInfo {
	return ToCompanionInfoWithUser(p, nil)
}

// ToCompanionInfoWithUser 将 CompanionProfile 和 User 转换为 CompanionInfo
// 如果 u 为 nil，则只填充陪玩信息，不填充 bio 和 avatar_url
func ToCompanionInfoWithUser(p *model.CompanionProfile, u *model.User) *user.CompanionInfo {
	if p == nil {
		return nil
	}
	info := &user.CompanionInfo{
		UserId:       p.UserID,
		GameSkill:    p.GameSkills,
		PricePerHour: p.PricePerHour,
		Status:       int32(p.Status),
		Rating:       p.Rating,
		TotalOrders:  p.TotalOrders,
		IsVerified:   p.IsVerified,
	}
	// 如果提供了用户信息，填充 bio 和 avatar_url
	if u != nil {
		info.AvatarUrl = u.AvatarURL
		info.Bio = u.Bio
	}
	return info
}
