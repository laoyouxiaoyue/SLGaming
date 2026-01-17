package logic

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"SLGaming/back/services/agent/agent"
	"SLGaming/back/services/agent/internal/svc"

	"github.com/cloudwego/eino-ext/components/embedding/ark"
	"github.com/cloudwego/eino-ext/components/indexer/milvus"
	"github.com/cloudwego/eino/schema"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
	"github.com/zeromicro/go-zero/core/logx"
)

const (
	collectionName = "companion"
	vectorDim      = 8120 // 向量维度（BinaryVector）
)

type AddCompanionToVectorDBLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAddCompanionToVectorDBLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddCompanionToVectorDBLogic {
	return &AddCompanionToVectorDBLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 添加陪玩信息到向量数据库
func (l *AddCompanionToVectorDBLogic) AddCompanionToVectorDB(in *agent.AddCompanionToVectorDBRequest) (*agent.AddCompanionToVectorDBResponse, error) {
	// 检查 Milvus 客户端是否已初始化
	if l.svcCtx.MilvusClient == nil {
		return &agent.AddCompanionToVectorDBResponse{
			Success: false,
			Message: "Milvus 客户端未初始化",
		}, fmt.Errorf("milvus client not initialized")
	}

	// 检查 LLM 配置
	cfg := l.svcCtx.Config()
	if cfg.LLM.APIKey == "" || cfg.LLM.Model == "" {
		return &agent.AddCompanionToVectorDBResponse{
			Success: false,
			Message: "LLM 配置不完整，需要 APIKey 和 Model",
		}, fmt.Errorf("llm config incomplete")
	}

	// 确保 Collection 存在
	if err := l.ensureCollection(); err != nil {
		l.Errorf("ensure collection failed: %v", err)
		return &agent.AddCompanionToVectorDBResponse{
			Success: false,
			Message: fmt.Sprintf("创建 Collection 失败: %v", err),
		}, err
	}

	// 初始化嵌入器
	timeout := 30 * time.Second
	embedder, err := ark.NewEmbedder(l.ctx, &ark.EmbeddingConfig{
		APIKey:  cfg.LLM.APIKey,
		Model:   cfg.LLM.Model,
		Timeout: &timeout,
	})
	if err != nil {
		l.Errorf("create embedder failed: %v", err)
		return &agent.AddCompanionToVectorDBResponse{
			Success: false,
			Message: fmt.Sprintf("创建嵌入器失败: %v", err),
		}, err
	}

	// 定义 Collection 字段
	fields := l.getCollectionFields()

	// 创建 Indexer
	indexer, err := milvus.NewIndexer(l.ctx, &milvus.IndexerConfig{
		Client:            l.svcCtx.MilvusClient,
		Collection:        collectionName,
		Fields:            fields,
		Embedding:         embedder,
		DocumentConverter: l.companionDocumentConverter,
	})
	if err != nil {
		l.Errorf("create indexer failed: %v", err)
		return &agent.AddCompanionToVectorDBResponse{
			Success: false,
			Message: fmt.Sprintf("创建索引器失败: %v", err),
		}, err
	}

	// 创建 Document
	doc := &schema.Document{
		ID:      strconv.FormatUint(in.UserId, 10), // 使用 user_id 作为文档 ID
		Content: in.Description,                    // 描述文本用于向量化
		MetaData: map[string]any{
			"user_id":        in.UserId,
			"companion_id":   int16(in.UserId), // 转换为 Int16
			"gender":         l.genderToInt16(in.Gender),
			"age":            int16(in.Age),
			"game":           in.GameSkill,
			"description":    in.Description,
			"price_per_hour": int16(in.PricePerHour),
			"rating":         in.Rating,
		},
	}

	// 使用 indexer.Store 存储文档
	ids, err := indexer.Store(l.ctx, []*schema.Document{doc})
	if err != nil {
		l.Errorf("store document failed: %v", err)
		return &agent.AddCompanionToVectorDBResponse{
			Success: false,
			Message: fmt.Sprintf("存储文档失败: %v", err),
		}, err
	}

	if len(ids) == 0 {
		return &agent.AddCompanionToVectorDBResponse{
			Success: false,
			Message: "存储失败，未返回 ID",
		}, fmt.Errorf("no id returned")
	}

	// 解析返回的 ID
	companionID := uint64(in.UserId)
	if idStr := ids[0]; idStr != "" {
		if id, err := strconv.ParseUint(idStr, 10, 64); err == nil {
			companionID = id
		}
	}

	l.Infof("成功存储陪玩信息到向量数据库, id=%s, companion_id=%d", ids[0], companionID)

	return &agent.AddCompanionToVectorDBResponse{
		CompanionId: companionID,
		Success:     true,
		Message:     "成功添加陪玩信息到向量数据库",
	}, nil
}

// companionDocumentConverter 将 Document 转换为 Milvus 数据格式
// 注意：indexer 会自动处理向量嵌入，这里只需要转换元数据
func (l *AddCompanionToVectorDBLogic) companionDocumentConverter(ctx context.Context, docs []*schema.Document, vectors [][]float64) ([]interface{}, error) {
	if len(docs) == 0 {
		return nil, fmt.Errorf("empty documents")
	}

	results := make([]interface{}, len(docs))
	for i, doc := range docs {
		// 解析元数据
		userID, ok := doc.MetaData["user_id"].(uint64)
		if !ok {
			if idFloat, ok := doc.MetaData["user_id"].(float64); ok {
				userID = uint64(idFloat)
			} else {
				return nil, fmt.Errorf("invalid user_id in metadata")
			}
		}

		companionID, _ := doc.MetaData["companion_id"].(int16)
		if companionID == 0 {
			if idFloat, ok := doc.MetaData["companion_id"].(float64); ok {
				companionID = int16(idFloat)
			}
		}

		gender, _ := doc.MetaData["gender"].(int16)
		if gender == 0 {
			if genderFloat, ok := doc.MetaData["gender"].(float64); ok {
				gender = int16(genderFloat)
			}
		}

		age, _ := doc.MetaData["age"].(int16)
		if age == 0 {
			if ageFloat, ok := doc.MetaData["age"].(float64); ok {
				age = int16(ageFloat)
			}
		}

		game, _ := doc.MetaData["game"].(string)
		description, _ := doc.MetaData["description"].(string)
		if description == "" {
			description = doc.Content
		}

		pricePerHour, _ := doc.MetaData["price_per_hour"].(int16)
		if pricePerHour == 0 {
			if priceFloat, ok := doc.MetaData["price_per_hour"].(float64); ok {
				pricePerHour = int16(priceFloat)
			}
		}

		rating, _ := doc.MetaData["rating"].(float64)
		if rating == 0 {
			if ratingFloat, ok := doc.MetaData["rating"].(float32); ok {
				rating = float64(ratingFloat)
			}
		}

		// 将 float64 向量转换为 BinaryVector
		var vectorBytes []byte
		if i < len(vectors) && len(vectors[i]) > 0 {
			vectorBytes = l.float64ToBinaryVector(vectors[i])
		}

		// 构建返回的数据结构（根据 Milvus 表结构）
		result := map[string]interface{}{
			"id":             int64(userID),
			"vector":         vectorBytes,
			"companion_id":   companionID,
			"gender":         gender,
			"rating":         float32(rating),
			"age":            age,
			"game":           game,
			"description":    description,
			"price_per_hour": pricePerHour,
		}

		results[i] = result
	}

	return results, nil
}

// float64ToBinaryVector 将 float64 向量转换为 BinaryVector (byte 数组)
// BinaryVector 的每个维度用 1 bit 表示，所以 8120 维需要 8120/8 = 1015 bytes
// 使用符号量化：正数为 1，负数为 0
func (l *AddCompanionToVectorDBLogic) float64ToBinaryVector(vector []float64) []byte {
	// BinaryVector 需要 (dim + 7) / 8 字节
	byteLen := (vectorDim + 7) / 8
	binaryVector := make([]byte, byteLen)

	for i := 0; i < vectorDim && i < len(vector); i++ {
		// 将 float 值转换为 bit (大于等于 0 为 1，否则为 0)
		// 或者使用符号量化：正数为 1，负数为 0
		if vector[i] >= 0 {
			byteIndex := i / 8
			bitIndex := i % 8
			binaryVector[byteIndex] |= 1 << bitIndex
		}
	}

	return binaryVector
}

// genderToInt16 将性别字符串转换为 Int16
func (l *AddCompanionToVectorDBLogic) genderToInt16(gender string) int16 {
	switch gender {
	case "male":
		return 1
	case "female":
		return 2
	case "other":
		return 3
	default:
		return 0
	}
}

// ensureCollection 确保 Collection 存在，如果不存在则创建
func (l *AddCompanionToVectorDBLogic) ensureCollection() error {
	ctx := l.ctx
	client := l.svcCtx.MilvusClient

	// 检查 Collection 是否存在
	has, err := client.HasCollection(ctx, collectionName)
	if err != nil {
		return fmt.Errorf("check collection existence failed: %w", err)
	}

	if has {
		l.Infof("Collection %s already exists", collectionName)
		return nil
	}

	// 创建 Collection Schema
	fields := l.getCollectionFields()
	schema := &entity.Schema{
		CollectionName: collectionName,
		Description:    "Companion information collection for recommendation service",
		AutoID:         false,
		Fields:         fields,
	}

	// 创建 Collection
	err = client.CreateCollection(ctx, schema, 2)
	if err != nil {
		return fmt.Errorf("create collection failed: %w", err)
	}

	l.Infof("Collection %s created successfully", collectionName)

	// 创建索引（BinaryVector 使用 BIN_FLAT 或 BIN_IVF_FLAT）
	index, err := entity.NewIndexBinFlat(entity.HAMMING, 0)
	if err != nil {
		return fmt.Errorf("create index failed: %w", err)
	}

	err = client.CreateIndex(ctx, collectionName, "vector", index, false)
	if err != nil {
		return fmt.Errorf("create index on vector failed: %w", err)
	}

	l.Infof("Index created on vector field")
	return nil
}

// getCollectionFields 获取 Collection 字段定义（根据实际 Milvus 表结构）
func (l *AddCompanionToVectorDBLogic) getCollectionFields() []*entity.Field {
	return []*entity.Field{
		{
			Name:       "id",
			DataType:   entity.FieldTypeInt64,
			PrimaryKey: true,
			AutoID:     false,
		},
		{
			Name:     "vector",
			DataType: entity.FieldTypeBinaryVector,
			TypeParams: map[string]string{
				"dim": strconv.Itoa(vectorDim),
			},
		},
		{
			Name:     "companion_id",
			DataType: entity.FieldTypeInt16,
		},
		{
			Name:     "gender",
			DataType: entity.FieldTypeInt16,
		},
		{
			Name:     "rating",
			DataType: entity.FieldTypeFloat,
		},
		{
			Name:     "age",
			DataType: entity.FieldTypeInt16,
		},
		{
			Name:     "game",
			DataType: entity.FieldTypeVarChar,
			TypeParams: map[string]string{
				"max_length": "1024",
			},
		},
		{
			Name:     "description",
			DataType: entity.FieldTypeVarChar,
			TypeParams: map[string]string{
				"max_length": "1024",
			},
		},
		{
			Name:     "price_per_hour",
			DataType: entity.FieldTypeInt16,
		},
	}
}
