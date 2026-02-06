package logic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"SLGaming/back/services/agent/agent"
	"SLGaming/back/services/agent/internal/svc"

	"github.com/cloudwego/eino-ext/components/model/ark"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
	"github.com/zeromicro/go-zero/core/logx"
)

type ModerateAvatarLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewModerateAvatarLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ModerateAvatarLogic {
	return &ModerateAvatarLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 头像多模态审核
func (l *ModerateAvatarLogic) ModerateAvatar(in *agent.ModerateAvatarRequest) (*agent.ModerateAvatarResponse, error) {
	start := time.Now()
	requestID := strings.TrimSpace(in.GetRequestId())
	scene := strings.TrimSpace(in.GetScene())
	if scene == "" {
		scene = "avatar"
	}
	userID := in.GetUserId()
	//l.Infof("ModerateAvatar start user_id=%d request_id=%s scene=%s image_url_len=%d image_base64_len=%d", userID, requestID, scene, len(in.GetImageUrl()), len(in.GetImageBase64()))
	logx.Infof("ModerateAvatar start user_id=%d request_id=%s scene=%s image_url_len=%d image_base64_len=%d", userID, requestID, scene, len(in.GetImageUrl()), len(in.GetImageBase64()))
	imageURL, err := l.resolveImageURL(in)
	if err != nil {
		l.Errorf("ModerateAvatar resolve image url failed user_id=%d request_id=%s err=%v", userID, requestID, err)
		return &agent.ModerateAvatarResponse{
			Decision:  agent.ModerationDecision_REJECT,
			RiskScore: 1,
			Labels:    []string{"invalid_input"},
			Suggestion: func() string {
				if err != nil {
					return err.Error()
				}
				return "头像不可用"
			}(),
			RequestId: in.GetRequestId(),
		}, nil
	}

	l.Infof("ModerateAvatar image resolved user_id=%d request_id=%s image_kind=%s", userID, requestID, func() string {
		if strings.HasPrefix(imageURL, "data:image") {
			return "data_url"
		}
		return "url"
	}())

	cfg := l.svcCtx.Config()
	apiKey := strings.TrimSpace(os.Getenv("ARK_API_KEY"))
	if apiKey == "" {
		apiKey = cfg.LLM.ChatAPIKey
	}
	if apiKey == "" {
		apiKey = cfg.LLM.APIKey
	}
	if apiKey == "" {
		l.Errorf("ModerateAvatar api key missing user_id=%d request_id=%s", userID, requestID)
		return &agent.ModerateAvatarResponse{
			Decision:   agent.ModerationDecision_REJECT,
			RiskScore:  1,
			Labels:     []string{"config_error"},
			Suggestion: "审核服务未配置，请稍后重试",
			RequestId:  in.GetRequestId(),
		}, nil
	}
	// 固定API Key
	apiKey = "4080e00c-d706-474e-a68c-cbee598c9001"
	model := "doubao-seed-1-6-lite-251015"

	l.Infof("ModerateAvatar calling model user_id=%d request_id=%s model=%s", userID, requestID, model)
	result, callErr := l.callModerationModel(apiKey, model, imageURL, scene)
	if callErr != nil {
		l.Errorf("ModerateAvatar model call failed user_id=%d request_id=%s scene=%s err=%v", userID, requestID, scene, callErr)
		return &agent.ModerateAvatarResponse{
			Decision:   agent.ModerationDecision_REJECT,
			RiskScore:  1,
			Labels:     []string{"model_error"},
			Suggestion: "审核失败，请更换头像或稍后重试",
			RequestId:  in.GetRequestId(),
		}, nil
	}

	l.Infof("ModerateAvatar model response received user_id=%d request_id=%s raw_decision=%s risk_score=%.2f", userID, requestID, result.Decision, result.RiskScore)

	decision := agent.ModerationDecision_REJECT
	if strings.EqualFold(result.Decision, "pass") {
		decision = agent.ModerationDecision_PASS
	}
	riskScore := clamp01(result.RiskScore)
	if riskScore == 0 && decision == agent.ModerationDecision_REJECT {
		riskScore = 0.9
	}
	if riskScore == 0 && decision == agent.ModerationDecision_PASS {
		riskScore = 0.1
	}

	suggestion := strings.TrimSpace(result.Suggestion)
	if suggestion == "" {
		if decision == agent.ModerationDecision_PASS {
			suggestion = "头像合规"
		} else {
			suggestion = "头像不合规，请更换"
		}
	}

	l.Infof("ModerateAvatar done user_id=%d request_id=%s scene=%s decision=%s risk=%.2f labels=%v duration=%s", userID, requestID, scene, decision.String(), riskScore, result.Labels, time.Since(start))

	return &agent.ModerateAvatarResponse{
		Decision:   decision,
		RiskScore:  riskScore,
		Labels:     result.Labels,
		Suggestion: suggestion,
		RequestId:  in.GetRequestId(),
	}, nil
}

type moderationResult struct {
	Decision   string   `json:"decision"`
	RiskScore  float64  `json:"risk_score"`
	Labels     []string `json:"labels"`
	Suggestion string   `json:"suggestion"`
}

func (l *ModerateAvatarLogic) resolveImageURL(in *agent.ModerateAvatarRequest) (string, error) {
	imageURL := strings.TrimSpace(in.GetImageUrl())
	imageBase64 := strings.TrimSpace(in.GetImageBase64())

	if imageURL == "" && imageBase64 == "" {
		return "", fmt.Errorf("头像图片不能为空")
	}
	if imageURL != "" {
		return imageURL, nil
	}
	if strings.HasPrefix(imageBase64, "data:image") {
		return imageBase64, nil
	}
	return "data:image/png;base64," + imageBase64, nil
}

func (l *ModerateAvatarLogic) callModerationModel(apiKey, model, imageURL, scene string) (*moderationResult, error) {
	if scene == "" {
		scene = "avatar"
	}

	timeout := 30 * time.Second
	chatModel, err := ark.NewChatModel(l.ctx, &ark.ChatModelConfig{
		APIKey:  apiKey,
		Model:   model,
		Timeout: &timeout,
	})
	if err != nil {
		return nil, fmt.Errorf("create ark chat model failed: %w", err)
	}

	moderationTask := fmt.Sprintf(`对用户头像进行合规审核，场景为：%s。
图片地址（URL 或 data URL）：%s
仅返回 JSON，不要任何额外文字，格式如下：
{"decision":"pass|reject","risk_score":0-1,"labels":["porn","violence","hate","illegal","minor","politics","copyright","watermark","other"],"suggestion":"给用户的简短提示"}
如果不确定，一律判定为 reject。`, scene, imageURL)

	template := prompt.FromMessages(schema.FString,
		schema.SystemMessage("你是一个{role}"),
		&schema.Message{
			Role:    schema.User,
			Content: "请帮帮我，{role}，{task}",
		},
	)
	params := map[string]any{
		"role": "严格的内容安全审核助手",
		"task": moderationTask,
	}
	messages, err := template.Format(l.ctx, params)
	if err != nil {
		return nil, fmt.Errorf("format prompt failed: %w", err)
	}

	resp, err := chatModel.Generate(l.ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("ark generate failed: %w", err)
	}
	content := ""
	if resp != nil {
		content = strings.TrimSpace(resp.Content)
	}
	if content == "" {
		return nil, fmt.Errorf("empty response content")
	}

	return parseModerationResult(content)
}

func parseModerationResult(content string) (*moderationResult, error) {
	var result moderationResult
	if err := json.Unmarshal([]byte(content), &result); err == nil {
		return &result, nil
	}

	start := bytes.IndexByte([]byte(content), '{')
	end := bytes.LastIndexByte([]byte(content), '}')
	if start >= 0 && end > start {
		if err := json.Unmarshal([]byte(content[start:end+1]), &result); err == nil {
			return &result, nil
		}
	}

	return nil, fmt.Errorf("parse moderation json failed, content: %s", content)
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
