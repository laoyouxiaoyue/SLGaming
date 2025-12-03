package logic

import (
	"context"
	"fmt"
	"strings"
	"text/template"
	"time"

	"SLGaming/back/services/code/code"
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
	purpose := in.GetPurpose()
	getTemplate := l.getTemplate(purpose)

	// 手机号限流检查
	if err := l.checkPhoneRateLimit(in.GetPhone(), purpose, getTemplate.MaxDailySends); err != nil {
		return nil, err
	}

	expire := time.Duration(getTemplate.ExpireSeconds) * time.Second
	if expire <= 0 {
		expire = defaultExpireSeconds * time.Second
	}

	key := fmt.Sprintf("code:%s:%s", in.GetPurpose(), in.GetPhone())

	codeValue, err := generateCode(getTemplate.CodeLength)
	if err != nil {
		return nil, err
	}

	// 4. 检查验证码是否已存在（未过期）
	ttl, err := l.svcCtx.Redis.Ttl(key)
	if err != nil {
		// Redis 查询失败，记录日志但不阻止发送（可能是网络问题）
		l.Errorf("check code ttl failed: %v", err)
	} else if ttl > 0 {
		return nil, fmt.Errorf("验证码尚未过期，请 %d 秒后再试", ttl)
	}

	// 5. 使用 SETEX 设置验证码
	err = l.svcCtx.Redis.Setex(key, codeValue, int(expire/time.Second))
	if err != nil {
		return nil, fmt.Errorf("set code failed: %w", err)
	}

	// 更新限流计数
	l.updateRateLimitCounters(in.GetPhone(), purpose)

	logx.Infof("send code content: %s", renderTemplate(getTemplate.Content, codeValue, int(expire/time.Minute)))

	return &code.SendCodeResponse{
		RequestId: uuid.NewString(),
		ExpireAt:  time.Now().Add(expire).Unix(),
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
	// 暂时返回固定验证码，方便测试
	return "123456", nil

	// 原随机生成代码（已注释，需要时恢复）：
	// if length <= 0 {
	// 	length = defaultCodeLength
	// }
	// result := make([]byte, length)
	// for i := 0; i < length; i++ {
	// 	n, err := rand.Int(rand.Reader, big.NewInt(10))
	// 	if err != nil {
	// 		return "", err
	// 	}
	// 	result[i] = byte('0' + n.Int64())
	// }
	// return string(result), nil
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

// getClientIP 从 gRPC context 中获取客户端IP
// checkPhoneRateLimit 检查手机号限流
func (l *SendCodeLogic) checkPhoneRateLimit(phone, purpose string, maxDailySends int) error {
	if maxDailySends <= 0 {
		maxDailySends = 10 // 默认值
	}

	cfg := l.svcCtx.Config.RateLimit
	sendInterval := cfg.PhoneSendInterval
	if sendInterval <= 0 {
		sendInterval = defaultPhoneSendInterval
	}

	// 1. 检查手机号发送间隔
	phoneLockKey := fmt.Sprintf("rate:phone:lock:%s", phone)
	exists, err := l.svcCtx.Redis.Exists(phoneLockKey)
	if err != nil {
		// Redis 查询失败，记录日志但不阻止发送（可能是网络问题）
		l.Errorf("check phone lock exists failed: %v", err)
	} else if exists {
		ttl, err := l.svcCtx.Redis.Ttl(phoneLockKey)
		if err != nil {
			l.Errorf("check phone lock ttl failed: %v", err)
			// 如果查询 TTL 失败，假设锁还存在，返回错误
			return fmt.Errorf("手机号发送过于频繁，请稍后再试")
		}
		if ttl > 0 {
			return fmt.Errorf("手机号发送过于频繁，请 %d 秒后再试", ttl)
		}
	}

	// 2. 检查手机号每日发送次数（按purpose区分）
	today := time.Now().Format("20060102")
	phoneDailyKey := fmt.Sprintf("rate:phone:daily:%s:%s:%s", phone, purpose, today)
	countStr, err := l.svcCtx.Redis.Get(phoneDailyKey)
	if err != nil {
		// Redis 查询失败，记录日志但不阻止发送（可能是网络问题）
		l.Errorf("get phone daily count failed: %v", err)
	} else {
		count := 0
		if countStr != "" {
			if _, err := fmt.Sscanf(countStr, "%d", &count); err != nil {
				l.Errorf("parse phone daily count failed: %v, value: %s", err, countStr)
			}
		}
		if count >= maxDailySends {
			return fmt.Errorf("该手机号今日发送次数已达上限（%d次）", maxDailySends)
		}
	}

	return nil
}

// updateRateLimitCounters 更新限流计数器
// 注意：这些操作失败不应该影响主流程，只记录日志
func (l *SendCodeLogic) updateRateLimitCounters(phone, purpose string) {
	cfg := l.svcCtx.Config.RateLimit
	today := time.Now().Format("20060102")

	// 设置手机号发送锁
	phoneSendInterval := cfg.PhoneSendInterval
	if phoneSendInterval <= 0 {
		phoneSendInterval = defaultPhoneSendInterval
	}
	phoneLockKey := fmt.Sprintf("rate:phone:lock:%s", phone)
	if err := l.svcCtx.Redis.Setex(phoneLockKey, "1", phoneSendInterval); err != nil {
		l.Errorf("set phone lock failed: %v, key: %s", err, phoneLockKey)
	}

	// 增加手机号每日计数
	phoneDailyKey := fmt.Sprintf("rate:phone:daily:%s:%s:%s", phone, purpose, today)
	_, err := l.svcCtx.Redis.Incr(phoneDailyKey)
	if err != nil {
		l.Errorf("incr phone daily count failed: %v, key: %s", err, phoneDailyKey)
	} else {
		// 设置过期时间为当天结束
		remainingSeconds := 86400 - (time.Now().Unix() % 86400)
		if err := l.svcCtx.Redis.Expire(phoneDailyKey, int(remainingSeconds)); err != nil {
			l.Errorf("expire phone daily count failed: %v, key: %s", err, phoneDailyKey)
		}
	}
}
