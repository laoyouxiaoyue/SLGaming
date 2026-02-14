package helper

import (
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"
)

type LogOperation string

const (
	OpSendCode   LogOperation = "send_code"
	OpVerifyCode LogOperation = "verify_code"
	OpServer     LogOperation = "server"
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
