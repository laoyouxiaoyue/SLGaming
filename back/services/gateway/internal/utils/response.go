package utils

import (
	"context"
	"net/http"

	"SLGaming/back/services/gateway/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// WriteResponse 根据响应码返回正确的 HTTP 状态码
func WriteResponse(ctx context.Context, w http.ResponseWriter, resp interface{}) {
	var code int32 = 0
	var statusCode int = http.StatusOK

	// 根据响应类型提取 code
	switch v := resp.(type) {
	case *types.RegisterResponse:
		if v != nil {
			code = v.Code
		}
	case *types.LoginResponse:
		if v != nil {
			code = v.Code
		}
	case *types.LoginByCodeResponse:
		if v != nil {
			code = v.Code
		}
	case *types.ForgetPasswordResponse:
		if v != nil {
			code = v.Code
		}
	case *types.SendCodeResponse:
		if v != nil {
			code = v.Code
		}
	case *types.GetUserResponse:
		if v != nil {
			code = v.Code
		}
	case *types.UpdateUserResponse:
		if v != nil {
			code = v.Code
		}
	case *types.RefreshTokenResponse:
		if v != nil {
			code = v.Code
		}
	case *types.LogoutResponse:
		if v != nil {
			code = v.Code
		}
	case *types.GetWalletResponse:
		if v != nil {
			code = v.Code
		}
	case *types.GetCompanionProfileResponse:
		if v != nil {
			code = v.Code
		}
	case *types.UpdateCompanionProfileResponse:
		if v != nil {
			code = v.Code
		}
	case *types.GetCompanionListResponse:
		if v != nil {
			code = v.Code
		}
	}

	// 根据业务 code 映射到 HTTP 状态码
	switch code {
	case 0:
		statusCode = http.StatusOK
	case 400:
		statusCode = http.StatusBadRequest
	case 401:
		statusCode = http.StatusUnauthorized
	case 403:
		statusCode = http.StatusForbidden
	case 404:
		statusCode = http.StatusNotFound
	case 500:
		statusCode = http.StatusInternalServerError
	default:
		// 如果 code 不是 0，默认返回 400
		if code != 0 {
			statusCode = http.StatusBadRequest
		}
	}

	// 写入响应（httpx.WriteJsonCtx 会自动设置状态码）
	httpx.WriteJsonCtx(ctx, w, statusCode, resp)
}
