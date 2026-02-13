package bloom

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"SLGaming/back/services/user/internal/model"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"gorm.io/gorm"
)

// GenericBloomFilter 泛型布隆过滤器
type GenericBloomFilter[T any] struct {
	redis    *redis.Redis
	key      string
	toBytes  func(T) []byte
	options  BloomOptions
	initOnce sync.Once // 确保只初始化一次
}

// BloomOptions 配置
type BloomOptions struct {
	ExpectedElements  uint64
	FalsePositiveRate float64
	AutoCreate        bool // 自动创建过滤器（如果不存在）
}

var DefaultBloomOptions = BloomOptions{
	ExpectedElements:  100000,
	FalsePositiveRate: 0.001,
}

// NewGenericBloomFilter 创建泛型布隆过滤器
func NewGenericBloomFilter[T any](
	redis *redis.Redis,
	key string,
	toBytes func(T) []byte,
	options ...BloomOptions,
) *GenericBloomFilter[T] {
	opts := DefaultBloomOptions
	if len(options) > 0 {
		opts = options[0]
	}

	bf := &GenericBloomFilter[T]{
		redis:   redis,
		key:     key,
		options: opts,
		toBytes: toBytes,
	}

	bf.init()
	return bf
}

// init 初始化布隆过滤器结构
func (bf *GenericBloomFilter[T]) init() {
	if bf.redis == nil {
		return
	}

	bf.redis.Eval(`
		local exists = redis.call('EXISTS', KEYS[1])
		if exists == 0 then
			return redis.call('BF.RESERVE', KEYS[1], ARGV[1], ARGV[2])
		end
		return 0
	`, []string{bf.key}, fmt.Sprintf("%f", bf.options.FalsePositiveRate), fmt.Sprintf("%d", bf.options.ExpectedElements))
}

// Add 添加元素到布隆过滤器
func (bf *GenericBloomFilter[T]) Add(ctx context.Context, element T) error {
	if bf.redis == nil {
		return nil
	}

	data := bf.toBytes(element)
	_, err := bf.redis.Eval("return redis.call('BF.ADD', KEYS[1], ARGV[1])", []string{bf.key}, string(data))
	return err
}

// MightContain 判断元素是否可能存在
func (bf *GenericBloomFilter[T]) MightContain(ctx context.Context, element T) (bool, error) {
	if bf.redis == nil {
		return true, nil
	}

	data := bf.toBytes(element)
	result, err := bf.redis.Eval("return redis.call('BF.EXISTS', KEYS[1], ARGV[1])", []string{bf.key}, string(data))
	if err != nil {
		return true, nil
	}

	switch v := result.(type) {
	case int64:
		return v == 1, nil
	case int:
		return v == 1, nil
	default:
		return true, nil
	}
}

// AddBatch 批量添加元素
func (bf *GenericBloomFilter[T]) AddBatch(ctx context.Context, elements []T) error {
	if bf.redis == nil || len(elements) == 0 {
		return nil
	}

	for _, elem := range elements {
		if err := bf.Add(ctx, elem); err != nil {
			return err
		}
	}
	return nil
}

// Info 获取布隆过滤器信息
func (bf *GenericBloomFilter[T]) Info(ctx context.Context) (map[string]interface{}, error) {
	if bf.redis == nil {
		return nil, fmt.Errorf("redis not initialized")
	}

	// 使用 pcall 调用 BF.INFO，避免脚本错误
	result, err := bf.redis.Eval(`
		local ok, res = pcall(redis.call, 'BF.INFO', KEYS[1])
		if not ok then
			return nil
		end
		return res
	`, []string{bf.key})
	if err != nil {
		return nil, err
	}

	info := make(map[string]interface{})
	if arr, ok := result.([]interface{}); ok {
		for i := 0; i < len(arr)-1; i += 2 {
			if key, ok := arr[i].(string); ok {
				info[key] = arr[i+1]
			}
		}
	}
	return info, nil
}

// Count 获取已插入元素数量（估算）
func (bf *GenericBloomFilter[T]) Count(ctx context.Context) (uint64, error) {
	info, err := bf.Info(ctx)
	if err != nil {
		return 0, err
	}

	if count, ok := info["Number of items inserted"]; ok {
		switch v := count.(type) {
		case int64:
			return uint64(v), nil
		case uint64:
			return v, nil
		}
	}
	return 0, nil
}

// Delete 删除布隆过滤器
func (bf *GenericBloomFilter[T]) Delete(ctx context.Context) error {
	if bf.redis == nil {
		return nil
	}
	_, err := bf.redis.Del(bf.key)
	return err
}

// ==================== 便捷函数 ====================

func Int64ToBytes(v int64) []byte {
	return []byte(strconv.FormatInt(v, 10))
}

func StringToBytes(v string) []byte {
	return []byte(v)
}

func Uint64ToBytes(v uint64) []byte {
	return []byte(strconv.FormatUint(v, 10))
}

// ==================== 业务封装 ====================

// UserBloomFilters 用户服务布隆过滤器
type UserBloomFilters struct {
	UserID *GenericBloomFilter[int64]
	Phone  *GenericBloomFilter[string]
	UID    *GenericBloomFilter[uint64]

	lastSyncTime time.Time
	syncMutex    sync.RWMutex
}

