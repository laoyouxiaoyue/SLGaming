package logic

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"SLGaming/back/services/code/code"
	"SLGaming/back/services/code/internal/helper"
	"SLGaming/back/services/code/internal/metrics"
	"SLGaming/back/services/code/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/peer"
)

const (
	defaultVerifyPhoneDailyLimit = 50
	defaultVerifyIPDailyLimit    = 200
)

type VerifyCodeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewVerifyCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *VerifyCodeLogic {
	return &VerifyCodeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *VerifyCodeLogic) VerifyCode(in *code.VerifyCodeRequest) (*code.VerifyCodeResponse, error) {
	start := time.Now()
	phone := strings.TrimSpace(in.GetPhone())
	if phone == "" {
		phone = "unknown"
	}
	purpose := in.GetPurpose()
	maskedPhone := helper.MaskPhone(phone)

	clientIP := l.getClientIP()
	if clientIP == "" {
		clientIP = "unknown"
	}

	helper.LogRequest(l.Logger, helper.OpVerifyCode, map[string]interface{}{
		"phone":     maskedPhone,
		"purpose":   purpose,
		"client_ip": clientIP,
	})

	if err := l.checkVerifyPhoneDailyLimit(phone); err != nil {
		metrics.CodeRateLimitTotal.WithLabelValues("verify_phone_daily").Inc()
		metrics.CodeVerifyTotal.WithLabelValues("failure").Inc()
		metrics.CodeVerifyDuration.Observe(time.Since(start).Seconds())
		return nil, err
	}
	if err := l.checkVerifyIPDailyLimit(clientIP); err != nil {
		metrics.CodeRateLimitTotal.WithLabelValues("verify_ip_daily").Inc()
		metrics.CodeVerifyTotal.WithLabelValues("failure").Inc()
		metrics.CodeVerifyDuration.Observe(time.Since(start).Seconds())
		return nil, err
	}

	key := fmt.Sprintf("code:%s:%s", purpose, phone)
	val, err := l.svcCtx.Redis.Get(key)
	if err != nil {
		helper.LogError(l.Logger, helper.OpVerifyCode, "redis get code failed", err, map[string]interface{}{
			"phone": maskedPhone,
			"key":   key,
		})
		metrics.CodeRedisErrorTotal.Inc()
		metrics.CodeVerifyTotal.WithLabelValues("failure").Inc()
		metrics.CodeVerifyDuration.Observe(time.Since(start).Seconds())
		l.recordVerifyPhoneUsage(phone)
		l.recordVerifyIPUsage(clientIP)
		return &code.VerifyCodeResponse{
			Passed: false,
		}, nil
	}

	passed := val == in.GetCode()
	if passed {
		_, err := l.svcCtx.Redis.Del(key)
		if err != nil {
			helper.LogError(l.Logger, helper.OpVerifyCode, "redis del code failed", err, map[string]interface{}{
				"phone": maskedPhone,
				"key":   key,
			})
		}
		metrics.CodeVerifyTotal.WithLabelValues("success").Inc()
	} else {
		metrics.CodeVerifyTotal.WithLabelValues("failure").Inc()
	}

	metrics.CodeVerifyDuration.Observe(time.Since(start).Seconds())
	l.recordVerifyPhoneUsage(phone)
	l.recordVerifyIPUsage(clientIP)

	helper.LogSuccess(l.Logger, helper.OpVerifyCode, map[string]interface{}{
		"phone":     maskedPhone,
		"purpose":   purpose,
		"passed":    passed,
		"client_ip": clientIP,
	})

	return &code.VerifyCodeResponse{
		Passed: passed,
	}, nil
}

