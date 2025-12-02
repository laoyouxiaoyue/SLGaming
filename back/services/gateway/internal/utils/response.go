package utils

import (
	"context"
	"net/http"
	"reflect"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// WriteResponse 根据响应码返回正确的 HTTP 状态码
func WriteResponse(ctx context.Context, w http.ResponseWriter, resp interface{}) {
	code := extractCode(resp)
	statusCode := mapStatusCode(code)
	httpx.WriteJsonCtx(ctx, w, statusCode, resp)
}

// extractCode 通用地从响应结构体中提取 Code 字段
// 优先查找顶层 Code 字段，其次查找 BaseResp.Code
func extractCode(resp interface{}) int32 {
	if resp == nil {
		return 0
	}

	v := reflect.ValueOf(resp)
	if v.Kind() == reflect.Ptr && !v.IsNil() {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return 0
	}

	// 顶层 Code 字段
	if f := v.FieldByName("Code"); f.IsValid() && f.Kind() == reflect.Int32 {
		return int32(f.Int())
	}

	// BaseResp.Code
	if base := v.FieldByName("BaseResp"); base.IsValid() && base.Kind() == reflect.Struct {
		if f := base.FieldByName("Code"); f.IsValid() && f.Kind() == reflect.Int32 {
			return int32(f.Int())
		}
	}

	return 0
}

// mapStatusCode 将业务 code 映射为 HTTP 状态码
func mapStatusCode(code int32) int {
	switch code {
	case 0:
		return http.StatusOK
	case 400:
		return http.StatusBadRequest
	case 401:
		return http.StatusUnauthorized
	case 403:
		return http.StatusForbidden
	case 404:
		return http.StatusNotFound
	case 500:
		return http.StatusInternalServerError
	default:
		if code != 0 {
			return http.StatusBadRequest
		}
		return http.StatusOK
	}
}
