// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"SLGaming/back/services/user/userclient"
	"context"
	"fmt"
	"strings"
	"time"

	"SLGaming/back/pkg/snowflake"
	"SLGaming/back/services/gateway/internal/middleware"
	"SLGaming/back/services/gateway/internal/svc"
	"SLGaming/back/services/gateway/internal/types"
	"SLGaming/back/services/gateway/internal/utils"

	"github.com/smartwalle/alipay/v3"
	"github.com/zeromicro/go-zero/core/logx"
)

const (
	payTypeAlipayWap  = "alipay_wap"
	payTypeAlipayPage = "alipay_page"
	payTypeAlipayApp  = "alipay_app"
)

type RechargeCreateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRechargeCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RechargeCreateLogic {
	return &RechargeCreateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RechargeCreateLogic) RechargeCreate(req *types.RechargeCreateRequest) (resp *types.RechargeCreateResponse, err error) {
	userID, err := middleware.GetUserID(l.ctx)
	if err != nil {
		return &types.RechargeCreateResponse{
			BaseResp: types.BaseResp{Code: 401, Msg: "未登录或登录已过期"},
		}, nil
	}

	if req.Amount <= 0 {
		return &types.RechargeCreateResponse{
			BaseResp: types.BaseResp{Code: 400, Msg: "充值金额必须大于0"},
		}, nil
	}

	payType := strings.TrimSpace(req.PayType)
	if payType == "" || payType == payTypeAlipayWap {
		payType = payTypeAlipayPage
	}
	if payType != payTypeAlipayWap && payType != payTypeAlipayPage && payType != payTypeAlipayApp {
		return &types.RechargeCreateResponse{
			BaseResp: types.BaseResp{Code: 400, Msg: "支付方式不支持"},
		}, nil
	}

	orderNo := "RC" + snowflake.EncodeBase36(snowflake.GenID())
	now := time.Now()
	expiresIn := int64(30 * time.Minute / time.Second)
	order := &rechargeOrder{
		OrderNo:   orderNo,
		UserId:    userID,
		Amount:    req.Amount,
		Status:    rechargeStatusPending,
		PayType:   payType,
		CreatedAt: now.Unix(),
		ExpiresAt: now.Add(30 * time.Minute).Unix(),
	}
	if err := saveRechargeOrder(l.svcCtx.CacheRedis, order, 30*time.Minute); err != nil {
		code, msg := utils.HandleError(err, l.Logger, "SaveRechargeOrder")
		return &types.RechargeCreateResponse{
			BaseResp: types.BaseResp{Code: code, Msg: msg},
		}, nil
	}

	if l.svcCtx.UserRPC != nil {
		_, err = l.svcCtx.UserRPC.CreateRechargeOrder(l.ctx, &userclient.CreateRechargeOrderRequest{
			UserId:  userID,
			OrderNo: orderNo,
			Amount:  req.Amount,
			PayType: payType,
			Remark:  "pending",
			Status:  rechargeStatusPending,
		})
		if err != nil {
			code, msg := utils.HandleRPCError(err, l.Logger, "CreateRechargeOrder")
			return &types.RechargeCreateResponse{
				BaseResp: types.BaseResp{Code: code, Msg: msg},
			}, nil
		}
	}

	if l.svcCtx.Alipay == nil {
		return &types.RechargeCreateResponse{
			BaseResp: types.BaseResp{Code: 500, Msg: "支付宝未配置"},
		}, nil
	}

	amountStr := fmt.Sprintf("%.2f", float64(req.Amount))
	notifyURL := strings.TrimSpace(l.svcCtx.Config.Alipay.NotifyURL)
	returnURL := strings.TrimSpace(l.svcCtx.Config.Alipay.ReturnURL)
	if req.ReturnUrl != "" {
		returnURL = req.ReturnUrl
	}

	var payUrl string
	var payForm string
	subject := "余额充值"

	switch payType {
	case payTypeAlipayWap:
		p := alipay.TradeWapPay{}
		p.NotifyURL = notifyURL
		p.ReturnURL = returnURL
		p.Subject = subject
		p.OutTradeNo = orderNo
		p.TotalAmount = amountStr
		p.ProductCode = "QUICK_WAP_WAY"
		u, err := l.svcCtx.Alipay.TradeWapPay(p)
		if err != nil {
			code, msg := utils.HandleError(err, l.Logger, "TradeWapPay")
			return &types.RechargeCreateResponse{
				BaseResp: types.BaseResp{Code: code, Msg: msg},
			}, nil
		}
		payUrl = u.String()
	case payTypeAlipayPage:
		p := alipay.TradePagePay{}
		p.NotifyURL = notifyURL
		p.ReturnURL = returnURL
		p.Subject = subject
		p.OutTradeNo = orderNo
		p.TotalAmount = amountStr
		p.ProductCode = "FAST_INSTANT_TRADE_PAY"
		u, err := l.svcCtx.Alipay.TradePagePay(p)
		if err != nil {
			code, msg := utils.HandleError(err, l.Logger, "TradePagePay")
			return &types.RechargeCreateResponse{
				BaseResp: types.BaseResp{Code: code, Msg: msg},
			}, nil
		}
		payUrl = u.String()
	case payTypeAlipayApp:
		p := alipay.TradeAppPay{}
		p.NotifyURL = notifyURL
		p.Subject = subject
		p.OutTradeNo = orderNo
		p.TotalAmount = amountStr
		p.ProductCode = "QUICK_MSECURITY_PAY"
		orderStr, err := l.svcCtx.Alipay.TradeAppPay(p)
		if err != nil {
			code, msg := utils.HandleError(err, l.Logger, "TradeAppPay")
			return &types.RechargeCreateResponse{
				BaseResp: types.BaseResp{Code: code, Msg: msg},
			}, nil
		}
		payForm = orderStr
	}

	return &types.RechargeCreateResponse{
		BaseResp: types.BaseResp{Code: 0, Msg: "success"},
		Data: types.RechargeCreateData{
			OrderNo:   orderNo,
			PayUrl:    payUrl,
			PayForm:   payForm,
			ExpiresIn: expiresIn,
		},
	}, nil
}
