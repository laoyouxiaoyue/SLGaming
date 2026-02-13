package logic

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"

	"SLGaming/back/services/user/internal/helper"
	"SLGaming/back/services/user/internal/model"
	"SLGaming/back/services/user/internal/svc"
	"SLGaming/back/services/user/user"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

// 缓存相关常量
const (
	UserCachePrefix = "user:info:"
	EmptyCacheTTL   = 300  // 空值缓存5分钟
	UserCacheTTL    = 1800 // 用户缓存30分钟
)

// GetUserCacheKey 获取用户缓存键
func GetUserCacheKey(userID int64) string {
	return UserCachePrefix + strconv.FormatInt(userID, 10)
}

type GetUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserLogic {
	return &GetUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetUserLogic) GetUser(in *user.GetUserRequest) (*user.GetUserResponse, error) {
	// 处理ID查询的缓存和布隆过滤器逻辑
	if in.GetId() != 0 {
		return l.getUserById(int64(in.GetId()))
	}

	// 其他查询条件（uid、phone）走原有逻辑
	db := l.svcCtx.DB().WithContext(l.ctx)
	var u model.User
	var err error

	switch {
	case in.GetUid() != 0:
		err = db.Where("uid = ?", in.GetUid()).First(&u).Error
	case in.GetPhone() != "":
		err = db.Where("phone = ?", in.GetPhone()).First(&u).Error
	default:
		return nil, status.Error(codes.InvalidArgument, "missing query condition")
	}

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	// 获取用户信息
	userInfo := helper.ToUserInfo(&u)

	// 从数据库获取的用户信息已经包含最新的粉丝数和关注数
	// 不需要再从Redis缓存获取，避免覆盖最新值

	// 获取钱包信息
	l.getUserWalletInfo(&u, userInfo)

	return &user.GetUserResponse{
		User: userInfo,
	}, nil
}

// getUserById 根据ID获取用户，集成布隆过滤器和缓存
func (l *GetUserLogic) getUserById(userID int64) (*user.GetUserResponse, error) {
	// 步骤1：布隆过滤器判断（如果配置了）
	if l.svcCtx.BloomFilter != nil {
		exists, err := l.svcCtx.BloomFilter.UserID.MightContain(l.ctx, userID)
		if err != nil {
			l.Errorf("bloom filter check failed: %v", err)
			// 布隆过滤器查询失败，降级到正常流程
		} else if !exists {
			// 用户ID肯定不存在，直接返回
			return nil, status.Error(codes.NotFound, "user not found")
		}
	}

	// 步骤2：尝试从缓存获取
	cacheKey := GetUserCacheKey(userID)
	var cachedUser user.UserInfo

	if l.svcCtx.Redis != nil {
		cacheData, err := l.svcCtx.Redis.Get(cacheKey)
		if err == nil && cacheData != "" {
			if err := json.Unmarshal([]byte(cacheData), &cachedUser); err == nil {
				// 从计数缓存获取最新的粉丝数和关注数
				if l.svcCtx.UserCache != nil {
					if followerCount, err := l.svcCtx.UserCache.GetFollowerCount(userID); err == nil && followerCount > 0 {
						cachedUser.FollowerCount = followerCount
					}
					if followingCount, err := l.svcCtx.UserCache.GetFollowingCount(userID); err == nil && followingCount > 0 {
						cachedUser.FollowingCount = followingCount
					}
				}
				return &user.GetUserResponse{
					User: &cachedUser,
				}, nil
			}
		}
	}

	// 步骤3：缓存未命中，查询数据库
	db := l.svcCtx.DB().WithContext(l.ctx)
	var u model.User
	err := db.Where("id = ?", userID).First(&u).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 数据库未命中，缓存空值
			if l.svcCtx.Redis != nil {
				if err := l.svcCtx.Redis.Setex(cacheKey, "", EmptyCacheTTL); err != nil {
					l.Errorf("set empty user cache failed: %v", err)
				}
			}
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	// 步骤4：数据库命中，构建响应
	userInfo := helper.ToUserInfo(&u)

	// 从数据库获取的用户信息已经包含最新的粉丝数和关注数
	// 不需要再从Redis缓存获取，避免覆盖最新值

	// 获取钱包信息
	l.getUserWalletInfo(&u, userInfo)

	// 步骤5：更新缓存
	if l.svcCtx.Redis != nil {
		userInfoJSON, err := json.Marshal(userInfo)
		if err == nil {
			if err := l.svcCtx.Redis.Setex(cacheKey, string(userInfoJSON), UserCacheTTL); err != nil {
				l.Errorf("set user cache failed: %v", err)
			}
		}
	}

	return &user.GetUserResponse{
		User: userInfo,
	}, nil
}

// getUserWalletInfo 获取用户钱包信息
func (l *GetUserLogic) getUserWalletInfo(u *model.User, userInfo *user.UserInfo) {
	db := l.svcCtx.DB().WithContext(l.ctx)
	var wallet model.UserWallet
	walletErr := db.Where("user_id = ?", u.ID).First(&wallet).Error
	if errors.Is(walletErr, gorm.ErrRecordNotFound) {
		// 如果钱包不存在，使用默认值（余额为0）
		userInfo.Balance = 0
		userInfo.FrozenBalance = 0
	} else if walletErr != nil {
		// 如果查询钱包出错，记录日志但不影响用户信息返回
		l.Errorf("failed to get wallet for user %d: %v", u.ID, walletErr)
		userInfo.Balance = 0
		userInfo.FrozenBalance = 0
	} else {
		// 成功获取钱包信息
		userInfo.Balance = wallet.Balance
		userInfo.FrozenBalance = wallet.FrozenBalance
	}
}

// LogSuccess 记录成功日志的辅助函数
func LogSuccess(l logx.Logger, operation string, fields map[string]interface{}) {
	l.Infof("[%s] success: %v", operation, fields)
}

// LogError 记录错误日志的辅助函数
func LogError(l logx.Logger, operation string, message string, err error, fields map[string]interface{}) {
	if fields == nil {
		fields = make(map[string]interface{})
	}
	fields["error"] = err
	l.Errorf("[%s] %s: %v", operation, message, fields)
}

// LogWarning 记录警告日志的辅助函数
func LogWarning(l logx.Logger, operation string, message string, fields map[string]interface{}) {
	if fields == nil {
		fields = make(map[string]interface{})
	}
	l.Infof("[WARN][%s] %s: %v", operation, message, fields)
}
