// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"context"

	"SLGaming/back/services/gateway/internal/svc"
	"SLGaming/back/services/gateway/internal/types"
	"SLGaming/back/services/gateway/internal/utils"
	"SLGaming/back/services/user/userclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetCompanionRatingRankingLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetCompanionRatingRankingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCompanionRatingRankingLogic {
	return &GetCompanionRatingRankingLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetCompanionRatingRankingLogic) GetCompanionRatingRanking(req *types.GetCompanionRatingRankingRequest) (*types.GetCompanionRatingRankingResponse, error) {
	if l.svcCtx.UserRPC == nil {
		code, msg := utils.HandleRPCClientUnavailable(l.Logger, "UserRPC")
		return &types.GetCompanionRatingRankingResponse{
			BaseResp: types.BaseResp{Code: code, Msg: msg},
		}, nil
	}

	rpcResp, err := l.svcCtx.UserRPC.GetCompanionRatingRanking(l.ctx, &userclient.GetCompanionRatingRankingRequest{
		Page:     int32(req.Page),
		PageSize: int32(req.PageSize),
	})
	if err != nil {
		code, msg := utils.HandleRPCError(err, l.Logger, "GetCompanionRatingRanking")
		return &types.GetCompanionRatingRankingResponse{
			BaseResp: types.BaseResp{Code: code, Msg: msg},
		}, nil
	}

	items := make([]types.CompanionRankingItem, 0, len(rpcResp.Rankings))
	for _, item := range rpcResp.Rankings {
		items = append(items, types.CompanionRankingItem{
			UserId:      item.UserId,
			Nickname:    item.Nickname,
			AvatarUrl:   item.AvatarUrl,
			Rating:      item.Rating,
			TotalOrders: item.TotalOrders,
			Rank:        item.Rank,
			IsVerified:  item.IsVerified,
		})
	}

	resp := &types.GetCompanionRatingRankingResponse{
		BaseResp: types.BaseResp{Code: 0, Msg: "success"},
		Data: types.GetCompanionRatingRankingData{
			Rankings: items,
			Total:    rpcResp.Total,
			Page:     rpcResp.Page,
			PageSize: rpcResp.PageSize,
		},
	}

	return resp, nil
}
