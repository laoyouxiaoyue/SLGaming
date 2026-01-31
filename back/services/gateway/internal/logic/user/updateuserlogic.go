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

type UpdateUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateUserLogic {
	return &UpdateUserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateUserLogic) UpdateUser(req *types.UpdateUserRequest) (resp *types.UpdateUserResponse, err error) {
	// 从 context 中获取当前登录用户 ID
	currentUserID, err := middleware.GetUserID(l.ctx)
	if err != nil {
		return &types.UpdateUserResponse{
			BaseResp: types.BaseResp{
				Code: 401,
				Msg:  "未登录或登录已过期",
			},
		}, nil
	}

	// 如果请求中没有指定 ID，使用当前登录用户的 ID
	if req.Id == 0 {
		req.Id = currentUserID
	} else {
		// 如果指定了 ID，必须与当前登录用户的 ID 一致（防止用户修改他人信息）
		if req.Id != currentUserID {
			return &types.UpdateUserResponse{
				BaseResp: types.BaseResp{
					Code: 403,
					Msg:  "无权修改其他用户的信息",
				},
			}, nil
		}
	}

	if l.svcCtx.UserRPC == nil {
		code, msg := utils.HandleRPCClientUnavailable(l.Logger, "UserRPC")
		return &types.UpdateUserResponse{
			BaseResp: types.BaseResp{Code: code, Msg: msg},
		}, nil
	}

	// 获取当前用户角色（优先 JWT 里的 role）
	currentRole, roleErr := middleware.GetUserRole(l.ctx)
	if roleErr != nil {
		currentRole = 0
	}
	// 非管理员仅允许修改昵称/密码/手机号/bio，忽略 role 与 avatarUrl
	if currentRole != 3 {
		req.Role = 0
		req.AvatarUrl = ""
	}

	// 调用用户服务的 RPC
	rpcResp, err := l.svcCtx.UserRPC.UpdateUser(l.ctx, &userclient.UpdateUserRequest{
		Id:        req.Id,
		Nickname:  req.Nickname,
		Password:  req.Password,
		Phone:     req.Phone,
		Role:      int32(req.Role),
		AvatarUrl: req.AvatarUrl,
		Bio:       req.Bio,
	})
	if err != nil {
		code, msg := utils.HandleRPCError(err, l.Logger, "UpdateUser")
		return &types.UpdateUserResponse{
			BaseResp: types.BaseResp{
				Code: code,
				Msg:  msg,
			},
		}, nil
	}

	return &types.UpdateUserResponse{
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
