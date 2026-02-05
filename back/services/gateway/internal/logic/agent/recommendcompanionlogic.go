// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package agent

import (
	"context"
	"strings"

	"SLGaming/back/services/agent/agentclient"
	"SLGaming/back/services/gateway/internal/middleware"
	"SLGaming/back/services/gateway/internal/svc"
	"SLGaming/back/services/gateway/internal/types"
	"SLGaming/back/services/gateway/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type RecommendCompanionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRecommendCompanionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RecommendCompanionLogic {
	return &RecommendCompanionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RecommendCompanionLogic) RecommendCompanion(req *types.RecommendCompanionRequest) (resp *types.RecommendCompanionResponse, err error) {
	input := strings.TrimSpace(req.UserInput)
	if input == "" {
		return &types.RecommendCompanionResponse{
			BaseResp: types.BaseResp{Code: 400, Msg: "用户输入不能为空"},
		}, nil
	}

	if l.svcCtx.AgentRPC == nil {
		code, msg := utils.HandleRPCClientUnavailable(l.Logger, "AgentRPC")
		return &types.RecommendCompanionResponse{
			BaseResp: types.BaseResp{Code: code, Msg: msg},
		}, nil
	}

	userID, err := middleware.GetUserID(l.ctx)
	if err != nil {
		return &types.RecommendCompanionResponse{
			BaseResp: types.BaseResp{Code: 401, Msg: "未登录或登录已过期"},
		}, nil
	}

	rpcResp, err := l.svcCtx.AgentRPC.RecommendCompanion(l.ctx, &agentclient.RecommendCompanionRequest{
		UserInput: input,
		UserId:    userID,
	})
	if err != nil {
		code, msg := utils.HandleRPCError(err, l.Logger, "RecommendCompanion")
		return &types.RecommendCompanionResponse{
			BaseResp: types.BaseResp{Code: code, Msg: msg},
		}, nil
	}

	companions := make([]types.CompanionRecommendation, 0, len(rpcResp.Companions))
	for _, c := range rpcResp.Companions {
		companions = append(companions, types.CompanionRecommendation{
			UserId:       c.UserId,
			GameSkill:    c.GameSkill,
			Gender:       c.Gender,
			Age:          c.Age,
			Description:  c.Description,
			PricePerHour: c.PricePerHour,
			Rating:       c.Rating,
			Similarity:   c.Similarity,
		})
	}

	l.Infof("recommend companion success user_id=%d input_len=%d results=%d", userID, len(input), len(companions))

	return &types.RecommendCompanionResponse{
		BaseResp: types.BaseResp{Code: 0, Msg: "success"},
		Data: types.RecommendCompanionData{
			Companions:  companions,
			Explanation: rpcResp.Explanation,
		},
	}, nil
}
