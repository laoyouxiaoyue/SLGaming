package utils

import (
	"errors"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// HandleRPCError 处理 RPC 错误，返回用户友好的错误信息，隐藏内部错误详情
// 返回 (code, message) 用于构建 BaseResp
func HandleRPCError(err error, logger logx.Logger, operation string) (int32, string) {
	if err == nil {
		return 0, "success"
	}

	// 记录完整错误信息到日志（不返回给前端）
	if logger != nil {
		logger.Errorf("[%s] RPC error: %v", operation, err)
	}

	// 尝试解析为 gRPC status 错误
	st, ok := status.FromError(err)
	if !ok {
		// 非 gRPC 错误，返回通用错误信息
		return 500, "服务暂时不可用，请稍后重试"
	}

	// 根据 gRPC 错误码返回用户友好的错误信息
	code := st.Code()
	msg := st.Message()

	// 根据错误码映射为用户友好的消息
	switch code {
	case codes.InvalidArgument:
		// 参数错误：如果消息是用户友好的，直接返回；否则返回通用消息
		if isUserFriendlyMessage(msg) {
			return 400, msg
		}
		return 400, "请求参数错误"
	case codes.Unauthenticated:
		return 401, "未登录或登录已过期"
	case codes.PermissionDenied:
		// 权限错误：如果消息是用户友好的，直接返回；否则返回通用消息
		if isUserFriendlyMessage(msg) {
			return 403, msg
		}
		return 403, "权限不足"
	case codes.NotFound:
		// 资源不存在：如果消息是用户友好的，直接返回；否则返回通用消息
		if isUserFriendlyMessage(msg) {
			return 404, msg
		}
		return 404, "资源不存在"
	case codes.AlreadyExists:
		return 400, "资源已存在"
	case codes.ResourceExhausted:
		return 429, "请求过于频繁，请稍后重试"
	case codes.FailedPrecondition:
		// 前置条件失败：如果消息是用户友好的，直接返回；否则返回通用消息
		if isUserFriendlyMessage(msg) {
			return 400, msg
		}
		return 400, "操作失败，请检查数据状态"
	case codes.Aborted:
		return 409, "操作冲突，请重试"
	case codes.OutOfRange:
		return 400, "参数超出范围"
	case codes.Unimplemented:
		return 501, "功能暂未实现"
	case codes.Internal:
		// 内部错误：不返回具体错误信息，避免泄露内部细节
		return 500, "服务内部错误，请稍后重试"
	case codes.Unavailable:
		return 503, "服务暂时不可用，请稍后重试"
	case codes.DeadlineExceeded:
		return 504, "请求超时，请稍后重试"
	case codes.Canceled:
		return 499, "请求已取消"
	case codes.Unknown:
		// 未知错误：不返回具体错误信息
		return 500, "服务异常，请稍后重试"
	default:
		// 其他错误：不返回具体错误信息
		return 500, "服务异常，请稍后重试"
	}
}

// isUserFriendlyMessage 判断错误消息是否是用户友好的
// 如果消息包含技术细节（如数据库错误、SQL 错误等），返回 false
func isUserFriendlyMessage(msg string) bool {
	if msg == "" {
		return false
	}

	// 检查是否包含技术性关键词
	technicalKeywords := []string{
		"sql",
		"database",
		"connection",
		"timeout",
		"dial",
		"rpc",
		"grpc",
		"internal",
		"error",
		"failed",
		"panic",
		"stack",
		"trace",
		"goroutine",
		"nil pointer",
		"index out of range",
		"invalid memory",
		"EOF",
		"context canceled",
		"no such file",
		"permission denied",
		"access denied",
	}

	lowerMsg := strings.ToLower(msg)
	for _, keyword := range technicalKeywords {
		if strings.Contains(lowerMsg, keyword) {
			return false
		}
	}

	// 检查是否包含常见的用户友好消息
	userFriendlyPatterns := []string{
		"不存在",
		"已存在",
		"已过期",
		"无效",
		"错误",
		"失败",
		"不能",
		"不允许",
		"需要",
		"必须",
		"缺少",
		"格式",
		"长度",
		"范围",
	}

	for _, pattern := range userFriendlyPatterns {
		if strings.Contains(msg, pattern) {
			return true
		}
	}

	// 如果消息很短且不包含技术关键词，可能是用户友好的
	if len(msg) < 50 && !strings.Contains(msg, ":") && !strings.Contains(msg, "(") {
		return true
	}

	return false
}

// HandleError 处理通用错误，返回用户友好的错误信息
// 用于处理非 RPC 错误（如参数验证错误、业务逻辑错误等）
func HandleError(err error, logger logx.Logger, operation string) (int32, string) {
	if err == nil {
		return 0, "success"
	}

	// 记录完整错误信息到日志
	if logger != nil {
		logger.Errorf("[%s] error: %v", operation, err)
	}

	// 检查是否是用户友好的错误消息
	errMsg := err.Error()
	if isUserFriendlyMessage(errMsg) {
		return 400, errMsg
	}

	// 检查是否是常见的错误类型
	if errors.Is(err, errors.New("validation failed")) {
		return 400, "参数验证失败"
	}

	// 默认返回通用错误信息
	return 500, "操作失败，请稍后重试"
}

// BuildErrorResponse 构建错误响应（用于返回给前端）
// 返回格式：BaseResp{Code: code, Msg: message}
func BuildErrorResponse(code int32, message string) map[string]interface{} {
	return map[string]interface{}{
		"code": code,
		"msg":  message,
	}
}

// FormatRPCError 格式化 RPC 错误为字符串（仅用于日志）
func FormatRPCError(err error) string {
	if err == nil {
		return "nil"
	}

	st, ok := status.FromError(err)
	if !ok {
		return fmt.Sprintf("non-gRPC error: %v", err)
	}

	return fmt.Sprintf("gRPC error: code=%s, msg=%s", st.Code(), st.Message())
}