// NewUserBloomFilters 创建用户服务布隆过滤器
// 如果检测到布隆过滤器为空，自动从数据库导入数据
func NewUserBloomFilters(redis *redis.Redis, db *gorm.DB) *UserBloomFilters {
	ubf := &UserBloomFilters{
		UserID: NewGenericBloomFilter(
			redis,
			"user:id:bloom",
			Int64ToBytes,
			BloomOptions{
				ExpectedElements:  1000000,
				FalsePositiveRate: 0.001,
				AutoCreate:        true,
			},
		),
		Phone: NewGenericBloomFilter(
			redis,
			"user:phone:bloom",
			StringToBytes,
			BloomOptions{
				ExpectedElements:  1000000,
				FalsePositiveRate: 0.001,
				AutoCreate:        true,
			},
		),
		UID: NewGenericBloomFilter(
			redis,
			"user:uid:bloom",
			Uint64ToBytes,
			BloomOptions{
				ExpectedElements:  1000000,
				FalsePositiveRate: 0.001,
				AutoCreate:        true,
			},
		),
		lastSyncTime: time.Now(),
	}

	// 检查是否需要从数据库导入数据（首次部署或数据丢失）
	if db != nil && redis != nil {
		ubf.checkAndImportFromDB(context.Background(), db)
	}

	return ubf
}

// checkAndImportFromDB 检查布隆过滤器是否为空，如果是则从数据库导入
func (ubf *UserBloomFilters) checkAndImportFromDB(ctx context.Context, db *gorm.DB) {
	// 检查 UserID 布隆过滤器是否为空
	userIDCount, err := ubf.UserID.Count(ctx)
	if err != nil {
		logx.Errorf("[bloom] check user id bloom filter count failed: %v", err)
		return
	}

	// 如果已经有数据，说明不是首次部署，直接返回
	if userIDCount > 0 {
		logx.Infof("[bloom] bloom filter already has %d items, skip import", userIDCount)
		return
	}

	logx.Info("[bloom] bloom filter is empty, importing data from database...")

	start := time.Now()

	// 批量加载用户ID
	var userIDs []int64
	if err := db.Model(&model.User{}).Pluck("id", &userIDs).Error; err != nil {
		logx.Errorf("[bloom] load user ids from db failed: %v", err)
		return
	}

	// 批量加载手机号
	var phones []string
	if err := db.Model(&model.User{}).Pluck("phone", &phones).Error; err != nil {
		logx.Errorf("[bloom] load phones from db failed: %v", err)
		return
	}

	// 批量加载UID
	var uids []uint64
	if err := db.Model(&model.User{}).Pluck("uid", &uids).Error; err != nil {
		logx.Errorf("[bloom] load uids from db failed: %v", err)
		return
	}

	// 导入到布隆过滤器
	if err := ubf.InitFromData(ctx, userIDs, phones, uids); err != nil {
		logx.Errorf("[bloom] import data to bloom filter failed: %v", err)
		return
	}

	logx.Infof("[bloom] successfully imported %d users to bloom filter, cost: %v",
		len(userIDs), time.Since(start))
}

// InitFromData 从数据初始化所有布隆过滤器
func (ubf *UserBloomFilters) InitFromData(
	ctx context.Context,
	userIDs []int64,
	phones []string,
	uids []uint64,
) error {
	logx.Infof("[bloom] 初始化布隆过滤器数据: users=%d, phones=%d, uids=%d",
		len(userIDs), len(phones), len(uids))

	start := time.Now()

	if err := ubf.UserID.AddBatch(ctx, userIDs); err != nil {
		return fmt.Errorf("init user id bloom filter failed: %w", err)
	}

	if err := ubf.Phone.AddBatch(ctx, phones); err != nil {
		return fmt.Errorf("init phone bloom filter failed: %w", err)
	}

	if err := ubf.UID.AddBatch(ctx, uids); err != nil {
		return fmt.Errorf("init uid bloom filter failed: %w", err)
	}

	ubf.syncMutex.Lock()
	ubf.lastSyncTime = time.Now()
	ubf.syncMutex.Unlock()

	logx.Infof("[bloom] 布隆过滤器初始化完成，耗时: %v", time.Since(start))
	return nil
}

// GetStats 获取统计信息
func (ubf *UserBloomFilters) GetStats(ctx context.Context) map[string]interface{} {
	stats := make(map[string]interface{})

	if count, err := ubf.UserID.Count(ctx); err == nil {
		stats["user_id_count"] = count
	}
	if count, err := ubf.Phone.Count(ctx); err == nil {
		stats["phone_count"] = count
	}
	if count, err := ubf.UID.Count(ctx); err == nil {
		stats["uid_count"] = count
	}

	ubf.syncMutex.RLock()
	stats["last_sync_time"] = ubf.lastSyncTime
	ubf.syncMutex.RUnlock()

	return stats
}

// OrderBloomFilters 订单服务布隆过滤器
type OrderBloomFilters struct {
	OrderNo *GenericBloomFilter[string]
	UserID  *GenericBloomFilter[int64]
}

func NewOrderBloomFilters(redis *redis.Redis) *OrderBloomFilters {
	return &OrderBloomFilters{
		OrderNo: NewGenericBloomFilter(
			redis,
			"order:no:bloom",
			StringToBytes,
			BloomOptions{
				ExpectedElements:  10000000,
				FalsePositiveRate: 0.0001,
				AutoCreate:        true,
			},
		),
		UserID: NewGenericBloomFilter(
			redis,
			"order:userid:bloom",
			Int64ToBytes,
			BloomOptions{
				ExpectedElements:  1000000,
				FalsePositiveRate: 0.001,
				AutoCreate:        true,
			},
		),
	}
}
