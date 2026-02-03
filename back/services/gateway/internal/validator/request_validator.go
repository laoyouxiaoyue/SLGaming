package validator

import (
	"fmt"

	"SLGaming/back/services/gateway/internal/types"
)

// ValidateRegisterRequest 验证注册请求
func ValidateRegisterRequest(req *types.RegisterRequest) error {
	if err := ValidatePhone(req.Phone); err != nil {
		return err
	}
	if err := ValidateCode(req.Code); err != nil {
		return err
	}
	if err := ValidatePassword(req.Password); err != nil {
		return err
	}
	if req.Nickname != "" {
		if err := ValidateNickname(req.Nickname); err != nil {
			return err
		}
	}
	return nil
}

// ValidateLoginRequest 验证登录请求
func ValidateLoginRequest(req *types.LoginRequest) error {
	if err := ValidatePhone(req.Phone); err != nil {
		return err
	}
	if err := ValidatePassword(req.Password); err != nil {
		return err
	}
	return nil
}

// ValidateLoginByCodeRequest 验证验证码登录请求
func ValidateLoginByCodeRequest(req *types.LoginByCodeRequest) error {
	if err := ValidatePhone(req.Phone); err != nil {
		return err
	}
	if err := ValidateCode(req.Code); err != nil {
		return err
	}
	return nil
}

// ValidateSendCodeRequest 验证发送验证码请求
func ValidateSendCodeRequest(req *types.SendCodeRequest) error {
	if err := ValidatePhone(req.Phone); err != nil {
		return err
	}
	if err := ValidatePurpose(req.Purpose); err != nil {
		return err
	}
	return nil
}

// ValidateForgetPasswordRequest 验证忘记密码请求
func ValidateForgetPasswordRequest(req *types.ForgetPasswordRequest) error {
	if err := ValidatePhone(req.Phone); err != nil {
		return err
	}
	if err := ValidateCode(req.Code); err != nil {
		return err
	}
	if err := ValidatePassword(req.Password); err != nil {
		return err
	}
	return nil
}

// ValidateChangePasswordRequest 验证修改密码请求
func ValidateChangePasswordRequest(req *types.ChangePasswordRequest) error {
	if err := ValidatePhone(req.OldPhone); err != nil {
		return err
	}
	if err := ValidateCode(req.OldCode); err != nil {
		return err
	}
	if err := ValidatePassword(req.NewPassword); err != nil {
		return err
	}
	return nil
}

// ValidateUpdateUserRequest 验证更新用户请求
func ValidateUpdateUserRequest(req *types.UpdateUserRequest) error {
	if req.Id == 0 {
		return fmt.Errorf("用户ID不能为空")
	}
	if req.Phone != "" {
		if err := ValidatePhone(req.Phone); err != nil {
			return err
		}
	}
	if req.Password != "" {
		if err := ValidatePassword(req.Password); err != nil {
			return err
		}
	}
	if req.Nickname != "" {
		if err := ValidateNickname(req.Nickname); err != nil {
			return err
		}
	}
	return nil
}
