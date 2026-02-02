package validator

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// ValidatePhone 验证手机号
func ValidatePhone(phone string) error {
	phone = strings.TrimSpace(phone)
	if phone == "" {
		return fmt.Errorf("手机号不能为空")
	}
	phone = strings.ReplaceAll(phone, " ", "")
	phone = strings.ReplaceAll(phone, "-", "")
	matched, _ := regexp.MatchString(`^1[3-9]\d{9}$`, phone)
	if !matched {
		return fmt.Errorf("手机号格式不正确，请输入11位中国大陆手机号")
	}
	return nil
}

// ValidateCode 验证验证码
func ValidateCode(code string) error {
	code = strings.TrimSpace(code)
	if code == "" {
		return fmt.Errorf("验证码不能为空")
	}
	matched, _ := regexp.MatchString(`^\d{4,6}$`, code)
	if !matched {
		return fmt.Errorf("验证码格式不正确，请输入4-6位数字")
	}
	return nil
}

// ValidatePassword 验证密码
func ValidatePassword(password string) error {
	password = strings.TrimSpace(password)
	if password == "" {
		return fmt.Errorf("密码不能为空")
	}
	if len(password) < 6 {
		return fmt.Errorf("密码长度不能少于6位")
	}
	if len(password) > 20 {
		return fmt.Errorf("密码长度不能超过20位")
	}
	hasLetter := false
	hasDigit := false
	for _, char := range password {
		if unicode.IsLetter(char) {
			hasLetter = true
		}
		if unicode.IsDigit(char) {
			hasDigit = true
		}
		if hasLetter && hasDigit {
			return nil
		}
	}
	if !hasLetter {
		return fmt.Errorf("密码必须包含至少一个字母")
	}
	if !hasDigit {
		return fmt.Errorf("密码必须包含至少一个数字")
	}
	return nil
}

// ValidatePurpose 验证验证码用途
func ValidatePurpose(purpose string) error {
	purpose = strings.TrimSpace(purpose)
	if purpose == "" {
		return fmt.Errorf("验证码用途不能为空")
	}
	validPurposes := map[string]bool{
		"register":         true,
		"login":            true,
		"forget_password":  true,
		"resetpassword":    true,
		"change_phone":     true,
		"change_phone_new": true,
		"change_password":  true,
	}
	if !validPurposes[purpose] {
		return fmt.Errorf("验证码用途不正确，支持: register, login, forget_password, change_phone, change_phone_new, change_password")
	}
	return nil
}

// ValidateNickname 验证昵称
func ValidateNickname(nickname string) error {
	if nickname == "" {
		return nil
	}
	nickname = strings.TrimSpace(nickname)
	if len([]rune(nickname)) > 32 {
		return fmt.Errorf("昵称长度不能超过32个字符")
	}
	matched, _ := regexp.MatchString(`^[\p{L}\p{N}_]+$`, nickname)
	if !matched {
		return fmt.Errorf("昵称只能包含中文、英文、数字和下划线")
	}
	return nil
}
