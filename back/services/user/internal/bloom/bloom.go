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

const batchSize = 500

type GenericBloomFilter[T any] struct {
	redis      *redis.Redis   // Redis 连接
	key        string         // Redis 中的 key，如 "user:id:bloom"
	toBytes    func(T) []byte // 将泛型元素转换为字节的函数
	options    BloomOptions   // 布隆过滤器配置选项
	ready      bool           // 是否初始化成功
	readyMutex sync.RWMutex   // 保护 ready 字段的读写锁
}

type BloomOptions struct {
	ExpectedElements  uint64  // 预期元素数量，决定位数组大小
	FalsePositiveRate float64 // 误判率，如 0.001 表示 0.1%
	AutoCreate        bool    // 是否自动创建（保留字段，当前未使用）
}

var DefaultBloomOptions = BloomOptions{
	ExpectedElements:  100000,
	FalsePositiveRate: 0.001,
	AutoCreate:        true,
}

// NewGenericBloomFilter 创建泛型布隆过滤器实例
// 参数：
//   - redis: Redis 连接
//   - key: Redis 中的 key 名称
//   - toBytes: 元素转字节的函数
//   - options: 可选配置，不传则使用默认值
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

	if err := bf.init(); err != nil {
		logx.Errorf("[bloom] init bloom filter %s failed: %v", bf.key, err)
	} else {
		bf.ready = true
	}

	return bf
}

// init 在 Redis 中初始化布隆过滤器
// 使用 BF.RESERVE 命令创建，如果已存在则跳过
// 等价 Redis 命令: BF.RESERVE <key> <误判率> <预期容量>
func (bf *GenericBloomFilter[T]) init() error {
	if bf.redis == nil {
		return nil
	}

	_, err := bf.redis.Eval(`
		local exists = redis.call('EXISTS', KEYS[1])
		if exists == 0 then
			return redis.call('BF.RESERVE', KEYS[1], ARGV[1], ARGV[2])
		end
		return 0
	`, []string{bf.key}, fmt.Sprintf("%f", bf.options.FalsePositiveRate), fmt.Sprintf("%d", bf.options.ExpectedElements))
	return err
}

// IsReady 检查布隆过滤器是否初始化成功
func (bf *GenericBloomFilter[T]) IsReady() bool {
	bf.readyMutex.RLock()
	defer bf.readyMutex.RUnlock()
	return bf.ready
}

// Add 添加单个元素到布隆过滤器
// 等价 Redis 命令: BF.ADD <key> <element>
// 返回值: 错误信息，nil 表示成功
func (bf *GenericBloomFilter[T]) Add(ctx context.Context, element T) error {
	if bf.redis == nil {
		return nil
	}

	data := bf.toBytes(element)
	_, err := bf.redis.Eval("return redis.call('BF.ADD', KEYS[1], ARGV[1])", []string{bf.key}, string(data))
	return err
}

// MightContain 检查元素是否可能存在
// 等价 Redis 命令: BF.EXISTS <key> <element>
// 返回值:
//   - true: 元素可能存在（有假阳性可能）
//   - false: 元素一定不存在
func (bf *GenericBloomFilter[T]) MightContain(ctx context.Context, element T) (bool, error) {
	if bf.redis == nil {
		return true, nil // Redis 未初始化时降级返回 true，走数据库查询
	}

	data := bf.toBytes(element)
	result, err := bf.redis.Eval("return redis.call('BF.EXISTS', KEYS[1], ARGV[1])", []string{bf.key}, string(data))
	if err != nil {
		return true, nil // 查询失败时降级返回 true，走数据库查询
	}

	return parseBoolResult(result)
}

// MightContainBatch 批量检查多个元素是否存在
// 参数: elements - 待检查的元素切片
// 返回: 与输入顺序对应的布尔切片，true 表示可能存在
func (bf *GenericBloomFilter[T]) MightContainBatch(ctx context.Context, elements []T) ([]bool, error) {
	if bf.redis == nil || len(elements) == 0 {
		return make([]bool, len(elements)), nil
	}

	results := make([]bool, len(elements))

	// 每 500 个元素为一批，减少 Redis 调用次数
	for i := 0; i < len(elements); i += batchSize {
		end := i + batchSize
		if end > len(elements) {
			end = len(elements)
		}
		batch := elements[i:end]

		args := make([]any, 0, len(batch))
		for _, elem := range batch {
			args = append(args, string(bf.toBytes(elem)))
		}

		// Lua 脚本：批量调用 BF.EXISTS
		luaScript := `
			local results = {}
			for j = 1, #ARGV do
				table.insert(results, redis.call('BF.EXISTS', KEYS[1], ARGV[j]))
			end
			return results
		`
		result, err := bf.redis.Eval(luaScript, []string{bf.key}, args...)
		if err != nil {
			// 失败时该批次全部标记为 true，降级走数据库
			for j := i; j < end; j++ {
				results[j] = true
			}
			continue
		}

		if arr, ok := result.([]interface{}); ok {
			for j, v := range arr {
				if b, err := parseBoolResult(v); err == nil {
					results[i+j] = b
				} else {
					results[i+j] = true
				}
			}
		}
	}

	return results, nil
}

