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
	imageURL, err := l.resolveImageURL(in)
	if err != nil {
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

	cfg := l.svcCtx.Config()
	apiKey := strings.TrimSpace(os.Getenv("ARK_API_KEY"))
	if apiKey == "" {
		apiKey = cfg.LLM.ChatAPIKey
	}
	if apiKey == "" {
		apiKey = cfg.LLM.APIKey
	}
	if apiKey == "" {
		return &agent.ModerateAvatarResponse{
			Decision:   agent.ModerationDecision_REJECT,
			RiskScore:  1,
			Labels:     []string{"config_error"},
			Suggestion: "审核服务未配置，请稍后重试",
			RequestId:  in.GetRequestId(),
		}, nil
	}
	apiKey = "4080e00c-d706-474e-a68c-cbee598c9001"
	model := strings.TrimSpace(cfg.LLM.ChatModel)
	model = "doubao-seed-1-6-lite-251015"

	result, callErr := l.callModerationModel(apiKey, model, imageURL, in.GetScene())
	if callErr != nil {
		l.Errorf("moderation model call failed: %v", callErr)
		return &agent.ModerateAvatarResponse{
			Decision:   agent.ModerationDecision_REJECT,
			RiskScore:  1,
			Labels:     []string{"model_error"},
			Suggestion: "审核失败，请更换头像或稍后重试",
			RequestId:  in.GetRequestId(),
		}, nil
	}

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
