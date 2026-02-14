package helper

import (
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"
)

type LogOperation string

const (
	OpCreateOrder   LogOperation = "create_order"
	OpCancelOrder   LogOperation = "cancel_order"
	OpAcceptOrder   LogOperation = "accept_order"
	OpCompleteOrder LogOperation = "complete_order"
	OpStartOrder    LogOperation = "start_order"
	OpRateOrder     LogOperation = "rate_order"
	OpGetOrder      LogOperation = "get_order"
	OpGetOrderList  LogOperation = "get_order_list"
	OpDeleteOrder   LogOperation = "delete_order"
	OpPaymentEvent  LogOperation = "payment_event"
	OpServer        LogOperation = "server"
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

func MaskOrderNo(orderNo string) string {
	if len(orderNo) < 8 {
		return orderNo
	}
	return orderNo[:4] + "****" + orderNo[len(orderNo)-4:]
}
