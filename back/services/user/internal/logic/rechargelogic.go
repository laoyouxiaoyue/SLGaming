package logic

import (
	"context"
	"errors"
	"strconv"
	"time"

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
const userCachePrefix2 = "user:info:"

// getUserCacheKey 获取用户缓存键
func getUserCacheKey2(userID int64) string {
	return userCachePrefix2 + strconv.FormatInt(userID, 10)
}

type RechargeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRechargeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RechargeLogic {
	return &RechargeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RechargeLogic) Recharge(in *user.RechargeRequest) (*user.RechargeResponse, error) {
	userID := in.GetUserId()
	amount := in.GetAmount()

	if userID == 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	if amount <= 0 {
		return nil, status.Error(codes.InvalidArgument, "amount must be positive")
	}

	// 使用钱包服务统一处理
	walletService := helper.NewWalletService(l.svcCtx.DB())
	result, err := walletService.UpdateBalance(l.ctx, &helper.WalletUpdateRequest{
		UserID:     userID,
		Amount:     amount,
		Type:       helper.WalletOpRecharge,
		BizOrderID: in.GetBizOrderId(),
		Remark:     in.GetRemark(),
		Logger:     l.Logger,
	})
	if err != nil {
		return nil, err
	}

	// 落库充值订单（成功态）。若重复回调，做幂等更新。
	if in.GetBizOrderId() != "" {
		now := time.Now()
		db := l.svcCtx.DB()
		if db != nil {
			var existing model.RechargeOrder
			err := db.Where("order_no = ?", in.GetBizOrderId()).First(&existing).Error
			switch {
			case err == nil:
				updates := map[string]interface{}{
					"user_id":  userID,
					"amount":   amount,
					"status":   model.RechargeStatusSuccess,
					"remark":   in.GetRemark(),
					"paid_at":  &now,
					"pay_type": "alipay",
				}
				if err := db.Model(&existing).Updates(updates).Error; err != nil {
					l.Logger.Errorf("update recharge order failed: %v", err)
				}
			case errors.Is(err, gorm.ErrRecordNotFound):
				order := &model.RechargeOrder{
					UserID:  userID,
					OrderNo: in.GetBizOrderId(),
					Amount:  amount,
					Status:  model.RechargeStatusSuccess,
					PayType: "alipay",
					Remark:  in.GetRemark(),
					PaidAt:  &now,
				}
				if err := db.Create(order).Error; err != nil {
					l.Logger.Errorf("create recharge order failed: %v", err)
				}
			default:
				l.Logger.Errorf("query recharge order failed: %v", err)
			}
		}
	}

	// 清除用户缓存，确保余额立即更新
	if l.svcCtx.Redis != nil {
		cacheKey := getUserCacheKey2(int64(userID))
		if _, err := l.svcCtx.Redis.Del(cacheKey); err != nil {
			l.Logger.Errorf("delete user cache failed: %v", err)
		} else {
			l.Logger.Infof("user cache deleted successfully: %s", cacheKey)
		}
	}

	return &user.RechargeResponse{
		Wallet: result.ToWalletInfo(),
	}, nil
}
