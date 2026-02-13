package utils

import (
	"errors"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var operationMessages = map[string]string{
	"Login":                  "登录成功",
	"Register":               "注册成功",
	"Logout":                 "退出登录成功",
	"RefreshToken":           "令牌刷新成功",
	"GetUser":                "获取用户信息成功",
	"UpdateUser":             "更新用户信息成功",
	"GetWallet":              "获取钱包信息成功",
	"Recharge":               "充值成功",
	"RechargeCreate":         "创建充值订单成功",
	"RechargeQuery":          "查询充值记录成功",
	"RechargeList":           "获取充值列表成功",
	"Consume":                "消费成功",
	"GetCompanionList":       "获取陪玩列表成功",
	"GetCompanionProfile":    "获取陪玩资料成功",
	"GetCompanionById":       "获取陪玩详情成功",
	"ApplyCompanion":         "申请成为陪玩成功",
	"UpdateCompanionProfile": "更新陪玩资料成功",
	"UpdateCompanionStatus":  "更新陪玩状态成功",
	"ChangePassword":         "修改密码成功",
	"ChangePhone":            "修改手机号成功",
	"GetRanking":             "获取排行榜成功",
	"CreateOrder":            "创建订单成功",
	"GetOrder":               "获取订单成功",
	"GetOrderList":           "获取订单列表成功",
	"AcceptOrder":            "接单成功",
	"StartOrder":             "开始订单成功",
	"CompleteOrder":          "完成订单成功",
	"CancelOrder":            "取消订单成功",
	"RateOrder":              "评价订单成功",
	"DeleteOrder":            "删除订单成功",
	"FollowUser":             "关注成功",
	"UnfollowUser":           "取消关注成功",
	"GetFollowers":           "获取粉丝列表成功",
	"GetFollowing":           "获取关注列表成功",
	"GetMutualFollow":        "获取互关列表成功",
	"CheckFollowStatus":      "查询关注状态成功",
	"ListGameSkills":         "获取游戏技能列表成功",
	"VerifyCode":             "验证码验证成功",
	"RecommendCompanion":     "推荐陪玩成功",
	"GetCode":                "获取验证码成功",
	"UploadAvatar":           "头像上传成功",
	"AlipayNotify":           "支付宝回调处理成功",
}

func GetSuccessMsg(operation string) string {
	if msg, ok := operationMessages[operation]; ok {
		return msg
	}
	return "操作成功"
}

var errorMessages = map[string]map[codes.Code]string{
	"Login": {
		codes.InvalidArgument:  "登录失败：手机号或密码不能为空",
		codes.NotFound:         "登录失败：用户不存在",
		codes.PermissionDenied: "登录失败：密码错误",
		codes.Internal:         "登录失败：服务异常，请稍后重试",
	},
	"Register": {
		codes.InvalidArgument: "注册失败：参数错误",
		codes.AlreadyExists:   "注册失败：手机号已被注册",
		codes.Internal:        "注册失败：服务异常，请稍后重试",
	},
	"VerifyCode": {
		codes.InvalidArgument: "验证码错误：请输入正确的验证码",
		codes.NotFound:        "验证码错误：验证码不存在或已过期",
		codes.Internal:        "验证码验证失败：服务异常",
	},
	"CreateOrder": {
		codes.InvalidArgument:    "创建订单失败：参数错误",
		codes.NotFound:           "创建订单失败：陪玩不存在",
		codes.FailedPrecondition: "创建订单失败：陪玩当前不可接单",
		codes.PermissionDenied:   "创建订单失败：余额不足",
		codes.Internal:           "创建订单失败：服务异常",
	},
	"GetOrder": {
		codes.NotFound:         "订单不存在",
		codes.PermissionDenied: "无权查看该订单",
	},
	"AcceptOrder": {
		codes.InvalidArgument:    "接单失败：参数错误",
		codes.NotFound:           "接单失败：订单不存在",
		codes.FailedPrecondition: "接单失败：订单已被接取",
		codes.Internal:           "接单失败：服务异常",
	},
	"StartOrder": {
		codes.InvalidArgument:    "开始订单失败：参数错误",
		codes.NotFound:           "开始订单失败：订单不存在",
		codes.FailedPrecondition: "开始订单失败：订单状态不允许开始",
		codes.Internal:           "开始订单失败：服务异常",
	},
	"CompleteOrder": {
		codes.InvalidArgument:    "完成订单失败：参数错误",
		codes.NotFound:           "完成订单失败：订单不存在",
		codes.FailedPrecondition: "完成订单失败：订单状态不允许完成",
		codes.Internal:           "完成订单失败：服务异常",
	},
	"CancelOrder": {
		codes.InvalidArgument:    "取消订单失败：参数错误",
		codes.NotFound:           "取消订单失败：订单不存在",
		codes.FailedPrecondition: "取消订单失败：订单已取消或已完成",
		codes.Internal:           "取消订单失败：服务异常",
	},
	"GetUser": {
		codes.NotFound: "用户不存在",
		codes.Internal: "获取用户信息失败：服务异常",
	},
	"UpdateUser": {
		codes.InvalidArgument: "更新用户信息失败：参数错误",
		codes.Internal:        "更新用户信息失败：服务异常",
	},
	"GetCompanionProfile": {
		codes.NotFound: "陪玩资料不存在",
		codes.Internal: "获取陪玩资料失败：服务异常",
	},
	"UpdateCompanionProfile": {
		codes.InvalidArgument:    "更新陪玩资料失败：参数错误",
		codes.NotFound:           "更新陪玩资料失败：陪玩资料不存在",
		codes.FailedPrecondition: "更新陪玩资料失败：您不是陪玩用户",
		codes.Internal:           "更新陪玩资料失败：服务异常",
	},
	"FollowUser": {
		codes.InvalidArgument: "关注失败：参数错误",
		codes.AlreadyExists:   "关注失败：您已关注该用户",
		codes.NotFound:        "关注失败：用户不存在",
		codes.Internal:        "关注失败：服务异常",
	},
	"UnfollowUser": {
		codes.InvalidArgument: "取消关注失败：参数错误",
		codes.NotFound:        "取消关注失败：您未关注该用户",
		codes.Internal:        "取消关注失败：服务异常",
	},
	"GetWallet": {
		codes.NotFound: "钱包不存在",
		codes.Internal: "获取钱包信息失败：服务异常",
	},
	"Recharge": {
		codes.InvalidArgument: "充值失败：参数错误",
		codes.Internal:        "充值失败：服务异常",
	},
	"ChangePassword": {
		codes.InvalidArgument: "修改密码失败：原密码错误",
		codes.Internal:        "修改密码失败：服务异常",
	},
	"ChangePhone": {
		codes.InvalidArgument: "修改手机号失败：参数错误",
		codes.AlreadyExists:   "修改手机号失败：新手机号已被使用",
		codes.Internal:        "修改手机号失败：服务异常",
	},
}

func GetErrorMsg(operation string, code codes.Code) string {
	if opMap, ok := errorMessages[operation]; ok {
		if msg, ok := opMap[code]; ok {
			return msg
		}
	}
	return ""
}

// HandleRPCError 处理 RPC 错误，返回用户友好的错误信息，隐藏内部错误详情
// 返回 (code, message) 用于构建 BaseResp
func HandleRPCError(err error, logger logx.Logger, operation string) (int32, string) {
	if err == nil {
		return 0, GetSuccessMsg(operation)
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

	// 先尝试获取操作特定的错误消息
	if specificMsg := GetErrorMsg(operation, code); specificMsg != "" {
		return int32(code), specificMsg
	}

	// 如果没有特定错误消息，使用通用的
	switch code {
	case codes.InvalidArgument:
		if isUserFriendlyMessage(msg) {
			return 400, msg
		}
		return 400, "请求参数错误，请检查输入是否正确"
	case codes.Unauthenticated:
		return 401, "未登录或登录已过期，请重新登录"
	case codes.PermissionDenied:
		if isUserFriendlyMessage(msg) {
			return 403, msg
		}
		return 403, "权限不足，无法执行此操作"
	case codes.NotFound:
		if isUserFriendlyMessage(msg) {
			return 404, msg
		}
		return 404, "请求的资源不存在"
	case codes.AlreadyExists:
		if isUserFriendlyMessage(msg) {
			return 400, msg
		}
		return 400, "资源已存在，请勿重复创建"
	case codes.ResourceExhausted:
		return 429, "请求过于频繁，请稍后再试"
	case codes.FailedPrecondition:
		// 前置条件失败：如果消息是用户友好的，直接返回；否则返回通用消息
		if isUserFriendlyMessage(msg) {
			return 400, msg
		}
		return 400, "操作失败，当前状态不允许此操作"
	case codes.Aborted:
		return 409, "操作冲突，请刷新后重试"
	case codes.OutOfRange:
		return 400, "参数超出有效范围"
	case codes.Unimplemented:
		return 501, "该功能暂未开放，敬请期待"
	case codes.Internal:
		// 内部错误：不返回具体错误信息，避免泄露内部细节
		return 500, "服务器内部错误，请稍后重试"
	case codes.Unavailable:
		return 503, "服务暂时不可用，请稍后重试"
	case codes.DeadlineExceeded:
		return 504, "请求超时，请检查网络后重试"
	case codes.Canceled:
		return 499, "请求已取消"
	case codes.Unknown:
		// 未知错误：不返回具体错误信息
		return 500, "未知错误，请稍后重试"
	default:
		// 其他错误：不返回具体错误信息
		return 500, "操作失败，请稍后重试"
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

// HandleRPCClientUnavailable 处理 RPC 客户端未初始化错误
// 返回 (code, message) 用于构建 BaseResp
func HandleRPCClientUnavailable(logger logx.Logger, service string) (int32, string) {
	if logger != nil {
		logger.Errorf("[%s] RPC client not initialized", service)
	}
	return 503, "服务暂时不可用，请稍后重试"
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
