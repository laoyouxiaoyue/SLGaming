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
)

type GetWalletLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetWalletLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetWalletLogic {
	return &GetWalletLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetWalletLogic) GetWallet(in *user.GetWalletRequest) (*user.GetWalletResponse, error) {
	userID := in.GetUserId()
	if userID == 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	db := l.svcCtx.DB().WithContext(l.ctx)

	var wallet model.UserWallet
	err := db.Where("user_id = ?", userID).First(&wallet).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// 尚未创建钱包，为用户初始化一个默认钱包（余额为 0）
		wallet = model.UserWallet{
			UserID:        userID,
			Balance:       0,
			FrozenBalance: 0,
		}
		if err := db.Create(&wallet).Error; err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	} else if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &user.GetWalletResponse{
		Wallet: &user.WalletInfo{
			UserId:        wallet.UserID,
			Balance:       wallet.Balance,
			FrozenBalance: wallet.FrozenBalance,
		},
	}, nil
}
