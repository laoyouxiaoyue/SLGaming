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

type GetMyFollowingListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetMyFollowingListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMyFollowingListLogic {
	return &GetMyFollowingListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetMyFollowingListLogic) GetMyFollowingList(req *types.GetMyFollowingListRequest) (resp *types.GetMyFollowingListResponse, err error) {
	if l.svcCtx.UserRPC == nil {
		code, msg := utils.HandleRPCClientUnavailable(l.Logger, "UserRPC")
		return &types.GetMyFollowingListResponse{BaseResp: types.BaseResp{Code: code, Msg: msg}}, nil
	}

	userID, err := middleware.GetUserID(l.ctx)
	if err != nil || userID == 0 {
		return &types.GetMyFollowingListResponse{
			BaseResp: types.BaseResp{Code: 401, Msg: "未登录或认证失败"},
		}, nil
	}

	rpcResp, err := l.svcCtx.UserRPC.GetMyFollowingList(l.ctx, &userclient.GetMyFollowingListRequest{
		OperatorId: userID,
		Page:       int32(req.Page),
		PageSize:   int32(req.PageSize),
		Keyword:    req.Keyword,
	})
	if err != nil {
		code, msg := utils.HandleRPCError(err, l.Logger, "GetMyFollowingList")
		return &types.GetMyFollowingListResponse{BaseResp: types.BaseResp{Code: code, Msg: msg}}, nil
	}

	users := make([]types.UserFollowInfo, 0, len(rpcResp.GetUsers()))
	for _, u := range rpcResp.GetUsers() {
		users = append(users, types.UserFollowInfo{
			UserId:     u.GetUserId(),
			Nickname:   u.GetNickname(),
			AvatarUrl:  u.GetAvatarUrl(),
			IsMutual:   u.GetIsMutual(),
			FollowedAt: u.GetFollowedAt(),
		})
	}

	return &types.GetMyFollowingListResponse{
		BaseResp: types.BaseResp{Code: 0, Msg: "success"},
		Data: types.GetMyFollowingListData{
			Users:    users,
			Total:    int(rpcResp.GetTotal()),
			Page:     int(rpcResp.GetPage()),
			PageSize: int(rpcResp.GetPageSize()),
		},
	}, nil
}