// AddBatch 批量添加元素到布隆过滤器
// 每 500 个元素为一批，通过 Lua 脚本一次性添加，减少网络开销
// 100 万元素约需 2000 次 Redis 调用（vs 逐个添加需 100 万次）
func (bf *GenericBloomFilter[T]) AddBatch(ctx context.Context, elements []T) error {
	if bf.redis == nil || len(elements) == 0 {
		return nil
	}

	for i := 0; i < len(elements); i += batchSize {
		end := i + batchSize
		if end > len(elements) {
			end = len(elements)
		}
		batch := elements[i:end]

		args := make([]any, 0, len(batch))
		for _, elem := range batch {
			args = append(args, string(bf.toBytes(elem)))
		}

		// Lua 脚本：批量调用 BF.ADD
		luaScript := `
			local key = KEYS[1]
			for j = 1, #ARGV do
				redis.call('BF.ADD', key, ARGV[j])
			end
			return #ARGV
		`
		_, err := bf.redis.Eval(luaScript, []string{bf.key}, args...)
		if err != nil {
			return fmt.Errorf("batch add failed at index %d: %w", i, err)
		}
	}
	return nil
}

// parseBoolResult 解析 Redis 返回的布尔结果
// Redis 返回 int64 或 int 类型，1 表示 true，0 表示 false
func parseBoolResult(result interface{}) (bool, error) {
	switch v := result.(type) {
	case int64:
		return v == 1, nil
	case int:
		return v == 1, nil
	default:
		return true, fmt.Errorf("unexpected result type: %T", result)
	}
}

