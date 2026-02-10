package bloom

import (
	"SLGaming/back/services/user/internal/model"
	"SLGaming/back/services/user/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

// InitBloomFilter 初始化布隆过滤器，加载所有用户ID
func InitBloomFilter(svcCtx *svc.ServiceContext) error {
	if svcCtx.Redis == nil {
		logx.Info("Redis not initialized, skip bloom filter initialization")
		return nil
	}

	logx.Info("Initializing bloom filter with all user IDs...")

	// 从数据库获取所有用户ID
	db := svcCtx.DB()
	var userIDs []int64

	err := db.Model(&model.User{}).Select("id").Find(&userIDs).Error
	if err != nil {
		logx.Errorf("failed to get all user IDs: %v", err)
		return err
	}

	logx.Infof("Found %d users, adding to bloom filter...", len(userIDs))

	// 创建布隆过滤器实例并初始化
	bf := NewBloomFilter(svcCtx)
	err = bf.InitUserBloomFilter(userIDs)
	if err != nil {
		logx.Errorf("failed to initialize bloom filter: %v", err)
		return err
	}

	logx.Info("Bloom filter initialized successfully")
	return nil
}
