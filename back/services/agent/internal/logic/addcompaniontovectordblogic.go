package logic

import (
	"context"
	"fmt"
	"strconv"

	"SLGaming/back/services/agent/agent"
	"SLGaming/back/services/agent/internal/svc"

	"SLGaming/back/services/agent/internal/embedder"

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

	// 初始化自定义 DashScope 嵌入器
	emb, err := embedder.NewDashScopeEmbedder(l.ctx, cfg.LLM.APIKey, cfg.LLM.Model)
	if err != nil {
		l.Errorf("create dashscope embedder failed: %v", err)
		return &agent.AddCompanionToVectorDBResponse{
			Success: false,
			Message: fmt.Sprintf("创建 DashScope 嵌入器失败: %v", err),
		}, err
	}

	// 对描述文本进行向量化
	vectors, err := emb.EmbedStrings(l.ctx, []string{in.Description})
	if err != nil {
		l.Errorf("embedding failed: %v", err)
		return &agent.AddCompanionToVectorDBResponse{
			Success: false,
			Message: fmt.Sprintf("向量化失败: %v", err),
		}, err
	}

	if len(vectors) == 0 || len(vectors[0]) == 0 {
		return &agent.AddCompanionToVectorDBResponse{
			Success: false,
			Message: "向量化结果为空",
		}, fmt.Errorf("empty embedding result")
	}

	// 确保 Collection 存在
	if err := l.ensureCollection(); err != nil {
		l.Errorf("ensure collection failed: %v", err)
		return &agent.AddCompanionToVectorDBResponse{
			Success: false,
			Message: fmt.Sprintf("创建 Collection 失败: %v", err),
		}, err
	}

	// 将 float64 向量转换为 BinaryVector
	vectorBytes := l.float64ToBinaryVector(vectors[0])

	// 准备插入数据（id 字段由 Milvus 自动生成，不传入）
	gender := l.genderToInt16(in.Gender)
	age := int16(in.Age)
	pricePerHour := int16(in.PricePerHour)
	rating := float32(in.Rating)

	data := []entity.Column{
		entity.NewColumnBinaryVector("vector", vectorDim, [][]byte{vectorBytes}),
		entity.NewColumnInt64("companion_id", []int64{int64(in.UserId)}),
		entity.NewColumnInt16("gender", []int16{gender}),
		entity.NewColumnFloat("rating", []float32{rating}),
		entity.NewColumnInt16("age", []int16{age}),
		entity.NewColumnVarChar("game", []string{in.GameSkill}),
		entity.NewColumnVarChar("description", []string{in.Description}),
		entity.NewColumnInt16("price_per_hour", []int16{pricePerHour}),
	}

	// 插入数据（id 字段由 Milvus 自动生成）
	_, err = l.svcCtx.MilvusClient.Insert(l.ctx, collectionName, "", data...)
	if err != nil {
		l.Errorf("insert data failed: %v", err)
		return &agent.AddCompanionToVectorDBResponse{
			Success: false,
			Message: fmt.Sprintf("插入数据失败: %v", err),
		}, err
	}

	// 刷新数据，确保立即可查询
	err = l.svcCtx.MilvusClient.Flush(l.ctx, collectionName, false)
	if err != nil {
		l.Errorf("flush collection failed: %v", err)
	}

	companionIDUint := uint64(in.UserId)
	l.Infof("成功存储陪玩信息到向量数据库, companion_id=%d (user_id)", companionIDUint)

	return &agent.AddCompanionToVectorDBResponse{
		CompanionId: companionIDUint, // 返回 UserID 作为 companion_id
		Success:     true,
		Message:     "成功添加陪玩信息到向量数据库",
	}, nil
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
		// 如果 Collection 已存在，先尝试使用
		// 如果后续插入时出现字段不匹配错误，可以手动删除 Collection 重新创建
		l.Infof("Collection %s already exists", collectionName)
		return nil
	}

	// 创建 Collection Schema
	fields := l.getCollectionFields()
	schema := &entity.Schema{
		CollectionName: collectionName,
		Description:    "Companion information collection for recommendation service",
		AutoID:         true, // id 字段自动生成
		Fields:         fields,
	}

	// 创建 Collection
	err = client.CreateCollection(ctx, schema, 2)
	if err != nil {
		return fmt.Errorf("create collection failed: %w", err)
	}

	l.Infof("Collection %s created successfully", collectionName)

	// 创建索引（BinaryVector 使用 BIN_FLAT）
	// BIN_FLAT 的 nlist 参数对于 FLAT 索引会被忽略，但必须提供有效值 [1, 65536]
	index, err := entity.NewIndexBinFlat(entity.HAMMING, 1024)
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
			AutoID:     true, // id 字段自动生成
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
			DataType: entity.FieldTypeInt64, // 存储 UserID，使用 Int64
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
