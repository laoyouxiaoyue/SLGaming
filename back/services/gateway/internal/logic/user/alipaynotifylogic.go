// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/url"
	"strconv"
	"strings"
	"time"

	"SLGaming/back/pkg/rechargemq"
	"SLGaming/back/services/gateway/internal/svc"
	"SLGaming/back/services/gateway/internal/types"
	"SLGaming/back/services/gateway/internal/utils"
	"SLGaming/back/services/user/userclient"

	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/zeromicro/go-zero/core/logx"
)

type AlipayNotifyLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAlipayNotifyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AlipayNotifyLogic {
	return &AlipayNotifyLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AlipayNotifyLogic) AlipayNotify(req *types.AlipayNotifyRequest) (resp *types.AlipayNotifyResponse, err error) {
	if l.svcCtx.Alipay == nil {
		return &types.AlipayNotifyResponse{
			BaseResp: types.BaseResp{Code: 500, Msg: "支付宝未配置"},
		}, nil
	}
	if req.Payload == nil || len(req.Payload) == 0 {
		return &types.AlipayNotifyResponse{
			BaseResp: types.BaseResp{Code: 400, Msg: "通知参数不能为空"},
		}, nil
	}

	outTradeNo := strings.TrimSpace(req.Payload["out_trade_no"])
	tradeStatus := strings.TrimSpace(req.Payload["trade_status"])
	totalAmount := strings.TrimSpace(req.Payload["total_amount"])
	tradeNo := strings.TrimSpace(req.Payload["trade_no"])
	if outTradeNo == "" || tradeStatus == "" {
		return &types.AlipayNotifyResponse{
			BaseResp: types.BaseResp{Code: 400, Msg: "通知参数缺失"},
		}, nil
	}

	// 验签
	values := url.Values{}
	for k, v := range req.Payload {
		values.Set(k, v)
	}
	if err := l.svcCtx.Alipay.VerifySign(values); err != nil {
		return &types.AlipayNotifyResponse{
			BaseResp: types.BaseResp{Code: 400, Msg: "验签失败"},
		}, nil
	}

	order, err := loadRechargeOrder(l.svcCtx.CacheRedis, outTradeNo)
	if err != nil {
		code, msg := utils.HandleError(err, l.Logger, "LoadRechargeOrder")
		return &types.AlipayNotifyResponse{
			BaseResp: types.BaseResp{Code: code, Msg: msg},
		}, nil
	}

	if order.Status == rechargeStatusSuccess {
		return &types.AlipayNotifyResponse{
			BaseResp: types.BaseResp{Code: 0, Msg: "success"},
		}, nil
	}

	if totalAmount != "" {
		if f, err := strconv.ParseFloat(totalAmount, 64); err == nil {
			amount := int64(math.Round(f))
			if order.Amount != amount {
				return &types.AlipayNotifyResponse{
					BaseResp: types.BaseResp{Code: 400, Msg: "金额校验失败"},
				}, nil
			}
		}
	}

	if tradeStatus == "TRADE_SUCCESS" || tradeStatus == "TRADE_FINISHED" {
		order.Status = rechargeStatusSuccess
	} else if tradeStatus == "TRADE_CLOSED" {
		order.Status = rechargeStatusClosed
	} else {
		order.Status = rechargeStatusFailed
	}

	if l.svcCtx.EventProducer != nil {
		paidAt := int64(0)
		if order.Status == rechargeStatusSuccess {
			paidAt = time.Now().Unix()
		}
		payload := &rechargemq.RechargeEventPayload{
			OrderNo: order.OrderNo,
			UserID:  order.UserId,
			Amount:  order.Amount,
			Status:  int(order.Status),
			PayType: order.PayType,
			TradeNo: tradeNo,
			PaidAt:  paidAt,
			Remark:  "alipay notify",
		}
		var tag string
		switch order.Status {
		case rechargeStatusSuccess:
			tag = rechargemq.EventTypeRechargeSuccess()
		case rechargeStatusClosed:
			tag = rechargemq.EventTypeRechargeClosed()
		default:
			tag = rechargemq.EventTypeRechargeFailed()
		}
		if err := l.publishRechargeEvent(tag, payload); err != nil {
			code, msg := utils.HandleError(err, l.Logger, "PublishRechargeEvent")
			return &types.AlipayNotifyResponse{
				BaseResp: types.BaseResp{Code: code, Msg: msg},
			}, nil
		}
	} else {
		if l.svcCtx.UserRPC == nil {
			code, msg := utils.HandleRPCClientUnavailable(l.Logger, "UserRPC")
			return &types.AlipayNotifyResponse{
				BaseResp: types.BaseResp{Code: code, Msg: msg},
			}, nil
		}
		if order.Status == rechargeStatusSuccess {
			_, err = l.svcCtx.UserRPC.Recharge(l.ctx, &userclient.RechargeRequest{
				UserId:     order.UserId,
				Amount:     order.Amount,
				BizOrderId: order.OrderNo,
				Remark:     "alipay recharge",
			})
			if err != nil {
				code, msg := utils.HandleRPCError(err, l.Logger, "Recharge")
				return &types.AlipayNotifyResponse{
					BaseResp: types.BaseResp{Code: code, Msg: msg},
				}, nil
			}
		}

		_, err = l.svcCtx.UserRPC.UpdateRechargeOrderStatus(l.ctx, &userclient.UpdateRechargeOrderStatusRequest{
			OrderNo: order.OrderNo,
			Status:  int32(order.Status),
			TradeNo: tradeNo,
			Remark:  "alipay notify",
			PaidAt:  time.Now().Unix(),
		})
		if err != nil {
			l.Logger.Errorf("update recharge order status failed: %v", err)
		}
	}

	// 更新订单状态
	if order.ExpiresAt == 0 {
		order.ExpiresAt = time.Now().Add(30 * time.Minute).Unix()
	}
	remaining := time.Until(time.Unix(order.ExpiresAt, 0))
	if remaining <= 0 {
		remaining = 5 * time.Minute
	}
	_ = saveRechargeOrder(l.svcCtx.CacheRedis, order, remaining)

	return &types.AlipayNotifyResponse{
		BaseResp: types.BaseResp{Code: 0, Msg: "success"},
	}, nil
}

func (l *AlipayNotifyLogic) publishRechargeEvent(tag string, payload *rechargemq.RechargeEventPayload) error {
	if l.svcCtx.EventProducer == nil {
		return fmt.Errorf("rocketmq producer not initialized")
	}
	if payload == nil {
		return fmt.Errorf("payload is nil")
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	msg := primitive.NewMessage(rechargemq.RechargeEventTopic(), body)
	msg.WithTag(tag)
	_, err = l.svcCtx.EventProducer.SendSync(l.ctx, msg)
	return err
}
