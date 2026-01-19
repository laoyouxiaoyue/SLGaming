package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino-ext/components/model/qwen"
	"github.com/cloudwego/eino/schema"
	"github.com/zeromicro/go-zero/core/logx"
)

// FilterCondition 过滤条件结构
type FilterCondition struct {
	AgeMin    *int     `json:"age_min,omitempty"`
	AgeMax    *int     `json:"age_max,omitempty"`
	PriceMin  *float64 `json:"price_min,omitempty"`
	PriceMax  *float64 `json:"price_max,omitempty"`
	Gender    *string  `json:"gender,omitempty"`     // male/female/other
	Game      *string  `json:"game,omitempty"`       // 游戏名称
	RatingMin *float64 `json:"rating_min,omitempty"` // 最低评分
	Refuse    bool     `json:"refuse,omitempty"`     // 是否拒绝：当用户输入与“陪玩推荐/筛选”无关时为 true
	Reason    string   `json:"reason,omitempty"`     // 拒绝原因（给用户看的简短说明）
}

// DashScopeLLM DashScope LLM 客户端（用于千问大模型）
type DashScopeLLM struct {
	chatModel *qwen.ChatModel
	logger    logx.Logger
}

// NewDashScopeLLM 创建 DashScope LLM 客户端
func NewDashScopeLLM(ctx context.Context, apiKey, model string) (*DashScopeLLM, error) {
	if model == "" {
		model = "qwen-plus" // 默认模型
	}

	// 创建千问 ChatModel
	chatModel, err := qwen.NewChatModel(ctx, &qwen.ChatModelConfig{
		BaseURL:     "https://dashscope.aliyuncs.com/compatible-mode/v1",
		APIKey:      apiKey,
		Timeout:     0, // 使用默认超时
		Model:       model,
		MaxTokens:   of(2048),
		Temperature: of(float32(0.1)), // 降低温度，提高确定性
		TopP:        of(float32(0.7)),
	})
	if err != nil {
		return nil, fmt.Errorf("create qwen chat model failed: %w", err)
	}

	return &DashScopeLLM{
		chatModel: chatModel,
		logger:    logx.WithContext(ctx),
	}, nil
}

// of 辅助函数，用于创建指针
func of[T any](t T) *T {
	return &t
}

// FormatFilterCondition 使用千问大模型将用户输入格式化为 Milvus 过滤条件
func (l *DashScopeLLM) FormatFilterCondition(ctx context.Context, userInput string) (*FilterCondition, error) {
	prompt := fmt.Sprintf(`你是一个智能助手，你的任务是：将“陪玩推荐/筛选”相关的用户输入解析为结构化过滤条件 JSON，用于 Milvus 过滤检索。

用户输入：%s

重要规则：
1) 先判断用户输入是否与“陪玩推荐/筛选”相关（例如：找陪玩、按年龄/价格/性别/游戏/评分筛选陪玩）。
2) 如果无关（例如：闲聊、问天气、问代码、问其他业务、辱骂等），请直接返回如下 JSON（只返回 JSON，不要额外文字）：
{"refuse": true, "reason": "你的问题与陪玩推荐无关，请描述你想找什么陪玩（游戏/年龄/价格/性别等）"}
3) 如果相关，则返回过滤条件 JSON（只返回 JSON，不要额外文字），并且不要包含未提到的字段。

当输入相关时，请从用户输入中提取以下信息（如果用户没有提到，则不要包含该字段）：
- age_min: 最小年龄（整数）
- age_max: 最大年龄（整数）
- price_min: 最低价格（浮点数）
- price_max: 最高价格（浮点数）
- gender: 性别（male/female/other，如果用户提到"男"、"女"等）
- game: 游戏名称（如果用户提到具体游戏）
- rating_min: 最低评分（浮点数，0-5之间）

示例输出：
{"age_min": 20, "age_max": 30, "price_max": 100, "gender": "female"}

现在请根据用户输入提取条件：`, userInput)

	// 使用 eino-ext 的 qwen 组件调用大模型
	resp, err := l.chatModel.Generate(ctx, []*schema.Message{
		schema.UserMessage(prompt),
	})
	if err != nil {
		return nil, fmt.Errorf("qwen generate failed: %w", err)
	}

	// 获取响应内容
	content := ""
	if resp != nil {
		// resp 是 *schema.Message，直接获取 Content
		if resp.Role == schema.Assistant {
			content = resp.Content
		}
	}

	if content == "" {
		return nil, fmt.Errorf("empty response from qwen")
	}

	l.logger.Infof("LLM response: %s", content)

	// 解析 JSON
	var filter FilterCondition
	if err := json.Unmarshal([]byte(content), &filter); err != nil {
		// 尝试提取 JSON（可能 LLM 返回了其他文字）
		// 简单处理：查找第一个 { 和最后一个 }
		start := bytes.IndexByte([]byte(content), '{')
		end := bytes.LastIndexByte([]byte(content), '}')
		if start >= 0 && end > start {
			if err := json.Unmarshal([]byte(content[start:end+1]), &filter); err != nil {
				return nil, fmt.Errorf("parse json failed: %w, content: %s", err, content)
			}
		} else {
			return nil, fmt.Errorf("parse json failed: %w, content: %s", err, content)
		}
	}

	return &filter, nil
}

// BuildMilvusFilterExpr 将 FilterCondition 转换为 Milvus 过滤表达式
func BuildMilvusFilterExpr(filter *FilterCondition) string {
	if filter == nil {
		return ""
	}

	var exprs []string

	if filter.AgeMin != nil {
		exprs = append(exprs, fmt.Sprintf("age >= %d", *filter.AgeMin))
	}
	if filter.AgeMax != nil {
		exprs = append(exprs, fmt.Sprintf("age <= %d", *filter.AgeMax))
	}
	if filter.PriceMin != nil {
		exprs = append(exprs, fmt.Sprintf("price_per_hour >= %d", int(*filter.PriceMin)))
	}
	if filter.PriceMax != nil {
		exprs = append(exprs, fmt.Sprintf("price_per_hour <= %d", int(*filter.PriceMax)))
	}
	if filter.Gender != nil {
		// gender 字段是 Int16，需要转换
		genderMap := map[string]int16{
			"male":   1,
			"female": 2,
			"other":  3,
		}
		if genderVal, ok := genderMap[*filter.Gender]; ok {
			exprs = append(exprs, fmt.Sprintf("gender == %d", genderVal))
		}
	}
	if filter.Game != nil {
		exprs = append(exprs, fmt.Sprintf("game == \"%s\"", *filter.Game))
	}
	if filter.RatingMin != nil {
		exprs = append(exprs, fmt.Sprintf("rating >= %.1f", *filter.RatingMin))
	}

	if len(exprs) == 0 {
		return ""
	}

	// 用 AND 连接所有条件
	result := exprs[0]
	for i := 1; i < len(exprs); i++ {
		result += " && " + exprs[i]
	}

	return result
}
