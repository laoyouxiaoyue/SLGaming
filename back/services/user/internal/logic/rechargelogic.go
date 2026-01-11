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

	return &user.RechargeResponse{
		Wallet: result.ToWalletInfo(),
	}, nil
}
