package logic

import (
	"fmt"
	"strings"

	"SLGaming/back/services/user/internal/model"
	"SLGaming/back/services/user/user"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const defaultSmsCodeLength = 6

func ensureNickname(nickname, phone string) string {
	name := strings.TrimSpace(nickname)
	if name != "" {
		return name
	}

	num := phone
	if len(num) > 4 {
		num = num[len(num)-4:]
	}
	return fmt.Sprintf("用户%s", num)
}

func hashPassword(password string) (string, error) {
	if strings.TrimSpace(password) == "" {
		return "", fmt.Errorf("password is empty")
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

func verifyPassword(hashed, password string) bool {
	if hashed == "" || strings.TrimSpace(password) == "" {
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password)) == nil
}

func userToProto(u *model.User) *user.UserInfo {
	if u == nil {
		return nil
	}
	return &user.UserInfo{
		Id:       u.ID,
		Uid:      u.UID,
		Nickname: u.Nickname,
		Phone:    u.Phone,
	}
}

func generateToken(userID, uid uint64) string {
	return fmt.Sprintf("%d-%d-%s", userID, uid, uuid.NewString())
}
