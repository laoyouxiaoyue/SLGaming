package logic

import (
	"context"

	"SLGaming/back/services/user/internal/helper"
	"SLGaming/back/services/user/internal/svc"
	"SLGaming/back/services/user/user"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
		helper.LogError(l.Logger, helper.OpConsume, "invalid user_id", nil, map[string]interface{}{
			"user_id": userID,
		})
		return nil, status.Error(codes.InvalidArgument, "user_id is required and must be greater than 0")
	}
	if amount <= 0 {
		helper.LogError(l.Logger, helper.OpConsume, "invalid amount", nil, map[string]interface{}{
			"user_id": userID,
			"amount":  amount,
		})
		return nil, status.Error(codes.InvalidArgument, "amount must be positive")
	}

	// 使用钱包服务统一处理
	walletService := helper.NewWalletService(l.svcCtx.DB())
	result, err := walletService.UpdateBalance(l.ctx, &helper.WalletUpdateRequest{
		UserID:     userID,
		Amount:     amount,
		Type:       helper.WalletOpConsume,
		BizOrderID: bizOrderID,
		Remark:     in.GetRemark(),
		Logger:     l.Logger,
	})
	if err != nil {
		return nil, err
	}

	return &user.ConsumeResponse{
		Wallet: result.ToWalletInfo(),
	}, nil
}
