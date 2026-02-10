// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package follow

import (
	"context"

	"SLGaming/back/services/gateway/internal/middleware"
	"SLGaming/back/services/gateway/internal/svc"
	"SLGaming/back/services/gateway/internal/types"
	"SLGaming/back/services/gateway/internal/utils"
	"SLGaming/back/services/user/userclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMutualFollowListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetMutualFollowListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMutualFollowListLogic {
	return &GetMutualFollowListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetMutualFollowListLogic) GetMutualFollowList(req *types.GetMutualFollowListRequest) (resp *types.GetMutualFollowListResponse, err error) {
	if l.svcCtx.UserRPC == nil {
		code, msg := utils.HandleRPCClientUnavailable(l.Logger, "UserRPC")
		return &types.GetMutualFollowListResponse{BaseResp: types.BaseResp{Code: code, Msg: msg}}, nil
	}

	userID, err := middleware.GetUserID(l.ctx)
	if err != nil || userID == 0 {
		return &types.GetMutualFollowListResponse{
			BaseResp: types.BaseResp{Code: 401, Msg: "未登录或认证失败"},
		}, nil
	}

	rpcResp, err := l.svcCtx.UserRPC.GetMutualFollowList(l.ctx, &userclient.GetMutualFollowListRequest{
		OperatorId: userID,
		Page:       int32(req.Page),
		PageSize:   int32(req.PageSize),
	})
	if err != nil {
		code, msg := utils.HandleRPCError(err, l.Logger, "GetMutualFollowList")
		return &types.GetMutualFollowListResponse{BaseResp: types.BaseResp{Code: code, Msg: msg}}, nil
	}

	users := make([]types.UserFollowInfo, 0, len(rpcResp.GetUsers()))
	for _, u := range rpcResp.GetUsers() {
		users = append(users, types.UserFollowInfo{
			UserId:      u.GetUserId(),
			Nickname:    u.GetNickname(),
			AvatarUrl:   u.GetAvatarUrl(),
			Role:        int(u.GetRole()),
			IsVerified:  u.GetIsVerified(),
			Rating:      u.GetRating(),
			TotalOrders: u.GetTotalOrders(),
			IsMutual:    u.GetIsMutual(),
			FollowedAt:  u.GetFollowedAt(),
		})
	}

	return &types.GetMutualFollowListResponse{
		BaseResp: types.BaseResp{Code: 0, Msg: "success"},
		Data: types.GetMutualFollowListData{
			Users:    users,
			Total:    int(rpcResp.GetTotal()),
			Page:     int(rpcResp.GetPage()),
			PageSize: int(rpcResp.GetPageSize()),
		},
	}, nil
}
