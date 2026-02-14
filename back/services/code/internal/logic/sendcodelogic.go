package logic

import (
	"context"
	"fmt"
	"strings"
	"text/template"
	"time"

	"SLGaming/back/services/code/code"
	"SLGaming/back/services/code/internal/helper"
	"SLGaming/back/services/code/internal/metrics"
	"SLGaming/back/services/code/internal/svc"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
)

const defaultCodeLength = 6
const defaultExpireSeconds = 300

const defaultPhoneSendInterval = 60

type SendCodeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSendCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendCodeLogic {
	return &SendCodeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SendCodeLogic) SendCode(in *code.SendCodeRequest) (*code.SendCodeResponse, error) {
	start := time.Now()
	purpose := in.GetPurpose()
	phone := in.GetPhone()
	maskedPhone := helper.MaskPhone(phone)

	helper.LogRequest(l.Logger, helper.OpSendCode, map[string]interface{}{
		"phone":   maskedPhone,
		"purpose": purpose,
	})

	getTemplate := l.getTemplate(purpose)

	if err := l.checkPhoneRateLimit(phone, purpose, getTemplate.MaxDailySends); err != nil {
		metrics.CodeRateLimitTotal.WithLabelValues("phone_rate_limit").Inc()
		return nil, err
	}

	expire := time.Duration(getTemplate.ExpireSeconds) * time.Second
	if expire <= 0 {
		expire = defaultExpireSeconds * time.Second
	}

	key := fmt.Sprintf("code:%s:%s", purpose, phone)

	codeValue, err := generateCode(getTemplate.CodeLength)
	if err != nil {
		helper.LogError(l.Logger, helper.OpSendCode, "generate code failed", err, map[string]interface{}{
			"phone": maskedPhone,
		})
		return nil, err
	}

	ttl, err := l.svcCtx.Redis.Ttl(key)
	if err != nil {
		helper.LogError(l.Logger, helper.OpSendCode, "check code ttl failed", err, map[string]interface{}{
			"phone": maskedPhone,
			"key":   key,
		})
	} else if ttl > 0 {
		helper.LogWarning(l.Logger, helper.OpSendCode, "code not expired", map[string]interface{}{
			"phone":      maskedPhone,
			"remain_ttl": ttl,
		})
		return nil, fmt.Errorf("验证码尚未过期，请 %d 秒后再试", ttl)
	}

	err = l.svcCtx.Redis.Setex(key, codeValue, int(expire/time.Second))
	if err != nil {
		metrics.CodeRedisErrorTotal.Inc()
		metrics.CodeSendTotal.WithLabelValues(purpose, "failure").Inc()
		metrics.CodeSendDuration.WithLabelValues(purpose).Observe(time.Since(start).Seconds())
		helper.LogError(l.Logger, helper.OpSendCode, "redis set code failed", err, map[string]interface{}{
			"phone": maskedPhone,
			"key":   key,
		})
		return nil, fmt.Errorf("set code failed: %w", err)
	}

	l.updateRateLimitCounters(phone, purpose)

	metrics.CodeSendTotal.WithLabelValues(purpose, "success").Inc()
	metrics.CodeSendDuration.WithLabelValues(purpose).Observe(time.Since(start).Seconds())

	expireAt := time.Now().Add(expire).Unix()
	helper.LogSuccess(l.Logger, helper.OpSendCode, map[string]interface{}{
		"phone":     maskedPhone,
		"purpose":   purpose,
		"expire_at": expireAt,
	})

	helper.LogInfo(l.Logger, helper.OpSendCode, "code content", map[string]interface{}{
		"content": renderTemplate(getTemplate.Content, codeValue, int(expire/time.Minute)),
	})

	return &code.SendCodeResponse{
		RequestId: uuid.NewString(),
		ExpireAt:  expireAt,
	}, nil
}

func (l *SendCodeLogic) getTemplate(purpose string) struct {
	CodeLength    int
	ExpireSeconds int64
	Content       string
	MaxDailySends int
} {
	if tpl, ok := l.svcCtx.Config.Template[purpose]; ok {
		result := struct {
			CodeLength    int
			ExpireSeconds int64
			Content       string
			MaxDailySends int
		}{
			CodeLength:    tpl.CodeLength,
			ExpireSeconds: tpl.ExpireSeconds,
			Content:       tpl.ContentTemplate,
			MaxDailySends: tpl.MaxDailySends,
		}
		if result.CodeLength <= 0 {
			result.CodeLength = defaultCodeLength
		}
		if result.ExpireSeconds <= 0 {
			result.ExpireSeconds = defaultExpireSeconds
		}
		if result.MaxDailySends <= 0 {
			result.MaxDailySends = 10
		}
		return result
	}
	return struct {
		CodeLength    int
		ExpireSeconds int64
		Content       string
		MaxDailySends int
	}{
		CodeLength:    defaultCodeLength,
		ExpireSeconds: defaultExpireSeconds,
		Content:       "",
		MaxDailySends: 10,
	}
}

func generateCode(length int) (string, error) {
	return "123456", nil
}

