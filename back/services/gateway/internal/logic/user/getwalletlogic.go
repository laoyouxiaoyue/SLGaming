// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"context"

	"SLGaming/back/services/gateway/internal/middleware"
	"SLGaming/back/services/gateway/internal/svc"
	"SLGaming/back/services/gateway/internal/types"
	"SLGaming/back/services/gateway/internal/utils"
	"SLGaming/back/services/user/userclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetWalletLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetWalletLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetWalletLogic {
	return &GetWalletLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetWalletLogic) GetWallet() (resp *types.GetWalletResponse, err error) {
	// 从 context 中获取当前登录用户 ID（由网关鉴权中间件注入）
	userID, err := middleware.GetUserID(l.ctx)

	// 调用 User RPC 的 GetWallet 接口
	rpcResp, err := l.svcCtx.UserRPC.GetWallet(l.ctx, &userclient.GetWalletRequest{
		UserId: userID,
	})
	if err != nil {
		code, msg := utils.HandleRPCError(err, l.Logger, "GetWallet")
		return &types.GetWalletResponse{
			BaseResp: types.BaseResp{
				Code: code,
				Msg:  msg,
			},
			Data: types.WalletInfo{},
		}, nil
	}

	wallet := rpcResp.GetWallet()
	if wallet == nil {
		return &types.GetWalletResponse{
			BaseResp: types.BaseResp{
				Code: 0,
				Msg:  "success",
			},
			Data: types.WalletInfo{},
		}, nil
	}

	return &types.GetWalletResponse{
		BaseResp: types.BaseResp{
			Code: 0,
			Msg:  "success",
		},
		Data: types.WalletInfo{
			UserId:        wallet.GetUserId(),
			Balance:       wallet.GetBalance(),
			FrozenBalance: wallet.GetFrozenBalance(),
		},
	}, nil
}
