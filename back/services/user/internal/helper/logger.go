package helper

import (
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"
)

// LogOperation 日志操作类型
type LogOperation string

const (
	OpRegister                  LogOperation = "register"
	OpLogin                     LogOperation = "login"
	OpLoginByCode               LogOperation = "login_by_code"
	OpGetUser                   LogOperation = "get_user"
	OpUpdateUser                LogOperation = "update_user"
	OpForgetPassword            LogOperation = "forget_password"
	OpGetWallet                 LogOperation = "get_wallet"
	OpConsume                   LogOperation = "consume"
	OpRecharge                  LogOperation = "recharge"
	OpRefund                    LogOperation = "refund"
	OpGetCompanionList          LogOperation = "get_companion_list"
	OpGetCompanionProfile       LogOperation = "get_companion_profile"
	OpUpdateCompanionProfile    LogOperation = "update_companion_profile"
	OpUpdateCompanionStats      LogOperation = "update_companion_stats"
	OpGetCompanionRatingRanking LogOperation = "get_companion_rating_ranking"
	OpGetCompanionOrdersRanking LogOperation = "get_companion_orders_ranking"
	OpServer                    LogOperation = "server"
	OpFollow                    LogOperation = "follow"
	OpUnfollow                  LogOperation = "unfollow"
	OpMQConsumer                LogOperation = "mq_consumer"
)

// LogRequest 记录请求开始日志
// 格式：[Operation] request started: key1=value1, key2=value2
func LogRequest(logger logx.Logger, operation LogOperation, fields map[string]interface{}) {
	if logger == nil {
		return
	}
	logger.Infof("[%s] request started: %s", operation, formatFields(fields))
}

// LogSuccess 记录操作成功日志
// 格式：[Operation] succeeded: key1=value1, key2=value2
func LogSuccess(logger logx.Logger, operation LogOperation, fields map[string]interface{}) {
	if logger == nil {
		return
	}
	logger.Infof("[%s] succeeded: %s", operation, formatFields(fields))
}

// LogError 记录错误日志
// 格式：[Operation] failed: reason, key1=value1, key2=value2, error=xxx
func LogError(logger logx.Logger, operation LogOperation, reason string, err error, fields map[string]interface{}) {
	if logger == nil {
		return
	}
	if fields == nil {
		fields = make(map[string]interface{})
	}
	if err != nil {
		fields["error"] = err.Error()
	}
	logger.Errorf("[%s] failed: %s, %s", operation, reason, formatFields(fields))
}

// LogWarning 记录警告日志
// 格式：[Operation] warning: reason, key1=value1, key2=value2
func LogWarning(logger logx.Logger, operation LogOperation, reason string, fields map[string]interface{}) {
	if logger == nil {
		return
	}
	logger.Infof("[%s] warning: %s, %s", operation, reason, formatFields(fields))
}

// LogInfo 记录信息日志
// 格式：[Operation] info: message, key1=value1, key2=value2
func LogInfo(logger logx.Logger, operation LogOperation, message string, fields map[string]interface{}) {
	if logger == nil {
		return
	}
	logger.Infof("[%s] info: %s, %s", operation, message, formatFields(fields))
}

// formatFields 格式化字段为 key=value 格式
func formatFields(fields map[string]interface{}) string {
	if len(fields) == 0 {
		return ""
	}

	var parts []string
	for k, v := range fields {
		parts = append(parts, fmt.Sprintf("%s=%v", k, v))
	}

	return strings.Join(parts, ", ")
}
