package logic

import (
	"context"
	"fmt"
	"time"

	"SLGaming/back/services/agent/agent"
	"SLGaming/back/services/agent/internal/embedder"
	"SLGaming/back/services/agent/internal/llm"
	"SLGaming/back/services/agent/internal/svc"

	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
	"github.com/zeromicro/go-zero/core/logx"
)

const (
	recommendTopK = 10 // 返回前10个结果
)

type RecommendCompanionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRecommendCompanionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RecommendCompanionLogic {
	return &RecommendCompanionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 根据用户输入推荐陪玩
func (l *RecommendCompanionLogic) RecommendCompanion(in *agent.RecommendCompanionRequest) (*agent.RecommendCompanionResponse, error) {
	start := time.Now()
	userID := in.GetUserId()
	userInput := in.GetUserInput()

	l.Infof("RecommendCompanion start user_id=%d user_input=%.150s", userID, userInput)

	// 检查 Milvus 客户端是否已初始化
	if l.svcCtx.MilvusClient == nil {
		l.Errorf("RecommendCompanion milvus client not initialized user_id=%d", userID)
		return &agent.RecommendCompanionResponse{
			Companions:  nil,
			Explanation: "Milvus 客户端未初始化",
		}, fmt.Errorf("milvus client not initialized")
	}

	// 检查 LLM 配置
	cfg := l.svcCtx.Config()
	if cfg.LLM.APIKey == "" {
		l.Errorf("RecommendCompanion llm config incomplete user_id=%d", userID)
		return &agent.RecommendCompanionResponse{
			Companions:  nil,
			Explanation: "LLM 配置不完整",
		}, fmt.Errorf("llm config incomplete")
	}
	l.Infof("RecommendCompanion llm config loaded user_id=%d model=%s", userID, cfg.LLM.Model)

	// 1. 使用千问大模型格式化过滤条件
	chatAPIKey := cfg.LLM.ChatAPIKey
	if chatAPIKey == "" {
		chatAPIKey = cfg.LLM.APIKey
	}
	chatModel := cfg.LLM.ChatModel
	if chatModel == "" {
		chatModel = "qwen-plus"
	}

	var filter *llm.FilterCondition
	llmClient, err := llm.NewDashScopeLLM(l.ctx, chatAPIKey, chatModel)
	if err != nil {
		l.Errorf("create dashscope llm failed: %v", err)
		// 如果 LLM 创建失败，继续使用空过滤条件
		filter = nil
	} else {
		filter, err = llmClient.FormatFilterCondition(l.ctx, in.UserInput)
		if err != nil {
			l.Errorf("format filter condition failed: %v", err)
			// 如果 LLM 格式化失败，继续使用空过滤条件
			filter = nil
		}
	}

	// 如果输入与陪玩推荐无关，直接拒绝（不进行向量化、不访问 Milvus）
	if filter != nil && filter.Refuse {
		reason := filter.Reason
		if reason == "" {
			reason = "你的问题与陪玩推荐无关，请描述你想找什么陪玩（游戏/年龄/价格/性别等）"
		}
		return &agent.RecommendCompanionResponse{
			Companions:  nil,
			Explanation: reason,
		}, nil
	}

	// 构建 Milvus 过滤表达式
	filterExpr := llm.BuildMilvusFilterExpr(filter)
	if filterExpr != "" {
		l.Infof("使用过滤条件: %s", filterExpr)
	}

	// 2. 对用户输入进行向量化
	emb, err := embedder.NewDashScopeEmbedder(l.ctx, cfg.LLM.APIKey, cfg.LLM.Model)
	if err != nil {
		l.Errorf("create embedder failed: %v", err)
		return &agent.RecommendCompanionResponse{
			Companions:  nil,
			Explanation: fmt.Sprintf("创建嵌入器失败: %v", err),
		}, err
	}

	vectors, err := emb.EmbedStrings(l.ctx, []string{in.UserInput})
	if err != nil {
		l.Errorf("embedding failed: %v", err)
		return &agent.RecommendCompanionResponse{
			Companions:  nil,
			Explanation: fmt.Sprintf("向量化失败: %v", err),
		}, err
	}

	if len(vectors) == 0 || len(vectors[0]) == 0 {
		return &agent.RecommendCompanionResponse{
			Companions:  nil,
			Explanation: "向量化结果为空",
		}, fmt.Errorf("empty embedding result")
	}

	// 3. 将 float64 向量转换为 BinaryVector
	vectorDim := 8120 // 与 addcompaniontovectordblogic.go 保持一致
	vectorBytes := l.float64ToBinaryVector(vectors[0], vectorDim)

	// 4. 使用 Milvus 进行向量检索（带过滤条件）
	collectionName := "companion"

	// 检查 Collection 是否存在
	has, err := l.svcCtx.MilvusClient.HasCollection(l.ctx, collectionName)
	if err != nil {
		l.Errorf("check collection existence failed: %v", err)
		return &agent.RecommendCompanionResponse{
			Companions:  nil,
			Explanation: fmt.Sprintf("检查 Collection 失败: %v", err),
		}, err
	}
	if !has {
		return &agent.RecommendCompanionResponse{
			Companions:  nil,
			Explanation: "Collection 不存在，请先添加陪玩数据",
		}, fmt.Errorf("collection %s does not exist", collectionName)
	}

	// 创建 BinaryVector
	binaryVector := entity.BinaryVector(vectorBytes)

	// 创建 SearchParam（对于 BIN_FLAT 索引，使用 BinFlatSearchParam）
	// 对于 FLAT 索引，nprobe 参数会被忽略，但必须提供有效值 [1, 65536]
	searchParam, err := entity.NewIndexBinFlatSearchParam(10) // nprobe 参数，对于 FLAT 索引会被忽略，使用默认值 10
	if err != nil {
		l.Errorf("create search param failed: %v", err)
		return &agent.RecommendCompanionResponse{
			Companions:  nil,
			Explanation: fmt.Sprintf("创建搜索参数失败: %v", err),
		}, err
	}

	searchResult, err := l.svcCtx.MilvusClient.Search(
		l.ctx,
		collectionName,
		[]string{}, // partitions
		filterExpr, // expr: 过滤表达式
		[]string{"companion_id", "gender", "age", "game", "description", "price_per_hour", "rating"}, // output fields
		[]entity.Vector{binaryVector}, // vectors
		"vector",                      // vector field name
		entity.HAMMING,                // metric type
		recommendTopK,                 // topK
		searchParam,                   // search params
	)
	if err != nil {
		l.Errorf("milvus search failed: %v", err)
		return &agent.RecommendCompanionResponse{
			Companions:  nil,
			Explanation: fmt.Sprintf("检索失败: %v", err),
		}, err
	}

	// 5. 解析检索结果
	var companions []*agent.CompanionRecommendation
	if len(searchResult) > 0 {
		result := searchResult[0]
		// 遍历所有结果
		for i := 0; i < result.ResultCount; i++ {
			// 获取字段值（按索引获取）
			companionID := l.getFieldValueInt64ByIndex(result.Fields, "companion_id", i)
			genderVal := l.getFieldValueInt16ByIndex(result.Fields, "gender", i)
			age := l.getFieldValueInt16ByIndex(result.Fields, "age", i)
			pricePerHour := l.getFieldValueInt16ByIndex(result.Fields, "price_per_hour", i)
			rating := l.getFieldValueFloatByIndex(result.Fields, "rating", i)
			game := l.getFieldValueStringByIndex(result.Fields, "game", i)
			description := l.getFieldValueStringByIndex(result.Fields, "description", i)

			// 转换性别
			genderStr := l.int16ToGender(genderVal)

			// 计算相似度（Hamming 距离转换为相似度，距离越小相似度越高）
			similarity := 1.0
			if i < len(result.Scores) {
				score := result.Scores[i]
				// Hamming 距离范围是 0 到 vectorDim，转换为相似度
				similarity = 1.0 - float64(score)/float64(vectorDim)
				if similarity < 0 {
					similarity = 0
				}
			}

			companions = append(companions, &agent.CompanionRecommendation{
				UserId:       uint64(companionID),
				GameSkill:    game,
				Gender:       genderStr,
				Age:          int32(age),
				Description:  description,
				PricePerHour: int64(pricePerHour),
				Rating:       float64(rating),
				Similarity:   similarity,
			})
		}
	}

	explanation := fmt.Sprintf("根据您的需求\"%s\"，为您推荐了 %d 位陪玩", in.UserInput, len(companions))
	if filterExpr != "" {
		explanation += fmt.Sprintf("（已应用过滤条件：%s）", filterExpr)
	}

	l.Infof("RecommendCompanion done user_id=%d companion_count=%d filter_expr=%s duration=%s", in.GetUserId(), len(companions), filterExpr, time.Since(start))

	return &agent.RecommendCompanionResponse{
		Companions:  companions,
		Explanation: explanation,
	}, nil
}

// float64ToBinaryVector 将 float64 向量转换为 BinaryVector
func (l *RecommendCompanionLogic) float64ToBinaryVector(vector []float64, vectorDim int) []byte {
	byteLen := (vectorDim + 7) / 8
	binaryVector := make([]byte, byteLen)

	for i := 0; i < vectorDim && i < len(vector); i++ {
		if vector[i] >= 0 {
			byteIndex := i / 8
			bitIndex := i % 8
			binaryVector[byteIndex] |= 1 << bitIndex
		}
	}

	return binaryVector
}

// int16ToGender 将 Int16 转换为性别字符串
func (l *RecommendCompanionLogic) int16ToGender(gender int16) string {
	switch gender {
	case 1:
		return "male"
	case 2:
		return "female"
	case 3:
		return "other"
	default:
		return "unknown"
	}
}

// getFieldValueInt64ByIndex 从搜索结果中按索引获取 Int64 字段值
func (l *RecommendCompanionLogic) getFieldValueInt64ByIndex(fields client.ResultSet, fieldName string, idx int) int64 {
	for _, field := range fields {
		if field.Name() == fieldName {
			if col, ok := field.(*entity.ColumnInt64); ok && idx < col.Len() {
				return col.Data()[idx]
			}
		}
	}
	return 0
}

// getFieldValueInt16ByIndex 从搜索结果中按索引获取 Int16 字段值
func (l *RecommendCompanionLogic) getFieldValueInt16ByIndex(fields client.ResultSet, fieldName string, idx int) int16 {
	for _, field := range fields {
		if field.Name() == fieldName {
			if col, ok := field.(*entity.ColumnInt16); ok && idx < col.Len() {
				return col.Data()[idx]
			}
		}
	}
	return 0
}

// getFieldValueFloatByIndex 从搜索结果中按索引获取 Float 字段值
func (l *RecommendCompanionLogic) getFieldValueFloatByIndex(fields client.ResultSet, fieldName string, idx int) float32 {
	for _, field := range fields {
		if field.Name() == fieldName {
			if col, ok := field.(*entity.ColumnFloat); ok && idx < col.Len() {
				return col.Data()[idx]
			}
		}
	}
	return 0
}

// getFieldValueStringByIndex 从搜索结果中按索引获取 String 字段值
func (l *RecommendCompanionLogic) getFieldValueStringByIndex(fields client.ResultSet, fieldName string, idx int) string {
	for _, field := range fields {
		if field.Name() == fieldName {
			if col, ok := field.(*entity.ColumnVarChar); ok && idx < col.Len() {
				return col.Data()[idx]
			}
		}
	}
	return ""
}
