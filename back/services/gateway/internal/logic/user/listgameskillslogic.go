// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"context"
	"fmt"

	"SLGaming/back/services/gateway/internal/svc"
	"SLGaming/back/services/gateway/internal/types"
	"SLGaming/back/services/gateway/internal/utils"
	"SLGaming/back/services/user/userclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListGameSkillsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListGameSkillsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListGameSkillsLogic {
	return &ListGameSkillsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListGameSkillsLogic) ListGameSkills() (resp *types.ListGameSkillsResponse, err error) {
	if l.svcCtx.UserRPC == nil {
		return nil, fmt.Errorf("user rpc client not initialized")
	}

	rpcResp, err := l.svcCtx.UserRPC.ListGameSkills(l.ctx, &userclient.ListGameSkillsRequest{})
	if err != nil {
		code, msg := utils.HandleRPCError(err, l.Logger, "ListGameSkills")
		return &types.ListGameSkillsResponse{
			BaseResp: types.BaseResp{Code: code, Msg: msg},
		}, nil
	}

	data := make([]types.GameSkill, 0, len(rpcResp.Skills))
	for _, gs := range rpcResp.Skills {
		data = append(data, types.GameSkill{
			Id:          gs.Id,
			Name:        gs.Name,
			Description: gs.Description,
		})
	}

	return &types.ListGameSkillsResponse{
		BaseResp: types.BaseResp{Code: 0, Msg: "success"},
		Data:     data,
	}, nil
}