func renderTemplate(content string, code string, expireMinutes int) string {
	if !strings.Contains(content, "{{") {
		return content
	}

	var buf strings.Builder
	tmpl, err := template.New("sms").Parse(content)
	if err != nil {
		logx.Errorf("parse template failed: %v", err)
		return content
	}
	err = tmpl.Execute(&buf, map[string]any{
		"Code":          code,
		"ExpireMinutes": expireMinutes,
	})
	if err != nil {
		logx.Errorf("render template failed: %v", err)
		return content
	}
	return buf.String()
}

func (l *SendCodeLogic) checkPhoneRateLimit(phone, purpose string, maxDailySends int) error {
	if maxDailySends <= 0 {
		maxDailySends = 10
	}

	cfg := l.svcCtx.Config.RateLimit
	sendInterval := cfg.PhoneSendInterval
	if sendInterval <= 0 {
		sendInterval = defaultPhoneSendInterval
	}

	maskedPhone := helper.MaskPhone(phone)

	phoneLockKey := fmt.Sprintf("rate:phone:lock:%s", phone)
	exists, err := l.svcCtx.Redis.Exists(phoneLockKey)
	if err != nil {
		helper.LogError(l.Logger, helper.OpSendCode, "check phone lock exists failed", err, map[string]interface{}{
			"phone": maskedPhone,
		})
	} else if exists {
		ttl, err := l.svcCtx.Redis.Ttl(phoneLockKey)
		if err != nil {
			helper.LogError(l.Logger, helper.OpSendCode, "check phone lock ttl failed", err, map[string]interface{}{
				"phone": maskedPhone,
			})
			return fmt.Errorf("手机号发送过于频繁，请稍后再试")
		}
		if ttl > 0 {
			metrics.CodeRateLimitTotal.WithLabelValues("phone_interval").Inc()
			helper.LogWarning(l.Logger, helper.OpSendCode, "rate limited: phone interval", map[string]interface{}{
				"phone":      maskedPhone,
				"remain_ttl": ttl,
				"type":       "phone_interval",
			})
			return fmt.Errorf("手机号发送过于频繁，请 %d 秒后再试", ttl)
		}
	}

	today := time.Now().Format("20060102")
	phoneDailyKey := fmt.Sprintf("rate:phone:daily:%s:%s:%s", phone, purpose, today)
	countStr, err := l.svcCtx.Redis.Get(phoneDailyKey)
	if err != nil {
		helper.LogError(l.Logger, helper.OpSendCode, "get phone daily count failed", err, map[string]interface{}{
			"phone": maskedPhone,
			"key":   phoneDailyKey,
		})
	} else {
		count := 0
		if countStr != "" {
			if _, err := fmt.Sscanf(countStr, "%d", &count); err != nil {
				helper.LogError(l.Logger, helper.OpSendCode, "parse phone daily count failed", err, map[string]interface{}{
					"phone": maskedPhone,
					"value": countStr,
				})
			}
		}
		if count >= maxDailySends {
			metrics.CodeRateLimitTotal.WithLabelValues("phone_daily_limit").Inc()
			helper.LogWarning(l.Logger, helper.OpSendCode, "rate limited: daily limit exceeded", map[string]interface{}{
				"phone":       maskedPhone,
				"daily_count": count,
				"max_limit":   maxDailySends,
				"type":        "phone_daily_limit",
			})
			return fmt.Errorf("该手机号今日发送次数已达上限（%d次）", maxDailySends)
		}
	}

	return nil
}

func (l *SendCodeLogic) updateRateLimitCounters(phone, purpose string) {
	cfg := l.svcCtx.Config.RateLimit
	today := time.Now().Format("20060102")
	maskedPhone := helper.MaskPhone(phone)

	phoneSendInterval := cfg.PhoneSendInterval
	if phoneSendInterval <= 0 {
		phoneSendInterval = defaultPhoneSendInterval
	}
	phoneLockKey := fmt.Sprintf("rate:phone:lock:%s", phone)
	if err := l.svcCtx.Redis.Setex(phoneLockKey, "1", phoneSendInterval); err != nil {
		helper.LogError(l.Logger, helper.OpSendCode, "set phone lock failed", err, map[string]interface{}{
			"phone": maskedPhone,
			"key":   phoneLockKey,
		})
	}

	phoneDailyKey := fmt.Sprintf("rate:phone:daily:%s:%s:%s", phone, purpose, today)
	_, err := l.svcCtx.Redis.Incr(phoneDailyKey)
	if err != nil {
		helper.LogError(l.Logger, helper.OpSendCode, "incr phone daily count failed", err, map[string]interface{}{
			"phone": maskedPhone,
			"key":   phoneDailyKey,
		})
	} else {
		remainingSeconds := 86400 - (time.Now().Unix() % 86400)
		if err := l.svcCtx.Redis.Expire(phoneDailyKey, int(remainingSeconds)); err != nil {
			helper.LogError(l.Logger, helper.OpSendCode, "expire phone daily count failed", err, map[string]interface{}{
				"phone": maskedPhone,
				"key":   phoneDailyKey,
			})
		}
	}
}
