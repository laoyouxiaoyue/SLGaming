// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"context"
	"encoding/json"
	"time"

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
		code, msg := utils.HandleRPCClientUnavailable(l.Logger, "UserRPC")
		return &types.ListGameSkillsResponse{
			BaseResp: types.BaseResp{Code: code, Msg: msg},
		}, nil
	}

	const cacheKey = "cache:gameskills:all"
	if l.svcCtx.CacheRedis != nil {
		if cached, cacheErr := l.svcCtx.CacheRedis.Get(cacheKey); cacheErr == nil && cached != "" {
			var cachedResp types.ListGameSkillsResponse
			if unmarshalErr := json.Unmarshal([]byte(cached), &cachedResp); unmarshalErr == nil {
				return &cachedResp, nil
			}
		}
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

	resp = &types.ListGameSkillsResponse{
		BaseResp: types.BaseResp{Code: 0, Msg: "success"},
		Data:     data,
	}

	if l.svcCtx.CacheRedis != nil {
		if payload, marshalErr := json.Marshal(resp); marshalErr == nil {
			_ = l.svcCtx.CacheRedis.Setex(cacheKey, string(payload), int((30 * time.Minute).Seconds()))
		}
	}

	return resp, nil
}
