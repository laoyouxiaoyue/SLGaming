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

type ConsumeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewConsumeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ConsumeLogic {
	return &ConsumeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ConsumeLogic) Consume(in *user.ConsumeRequest) (*user.ConsumeResponse, error) {
	userID := in.GetUserId()
	amount := in.GetAmount()
	bizOrderID := in.GetBizOrderId()

	// 参数校验
	if userID == 0 {
		l.Errorf("consume failed: invalid user_id, user_id=%d", userID)
		return nil, status.Error(codes.InvalidArgument, "user_id is required and must be greater than 0")
	}
	if amount <= 0 {
		l.Errorf("consume failed: invalid amount, user_id=%d, amount=%d", userID, amount)
		return nil, status.Error(codes.InvalidArgument, "amount must be positive")
	}

	// 记录请求开始日志
	l.Infof("consume request: user_id=%d, amount=%d, biz_order_id=%s, remark=%s",
		userID, amount, bizOrderID, in.GetRemark())

	db := l.svcCtx.DB().WithContext(l.ctx)

	var wallet model.UserWallet

	// 在事务中进行余额扣减与流水记录
	if err := db.Transaction(func(tx *gorm.DB) error {
		// 1. 幂等检查：如果提供了 BizOrderID，检查是否已经存在 CONSUME + biz_order_id 的流水
		if bizOrderID != "" {
			var existed model.WalletTransaction
			if err := tx.
				Where("type = ? AND biz_order_id = ?", "CONSUME", bizOrderID).
				First(&existed).Error; err == nil {
				// 已经消费过了，直接视为成功（幂等）
				l.Infof("consume idempotent: duplicate biz_order_id, user_id=%d, biz_order_id=%s, existing_transaction_id=%d",
					userID, bizOrderID, existed.ID)
				// 获取钱包信息用于返回
				if err := tx.Where("user_id = ?", userID).First(&wallet).Error; err != nil {
					l.Errorf("consume idempotent check failed to get wallet: user_id=%d, biz_order_id=%s, error=%v",
						userID, bizOrderID, err)
					return status.Error(codes.Internal, "failed to get wallet after idempotent check")
				}
				return nil
			} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				l.Errorf("consume idempotent check failed: user_id=%d, biz_order_id=%s, error=%v",
					userID, bizOrderID, err)
				return status.Error(codes.Internal, "failed to check idempotency")
			}
		}

		// 2. 加锁读取钱包记录，避免并发更新问题
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ?", userID).
			First(&wallet).Error

		if errors.Is(err, gorm.ErrRecordNotFound) {
			l.Errorf("consume failed: wallet not found, user_id=%d, amount=%d", userID, amount)
			return status.Error(codes.FailedPrecondition, "wallet not found, please create wallet first")
		} else if err != nil {
			l.Errorf("consume failed: database error when reading wallet, user_id=%d, error=%v", userID, err)
			return status.Error(codes.Internal, "failed to read wallet")
		}

		// 3. 余额检查
		if wallet.Balance < amount {
			l.Infof("consume failed: insufficient balance, user_id=%d, current_balance=%d, required_amount=%d, biz_order_id=%s",
				userID, wallet.Balance, amount, bizOrderID)
			return status.Error(codes.ResourceExhausted,
				"insufficient handsome coins, current balance is insufficient for this transaction")
		}

		// 4. 执行扣款
		before := wallet.Balance
		after := before - amount
		wallet.Balance = after

		if err := tx.Save(&wallet).Error; err != nil {
			l.Errorf("consume failed: failed to update wallet balance, user_id=%d, wallet_id=%d, error=%v",
				userID, wallet.ID, err)
			return status.Error(codes.Internal, "failed to update wallet balance")
		}

		// 5. 创建交易流水记录
		tr := &model.WalletTransaction{
			UserID:        userID,
			WalletID:      wallet.ID,
			ChangeAmount:  -amount, // 消费为负数
			BeforeBalance: before,
			AfterBalance:  after,
			Type:          "CONSUME",
			BizOrderID:    bizOrderID,
			Remark:        in.GetRemark(),
		}

		if err := tx.Create(tr).Error; err != nil {
			l.Errorf("consume failed: failed to create transaction record, user_id=%d, wallet_id=%d, biz_order_id=%s, error=%v",
				userID, wallet.ID, bizOrderID, err)
			return status.Error(codes.Internal, "failed to create transaction record")
		}

		// 记录成功日志
		l.Infof("consume succeeded: user_id=%d, wallet_id=%d, amount=%d, before_balance=%d, after_balance=%d, biz_order_id=%s, transaction_id=%d",
			userID, wallet.ID, amount, before, after, bizOrderID, tr.ID)

		return nil
	}); err != nil {
		// 如果是我们在事务中返回的 gRPC status error，直接透传
		if s, ok := status.FromError(err); ok {
			return nil, s.Err()
		}
		// 其他数据库错误
		l.Errorf("consume transaction failed: user_id=%d, amount=%d, biz_order_id=%s, error=%v",
			userID, amount, bizOrderID, err)
		return nil, status.Error(codes.Internal, "transaction failed, please try again later")
	}

	return &user.ConsumeResponse{
		Wallet: &user.WalletInfo{
			UserId:        wallet.UserID,
			Balance:       wallet.Balance,
			FrozenBalance: wallet.FrozenBalance,
		},
	}, nil
}
