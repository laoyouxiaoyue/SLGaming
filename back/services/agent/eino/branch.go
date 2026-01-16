package eino

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"
)

// RouteDecision 路由决策结构
type RouteDecision struct {
	Route     string `json:"route"`
	Reason    string `json:"reason"`
	UserInput string `json:"user_input"`
}

// newBranch branch initialization method of node 'MasterNode' in graph 'SLGaming'
// 解析 MasterNode 输出的 JSON，返回对应的路由节点
func newBranch(ctx context.Context, input string) (endNode string, err error) {
	var decision RouteDecision

	// 提取 JSON（可能包含其他文本）
	startIdx := strings.Index(input, "{")
	endIdx := strings.LastIndex(input, "}")
	if startIdx >= 0 && endIdx > startIdx {
		jsonStr := input[startIdx : endIdx+1]
		if err := json.Unmarshal([]byte(jsonStr), &decision); err == nil && decision.Route != "" {
			logx.Infof("路由决策: %s -> %s", decision.Reason, decision.Route)
			return decision.Route, nil
		}
	}

	// 降级：关键词匹配
	inputLower := strings.ToLower(input)
	if strings.Contains(inputLower, "推荐") || strings.Contains(inputLower, "查询") || strings.Contains(inputLower, "搜索") {
		return "RecommendNode", nil
	}
	if strings.Contains(inputLower, "订单") || strings.Contains(inputLower, "钱包") || strings.Contains(inputLower, "创建") {
		return "ToolsNode5", nil
	}
	if strings.Contains(inputLower, "向量") || strings.Contains(inputLower, "存储") || strings.Contains(inputLower, "更新") {
		return "AddCompanionEmbedding", nil
	}

	// 默认路由
	return "RecommendNode", nil
}

// extractUserInput 从 MasterNode 输出中提取用户原始输入
func extractUserInput(masterOutput string) string {
	var decision RouteDecision
	startIdx := strings.Index(masterOutput, "{")
	endIdx := strings.LastIndex(masterOutput, "}")
	if startIdx >= 0 && endIdx > startIdx {
		jsonStr := masterOutput[startIdx : endIdx+1]
		if err := json.Unmarshal([]byte(jsonStr), &decision); err == nil && decision.UserInput != "" {
			return decision.UserInput
		}
	}
	return masterOutput
}
