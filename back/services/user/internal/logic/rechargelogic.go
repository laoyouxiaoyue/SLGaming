package logic

import (
	"context"
	"errors"

	"SLGaming/back/services/user/internal/model"
	"SLGaming/back/services/user/internal/svc"
	"SLGaming/back/services/user/user"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

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

	db := l.svcCtx.DB().WithContext(l.ctx)

	var wallet model.UserWallet

	// 在事务中进行余额更新与流水记录
	if err := db.Transaction(func(tx *gorm.DB) error {
		// 1. 幂等检查：如果提供了 BizOrderID，检查是否已经存在 RECHARGE + biz_order_id 的流水
		if in.GetBizOrderId() != "" {
			var existed model.WalletTransaction
			if err := tx.
				Where("type = ? AND biz_order_id = ?", "RECHARGE", in.GetBizOrderId()).
				First(&existed).Error; err == nil {
				// 已经充值过了，直接视为成功（幂等）
				// 获取钱包信息用于返回
				if err := tx.Where("user_id = ?", userID).First(&wallet).Error; err != nil {
					return err
				}
				return nil
			} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
		}

		// 2. 加锁读取钱包记录，避免并发更新问题
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ?", userID).
			First(&wallet).Error

		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 没有钱包记录时，为用户创建一个新的钱包
			wallet = model.UserWallet{
				UserID:        userID,
				Balance:       0,
				FrozenBalance: 0,
			}
			if err := tx.Create(&wallet).Error; err != nil {
				return err
			}
		} else if err != nil {
			return err
		}

		before := wallet.Balance
		after := before + amount
		wallet.Balance = after

		if err := tx.Save(&wallet).Error; err != nil {
			return err
		}

		tr := &model.WalletTransaction{
			UserID:        userID,
			WalletID:      wallet.ID,
			ChangeAmount:  amount,
			BeforeBalance: before,
			AfterBalance:  after,
			Type:          "RECHARGE",
			BizOrderID:    in.GetBizOrderId(),
			Remark:        in.GetRemark(),
		}

		if err := tx.Create(tr).Error; err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &user.RechargeResponse{
		Wallet: &user.WalletInfo{
			UserId:        wallet.UserID,
			Balance:       wallet.Balance,
			FrozenBalance: wallet.FrozenBalance,
		},
	}, nil
}
