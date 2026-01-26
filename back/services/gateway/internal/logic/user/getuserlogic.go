// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"SLGaming/back/services/gateway/internal/middleware"
	"context"

	"SLGaming/back/services/gateway/internal/svc"
	"SLGaming/back/services/gateway/internal/types"
	"SLGaming/back/services/gateway/internal/utils"
	"SLGaming/back/services/user/userclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserLogic {
	return &GetUserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserLogic) GetUser(req *types.GetUserRequest) (resp *types.GetUserResponse, err error) {
	if l.svcCtx.UserRPC == nil {
		code, msg := utils.HandleRPCClientUnavailable(l.Logger, "UserRPC")
		return &types.GetUserResponse{
			BaseResp: types.BaseResp{Code: code, Msg: msg},
		}, nil
	}

	// 如果所有查询条件都为空，使用当前登录用户的 ID 作为默认值
	if req.Id == 0 && req.Uid == 0 && req.Phone == "" {
		userID, err := middleware.GetUserID(l.ctx)
		if err != nil {
			return &types.GetUserResponse{
				BaseResp: types.BaseResp{
					Code: 401,
					Msg:  "未登录或登录已过期",
				},
			}, nil
		}
		req.Id = userID
	}

	// 调用用户服务的 RPC
	rpcResp, err := l.svcCtx.UserRPC.GetUser(l.ctx, &userclient.GetUserRequest{
		Id:    req.Id,
		Uid:   req.Uid,
		Phone: req.Phone,
	})
	if err != nil {
		code, msg := utils.HandleRPCError(err, l.Logger, "GetUser")
		return &types.GetUserResponse{
			BaseResp: types.BaseResp{
				Code: code,
				Msg:  msg,
			},
		}, nil
	}

	return &types.GetUserResponse{
		BaseResp: types.BaseResp{
			Code: 0,
			Msg:  "success",
		},
		Data: types.UserInfo{
			Id:        rpcResp.User.Id,
			Uid:       rpcResp.User.Uid,
			Nickname:  rpcResp.User.Nickname,
			Phone:     rpcResp.User.Phone,
			Role:      int(rpcResp.User.Role),
			AvatarUrl: rpcResp.User.AvatarUrl,
			Bio:       rpcResp.User.Bio,
		},
	}, nil
}