// Info 获取布隆过滤器的详细信息
// 等价 Redis 命令: BF.INFO <key>
// 返回: 包含 Capacity, Size, Number of filters, Number of items inserted 等信息的 map
func (bf *GenericBloomFilter[T]) Info(ctx context.Context) (map[string]interface{}, error) {
	if bf.redis == nil {
		return nil, fmt.Errorf("redis not initialized")
	}

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

	// BF.INFO 返回的是数组格式 [key1, value1, key2, value2, ...]
	// 转换为 map 格式方便使用
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

// Count 获取布隆过滤器中已插入的元素数量
// 通过 BF.INFO 获取 "Number of items inserted" 字段
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
// 等价 Redis 命令: DEL <key>
// 删除成功后会标记 ready = false
func (bf *GenericBloomFilter[T]) Delete(ctx context.Context) error {
	if bf.redis == nil {
		return nil
	}
	_, err := bf.redis.Del(bf.key)
	if err == nil {
		bf.readyMutex.Lock()
		bf.ready = false
		bf.readyMutex.Unlock()
	}
	return err
}

// Int64ToBytes 将 int64 转换为字节切片
// 用于 GenericBloomFilter[int64] 的 toBytes 参数
func Int64ToBytes(v int64) []byte {
	return []byte(strconv.FormatInt(v, 10))
}

// StringToBytes 将 string 转换为字节切片
// 用于 GenericBloomFilter[string] 的 toBytes 参数
func StringToBytes(v string) []byte {
	return []byte(v)
}

// Uint64ToBytes 将 uint64 转换为字节切片
// 用于 GenericBloomFilter[uint64] 的 toBytes 参数
func Uint64ToBytes(v uint64) []byte {
	return []byte(strconv.FormatUint(v, 10))
}

// UserBloomFilters 用户相关的布隆过滤器集合
// 包含三个独立的布隆过滤器：用户ID、手机号、UID
type UserBloomFilters struct {
	UserID *GenericBloomFilter[int64]  // 用户 ID 布隆过滤器
	Phone  *GenericBloomFilter[string] // 手机号布隆过滤器
	UID    *GenericBloomFilter[uint64] // UID 布隆过滤器

	lastSyncTime time.Time    // 上次同步时间
	syncMutex    sync.RWMutex // 保护 lastSyncTime 和 initialized
	initialized  bool         // 是否已完成数据初始化
	initMutex    sync.Mutex   // 防止并发初始化
}

// NewUserBloomFilters 创建用户布隆过滤器集合
// 参数:
//   - redis: Redis 连接
//   - db: 数据库连接，用于异步导入现有数据
func NewUserBloomFilters(redis *redis.Redis, db *gorm.DB) *UserBloomFilters {
	ubf := &UserBloomFilters{
		UserID: NewGenericBloomFilter(
			redis,
			"user:id:bloom",
			Int64ToBytes,
			BloomOptions{
				ExpectedElements:  1000000, // 预期 100 万用户
				FalsePositiveRate: 0.001,   // 0.1% 误判率
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

	// 异步从数据库导入现有数据
	if db != nil && redis != nil {
		go ubf.checkAndImportFromDB(context.Background(), db)
	}

	return ubf
}

// IsReady 检查所有布隆过滤器是否都初始化成功
func (ubf *UserBloomFilters) IsReady() bool {
	return ubf.UserID.IsReady() && ubf.Phone.IsReady() && ubf.UID.IsReady()
}

// IsInitialized 检查是否已完成数据初始化
func (ubf *UserBloomFilters) IsInitialized() bool {
	ubf.syncMutex.RLock()
	defer ubf.syncMutex.RUnlock()
	return ubf.initialized
}

// checkAndImportFromDB 检查并从数据库导入数据
// 如果布隆过滤器为空，则从数据库加载所有用户数据
// 此方法在服务启动时异步执行
func (ubf *UserBloomFilters) checkAndImportFromDB(ctx context.Context, db *gorm.DB) {
	// 加锁防止并发导入
	ubf.initMutex.Lock()
	defer ubf.initMutex.Unlock()

	// 检查是否已有数据，避免重复导入
	userIDCount, err := ubf.UserID.Count(ctx)
	if err != nil {
		logx.Errorf("[bloom] check user id bloom filter count failed: %v", err)
		return
	}

	if userIDCount > 0 {
		logx.Infof("[bloom] bloom filter already has %d items, skip import", userIDCount)
		ubf.syncMutex.Lock()
		ubf.initialized = true
		ubf.syncMutex.Unlock()
		return
	}

	logx.Info("[bloom] bloom filter is empty, importing data from database...")

	start := time.Now()

	// 从数据库加载用户数据
	// Pluck 只提取指定列，效率比 Find 高
	var userIDs []int64
	if err := db.Model(&model.User{}).Pluck("id", &userIDs).Error; err != nil {
		logx.Errorf("[bloom] load user ids from db failed: %v", err)
		return
	}

	var phones []string
	if err := db.Model(&model.User{}).Pluck("phone", &phones).Error; err != nil {
		logx.Errorf("[bloom] load phones from db failed: %v", err)
		return
	}

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

// InitFromData 从给定数据初始化布隆过滤器
// 参数:
//   - userIDs: 用户 ID 切片
//   - phones: 手机号切片
//   - uids: UID 切片
//
// 三个布隆过滤器并行导入，提高初始化速度
func (ubf *UserBloomFilters) InitFromData(
	ctx context.Context,
	userIDs []int64,
	phones []string,
	uids []uint64,
) error {
	logx.Infof("[bloom] 初始化布隆过滤器数据: users=%d, phones=%d, uids=%d",
		len(userIDs), len(phones), len(uids))

	start := time.Now()

	// 并行导入三个布隆过滤器，3 倍加速
	var wg sync.WaitGroup
	errChan := make(chan error, 3)

	wg.Add(3)
	go func() {
		defer wg.Done()
		if err := ubf.UserID.AddBatch(ctx, userIDs); err != nil {
			errChan <- fmt.Errorf("init user id bloom filter failed: %w", err)
		}
	}()
	go func() {
		defer wg.Done()
		if err := ubf.Phone.AddBatch(ctx, phones); err != nil {
			errChan <- fmt.Errorf("init phone bloom filter failed: %w", err)
		}
	}()
	go func() {
		defer wg.Done()
		if err := ubf.UID.AddBatch(ctx, uids); err != nil {
			errChan <- fmt.Errorf("init uid bloom filter failed: %w", err)
		}
	}()

	wg.Wait()
	close(errChan)

	// 检查是否有错误
	for err := range errChan {
		if err != nil {
			return err
		}
	}

	ubf.syncMutex.Lock()
	ubf.lastSyncTime = time.Now()
	ubf.initialized = true
	ubf.syncMutex.Unlock()

	logx.Infof("[bloom] 布隆过滤器初始化完成，耗时: %v", time.Since(start))
	return nil
}

// GetStats 获取布隆过滤器的统计信息
// 返回包含各过滤器元素数量、初始化状态等信息的 map
func (ubf *UserBloomFilters) GetStats(ctx context.Context) map[string]interface{} {
	stats := make(map[string]interface{})

	// 获取各过滤器的元素数量
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
	stats["initialized"] = ubf.initialized
	ubf.syncMutex.RUnlock()

	stats["ready"] = ubf.IsReady()

	return stats
}
