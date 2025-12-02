package logic

import (
	"strings"

	"SLGaming/back/services/user/internal/model"
	"SLGaming/back/services/user/user"

	"golang.org/x/crypto/bcrypt"
)

func hashPassword(password string) (string, error) {
	if password == "" {
		password = "123456"
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

func verifyPassword(hashed, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password))
}

func ensureNickname(nickname, phone string) string {
	nickname = strings.TrimSpace(nickname)
	if nickname != "" {
		return nickname
	}
	if len(phone) >= 4 {
		return "用户" + phone[len(phone)-4:]
	}
	return "新用户"
}

func toUserInfo(u *model.User) *user.UserInfo {
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

func toCompanionInfo(p *model.CompanionProfile) *user.CompanionInfo {
	if p == nil {
		return nil
	}
	return &user.CompanionInfo{
		UserId:       p.UserID,
		GameSkills:   p.GameSkills,
		PricePerHour: p.PricePerHour,
		Status:       int32(p.Status),
		Rating:       p.Rating,
		TotalOrders:  p.TotalOrders,
		IsVerified:   p.IsVerified,
	}
}
