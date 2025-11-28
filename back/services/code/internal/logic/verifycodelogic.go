package logic

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"SLGaming/back/services/code/code"
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
	phone := strings.TrimSpace(in.GetPhone())
	if phone == "" {
		phone = "unknown"
	}

	clientIP := l.getClientIP()
	if clientIP == "" {
		clientIP = "unknown"
	}

	if err := l.checkVerifyPhoneDailyLimit(phone); err != nil {
		return nil, err
	}
	if err := l.checkVerifyIPDailyLimit(clientIP); err != nil {
		return nil, err
	}

	key := fmt.Sprintf("code:%s:%s", in.GetPurpose(), in.GetPhone())
	val, err := l.svcCtx.Redis.Get(key)
	if err != nil {
		// Redis 查询失败，可能是验证码不存在或网络问题
		// 返回验证失败，但不暴露具体错误信息
		l.Errorf("get verification code failed: %v, key: %s", err, key)
		l.recordVerifyPhoneUsage(phone)
		l.recordVerifyIPUsage(clientIP)
		return &code.VerifyCodeResponse{
			Passed: false,
		}, nil
	}

	passed := val == in.GetCode()
	if passed {
		// 验证成功，删除验证码
		_, err := l.svcCtx.Redis.Del(key)
		if err != nil {
			// 删除失败不影响验证结果，但记录日志
			l.Errorf("delete verification code failed: %v, key: %s", err, key)
		}
	}

	l.recordVerifyPhoneUsage(phone)
	l.recordVerifyIPUsage(clientIP)

	return &code.VerifyCodeResponse{
		Passed: passed,
	}, nil
}

// checkVerifyPhoneDailyLimit 检查手机号每日验证次数限制
func (l *VerifyCodeLogic) checkVerifyPhoneDailyLimit(phone string) error {
	cfg := l.svcCtx.Config.RateLimit
	dailyLimit := cfg.VerifyPhoneDailyLimit
	if dailyLimit <= 0 {
		dailyLimit = defaultVerifyPhoneDailyLimit
	}

	today := time.Now().Format("20060102")
	key := fmt.Sprintf("verify:phone:daily:%s:%s", phone, today)

	countStr, err := l.svcCtx.Redis.Get(key)
	if err != nil {
		// Redis 查询失败，记录日志但不阻止验证（可能是网络问题或计数不存在）
		l.Errorf("get verify phone daily count failed: %v, key: %s", err, key)
		return nil
	}

	count := 0
	if countStr != "" {
		if _, err := fmt.Sscanf(countStr, "%d", &count); err != nil {
			l.Errorf("parse verify phone daily count failed: %v, value: %s", err, countStr)
		}
	}

	if count >= dailyLimit {
		return fmt.Errorf("该手机号今日验证次数已达上限（%d次）", dailyLimit)
	}

	return nil
}

// checkVerifyIPDailyLimit 检查IP每日验证次数限制
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
		l.Errorf("get verify ip daily count failed: %v, key: %s", err, key)
		return nil
	}

	count := 0
	if countStr != "" {
		if _, err := fmt.Sscanf(countStr, "%d", &count); err != nil {
			l.Errorf("parse verify ip daily count failed: %v, value: %s", err, countStr)
		}
	}

	if count >= dailyLimit {
		return fmt.Errorf("该IP今日验证次数已达上限（%d次）", dailyLimit)
	}

	return nil
}

// recordVerifyPhoneUsage 记录手机号验证次数
func (l *VerifyCodeLogic) recordVerifyPhoneUsage(phone string) {
	today := time.Now().Format("20060102")
	key := fmt.Sprintf("verify:phone:daily:%s:%s", phone, today)

	if _, err := l.svcCtx.Redis.Incr(key); err != nil {
		l.Errorf("incr verify phone daily count failed: %v, key: %s", err, key)
		return
	}

	// 设置过期时间为当天结束
	remainingSeconds := 86400 - int(time.Now().Unix()%86400)
	if err := l.svcCtx.Redis.Expire(key, remainingSeconds); err != nil {
		l.Errorf("expire verify phone daily count failed: %v, key: %s", err, key)
	}
}

// recordVerifyIPUsage 记录IP验证次数
func (l *VerifyCodeLogic) recordVerifyIPUsage(ip string) {
	today := time.Now().Format("20060102")
	key := fmt.Sprintf("verify:ip:daily:%s:%s", ip, today)

	if _, err := l.svcCtx.Redis.Incr(key); err != nil {
		l.Errorf("incr verify ip daily count failed: %v, key: %s", err, key)
		return
	}

	remainingSeconds := 86400 - int(time.Now().Unix()%86400)
	if err := l.svcCtx.Redis.Expire(key, remainingSeconds); err != nil {
		l.Errorf("expire verify ip daily count failed: %v, key: %s", err, key)
	}
}

// getClientIP 从 gRPC context 中获取客户端IP
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
