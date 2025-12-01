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

	if userID == 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	if amount <= 0 {
		return nil, status.Error(codes.InvalidArgument, "amount must be positive")
	}

	db := l.svcCtx.DB().WithContext(l.ctx)

	var wallet model.UserWallet

	// 在事务中进行余额扣减与流水记录
	if err := db.Transaction(func(tx *gorm.DB) error {
		// 加锁读取钱包记录，避免并发更新问题
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ?", userID).
			First(&wallet).Error

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return status.Error(codes.FailedPrecondition, "wallet not found")
		} else if err != nil {
			return err
		}

		if wallet.Balance < amount {
			return status.Error(codes.ResourceExhausted, "insufficient handsome coins")
		}

		before := wallet.Balance
		after := before - amount
		wallet.Balance = after

		if err := tx.Save(&wallet).Error; err != nil {
			return err
		}

		tr := &model.WalletTransaction{
			UserID:        userID,
			WalletID:      wallet.ID,
			ChangeAmount:  -amount, // 消费为负数
			BeforeBalance: before,
			AfterBalance:  after,
			Type:          "CONSUME",
			BizOrderID:    in.GetBizOrderId(),
			Remark:        in.GetRemark(),
		}

		if err := tx.Create(tr).Error; err != nil {
			return err
		}

		return nil
	}); err != nil {
		// 如果是我们在事务中返回的 gRPC status error，直接透传
		if s, ok := status.FromError(err); ok {
			return nil, s.Err()
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &user.ConsumeResponse{
		Wallet: &user.WalletInfo{
			UserId:        wallet.UserID,
			Balance:       wallet.Balance,
			FrozenBalance: wallet.FrozenBalance,
		},
	}, nil
}
