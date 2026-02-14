package helper

import (
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"
)

type LogOperation string

const (
	OpLogin             LogOperation = "login"
	OpLoginByCode       LogOperation = "login_by_code"
	OpRegister          LogOperation = "register"
	OpLogout            LogOperation = "logout"
	OpRefreshToken      LogOperation = "refresh_token"
	OpGetUser           LogOperation = "get_user"
	OpUpdateUser        LogOperation = "update_user"
	OpUploadAvatar      LogOperation = "upload_avatar"
	OpGetWallet         LogOperation = "get_wallet"
	OpCreateOrder       LogOperation = "create_order"
	OpGetOrder          LogOperation = "get_order"
	OpGetOrderList      LogOperation = "get_order_list"
	OpCancelOrder       LogOperation = "cancel_order"
	OpAcceptOrder       LogOperation = "accept_order"
	OpCompleteOrder     LogOperation = "complete_order"
	OpStartOrder        LogOperation = "start_order"
	OpRateOrder         LogOperation = "rate_order"
	OpSendCode          LogOperation = "send_code"
	OpFollowUser        LogOperation = "follow_user"
	OpUnfollowUser      LogOperation = "unfollow_user"
	OpCheckFollowStatus LogOperation = "check_follow_status"
	OpRechargeCreate    LogOperation = "recharge_create"
	OpRechargeQuery     LogOperation = "recharge_query"
	OpRechargeList      LogOperation = "recharge_list"
	OpAlipayNotify      LogOperation = "alipay_notify"
	OpServer            LogOperation = "server"
	OpAuth              LogOperation = "auth"
	OpRateLimit         LogOperation = "rate_limit"
)

func LogRequest(logger logx.Logger, operation LogOperation, fields map[string]interface{}) {
	if logger == nil {
		return
	}
	logger.Infof("[%s] request started: %s", operation, formatFields(fields))
}

func LogSuccess(logger logx.Logger, operation LogOperation, fields map[string]interface{}) {
	if logger == nil {
		return
	}
	logger.Infof("[%s] succeeded: %s", operation, formatFields(fields))
}

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

func LogWarning(logger logx.Logger, operation LogOperation, reason string, fields map[string]interface{}) {
	if logger == nil {
		return
	}
	logger.Infof("[%s] warning: %s, %s", operation, reason, formatFields(fields))
}

func LogInfo(logger logx.Logger, operation LogOperation, message string, fields map[string]interface{}) {
	if logger == nil {
		return
	}
	logger.Infof("[%s] info: %s, %s", operation, message, formatFields(fields))
}

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

func MaskPhone(phone string) string {
	if len(phone) < 7 {
		return phone
	}
	return phone[:3] + "****" + phone[len(phone)-4:]
}
