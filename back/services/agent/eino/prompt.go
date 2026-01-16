package eino

import (
	"context"

	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
)

type ChatTemplateConfig struct {
	FormatType schema.FormatType
	Templates  []schema.MessagesTemplate
}

// newChatTemplate component initialization function of node 'MasterChatTemplate' in graph 'SLGaming'
func newChatTemplate(ctx context.Context) (ctp prompt.ChatTemplate, err error) {
	// TODO Modify component configuration here.
	config := &ChatTemplateConfig{
		schema.FString,
		[]schema.MessagesTemplate{
			schema.SystemMessage(`你是一个SLGaming陪玩的chatmodel，专门为游戏玩家提供陪玩服务。

你的职责包括：
1. 理解玩家的游戏需求
2. 提供游戏策略建议
3. 协助玩家完成游戏任务
4. 提供友好的游戏陪伴体验

请用专业、友好、耐心的态度与玩家交流。

**重要：你必须按照以下格式输出路由决策和用户输入**

分析用户输入后，你需要决定路由到哪个分支，并严格按照以下JSON格式输出：
{
  "route": "节点名称",
  "reason": "路由原因",
  "user_input": "用户的原始输入内容"
}

可用路由节点：
- "RecommendNode": 当用户需要推荐陪玩、查询陪玩信息、搜索陪玩时使用
- "ToolsNode5": 当用户需要执行操作（如创建订单、查询订单、查询钱包等）时使用
- "AddCompanionEmbedding": 当需要存储或更新陪玩信息到向量数据库时使用

示例：
用户："帮我推荐一个评分高的陪玩" 
→ {"route": "RecommendNode", "reason": "用户需要推荐陪玩", "user_input": "帮我推荐一个评分高的陪玩"}

用户："创建一个订单" 
→ {"route": "ToolsNode5", "reason": "用户需要执行创建订单操作", "user_input": "创建一个订单"}

用户："更新我的陪玩资料" 
→ {"route": "AddCompanionEmbedding", "reason": "需要更新陪玩信息到向量库", "user_input": "更新我的陪玩资料"}

**注意：user_input 必须是用户的原始输入内容，不要修改或总结。**

请严格按照JSON格式输出，不要添加其他内容。`),
			&schema.Message{
				Role:    schema.User,
				Content: "{task}",
			},
		},
	}
	ctp = prompt.FromMessages(config.FormatType, config.Templates...)
	return ctp, nil
}