func (l *VerifyCodeLogic) checkVerifyPhoneDailyLimit(phone string) error {
	cfg := l.svcCtx.Config.RateLimit
	dailyLimit := cfg.VerifyPhoneDailyLimit
	if dailyLimit <= 0 {
		dailyLimit = defaultVerifyPhoneDailyLimit
	}

	maskedPhone := helper.MaskPhone(phone)
	today := time.Now().Format("20060102")
	key := fmt.Sprintf("verify:phone:daily:%s:%s", phone, today)

	countStr, err := l.svcCtx.Redis.Get(key)
	if err != nil {
		helper.LogError(l.Logger, helper.OpVerifyCode, "get verify phone daily count failed", err, map[string]interface{}{
			"phone": maskedPhone,
			"key":   key,
		})
		return nil
	}

	count := 0
	if countStr != "" {
		if _, err := fmt.Sscanf(countStr, "%d", &count); err != nil {
			helper.LogError(l.Logger, helper.OpVerifyCode, "parse verify phone daily count failed", err, map[string]interface{}{
				"phone": maskedPhone,
				"value": countStr,
			})
		}
	}

	if count >= dailyLimit {
		helper.LogWarning(l.Logger, helper.OpVerifyCode, "rate limited: phone daily limit exceeded", map[string]interface{}{
			"phone":       maskedPhone,
			"daily_count": count,
			"max_limit":   dailyLimit,
			"type":        "verify_phone_daily",
		})
		return fmt.Errorf("该手机号今日验证次数已达上限（%d次）", dailyLimit)
	}

	return nil
}

func (l *VerifyCodeLogic) checkVerifyIPDailyLimit(ip string) error {
	cfg := l.svcCtx.Config.RateLimit
	dailyLimit := cfg.VerifyIPDailyLimit
	if dailyLimit <= 0 {
		dailyLimit = defaultVerifyIPDailyLimit
	}

	today := time.Now().Format("20060102")
	key := fmt.Sprintf("verify:ip:daily:%s:%s", ip, today)

	countStr, err := l.svcCtx.Redis.Get(key)
	if err != nil {
		helper.LogError(l.Logger, helper.OpVerifyCode, "get verify ip daily count failed", err, map[string]interface{}{
			"client_ip": ip,
			"key":       key,
		})
		return nil
	}

	count := 0
	if countStr != "" {
		if _, err := fmt.Sscanf(countStr, "%d", &count); err != nil {
			helper.LogError(l.Logger, helper.OpVerifyCode, "parse verify ip daily count failed", err, map[string]interface{}{
				"client_ip": ip,
				"value":     countStr,
			})
		}
	}

	if count >= dailyLimit {
		helper.LogWarning(l.Logger, helper.OpVerifyCode, "rate limited: ip daily limit exceeded", map[string]interface{}{
			"client_ip":   ip,
			"daily_count": count,
			"max_limit":   dailyLimit,
			"type":        "verify_ip_daily",
		})
		return fmt.Errorf("该IP今日验证次数已达上限（%d次）", dailyLimit)
	}

	return nil
}

func (l *VerifyCodeLogic) recordVerifyPhoneUsage(phone string) {
	today := time.Now().Format("20060102")
	key := fmt.Sprintf("verify:phone:daily:%s:%s", phone, today)
	maskedPhone := helper.MaskPhone(phone)

	if _, err := l.svcCtx.Redis.Incr(key); err != nil {
		helper.LogError(l.Logger, helper.OpVerifyCode, "incr verify phone daily count failed", err, map[string]interface{}{
			"phone": maskedPhone,
			"key":   key,
		})
		return
	}

	remainingSeconds := 86400 - int(time.Now().Unix()%86400)
	if err := l.svcCtx.Redis.Expire(key, remainingSeconds); err != nil {
		helper.LogError(l.Logger, helper.OpVerifyCode, "expire verify phone daily count failed", err, map[string]interface{}{
			"phone": maskedPhone,
			"key":   key,
		})
	}
}

func (l *VerifyCodeLogic) recordVerifyIPUsage(ip string) {
	today := time.Now().Format("20060102")
	key := fmt.Sprintf("verify:ip:daily:%s:%s", ip, today)

	if _, err := l.svcCtx.Redis.Incr(key); err != nil {
		helper.LogError(l.Logger, helper.OpVerifyCode, "incr verify ip daily count failed", err, map[string]interface{}{
			"client_ip": ip,
			"key":       key,
		})
		return
	}

	remainingSeconds := 86400 - int(time.Now().Unix()%86400)
	if err := l.svcCtx.Redis.Expire(key, remainingSeconds); err != nil {
		helper.LogError(l.Logger, helper.OpVerifyCode, "expire verify ip daily count failed", err, map[string]interface{}{
			"client_ip": ip,
			"key":       key,
		})
	}
}

func (l *VerifyCodeLogic) getClientIP() string {
	p, ok := peer.FromContext(l.ctx)
	if !ok {
		return ""
	}
	addr := p.Addr.String()
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return addr
	}
	return host
}
