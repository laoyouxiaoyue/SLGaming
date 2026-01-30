package logic

import (
	"context"
	"errors"

	"SLGaming/back/services/user/internal/helper"
	"SLGaming/back/services/user/internal/model"
	"SLGaming/back/services/user/internal/svc"
	"SLGaming/back/services/user/user"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

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
	db := l.svcCtx.DB().WithContext(l.ctx)
	var u model.User
	var err error

	switch {
	case in.GetId() != 0:
		err = db.Where("id = ?", in.GetId()).First(&u).Error
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

	// 获取钱包信息
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

	return &user.GetUserResponse{
		User: userInfo,
	}, nil
}
