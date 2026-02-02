package job

import (
	"context"
	"encoding/json"
	"time"

	pkgIoc "SLGaming/back/pkg/ioc"
	"SLGaming/back/pkg/rechargemq"
	"SLGaming/back/services/user/internal/logic"
	"SLGaming/back/services/user/internal/model"
	"SLGaming/back/services/user/internal/svc"
	"SLGaming/back/services/user/user"

	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/zeromicro/go-zero/core/logx"
)

// StartRechargeEventConsumer 启动充值回调事件的 RocketMQ Consumer
func StartRechargeEventConsumer(ctx context.Context, svcCtx *svc.ServiceContext) {
	cfg := svcCtx.Config().RocketMQ
	if len(cfg.NameServers) == 0 {
		logx.Infof("recharge event consumer not started: rocketmq not configured")
		return
	}

	mqCfg := &pkgIoc.RocketMQConfigAdapter{
		NameServers: cfg.NameServers,
		Namespace:   cfg.Namespace,
		AccessKey:   cfg.AccessKey,
		SecretKey:   cfg.SecretKey,
	}

	consumer, err := pkgIoc.InitRocketMQConsumer(
		mqCfg,
		"user-recharge-consumer",
		[]string{rechargemq.RechargeEventTopic()},
		func(c context.Context, msg *primitive.MessageExt) error {
			return handleRechargeEvent(c, svcCtx, msg)
		},
	)
	if err != nil {
		logx.Errorf("init recharge event consumer failed: %v", err)
		return
	}

	// 确保进程退出时关闭 consumer
	go func() {
		<-ctx.Done()
		pkgIoc.ShutdownRocketMQConsumer(consumer)
	}()

	logx.Infof("recharge event consumer started, topic=%s", rechargemq.RechargeEventTopic())
}

func handleRechargeEvent(ctx context.Context, svcCtx *svc.ServiceContext, msg *primitive.MessageExt) error {
	var payload rechargemq.RechargeEventPayload
	if err := json.Unmarshal(msg.Body, &payload); err != nil {
		logx.Errorf("unmarshal recharge event payload failed: %v, body=%s", err, string(msg.Body))
		return nil // 丢弃这条，避免一直重试
	}

	if payload.OrderNo == "" || payload.UserID == 0 || payload.Amount <= 0 {
		logx.Errorf("invalid recharge event payload: order_no=%s, user_id=%d, amount=%d", payload.OrderNo, payload.UserID, payload.Amount)
		return nil
	}

	switch msg.GetTags() {
	case rechargemq.EventTypeRechargeSuccess():
		return handleRechargeSuccess(ctx, svcCtx, &payload)
	case rechargemq.EventTypeRechargeClosed():
		return handleRechargeStatusUpdate(ctx, svcCtx, &payload, model.RechargeStatusClosed)
	case rechargemq.EventTypeRechargeFailed():
		return handleRechargeStatusUpdate(ctx, svcCtx, &payload, model.RechargeStatusFailed)
	default:
		// 未识别的 tag 忽略
		return nil
	}
}

func handleRechargeSuccess(ctx context.Context, svcCtx *svc.ServiceContext, payload *rechargemq.RechargeEventPayload) error {
	// 先入账（幂等）
	rechargeLogic := logic.NewRechargeLogic(ctx, svcCtx)
	if _, err := rechargeLogic.Recharge(&user.RechargeRequest{
		UserId:     payload.UserID,
		Amount:     payload.Amount,
		BizOrderId: payload.OrderNo,
		Remark:     payload.Remark,
	}); err != nil {
		logx.Errorf("recharge wallet failed: order_no=%s, user_id=%d, amount=%d, err=%v", payload.OrderNo, payload.UserID, payload.Amount, err)
		return err
	}

	// 更新订单状态/交易号/支付时间
	paidAt := payload.PaidAt
	if paidAt <= 0 {
		paidAt = time.Now().Unix()
	}
	updateLogic := logic.NewUpdateRechargeOrderStatusLogic(ctx, svcCtx)
	_, err := updateLogic.UpdateRechargeOrderStatus(&user.UpdateRechargeOrderStatusRequest{
		OrderNo: payload.OrderNo,
		Status:  int32(model.RechargeStatusSuccess),
		TradeNo: payload.TradeNo,
		Remark:  payload.Remark,
		PaidAt:  paidAt,
	})
	if err != nil {
		logx.Errorf("update recharge order status failed: order_no=%s, err=%v", payload.OrderNo, err)
		return err
	}

	return nil
}

func handleRechargeStatusUpdate(ctx context.Context, svcCtx *svc.ServiceContext, payload *rechargemq.RechargeEventPayload, status int) error {
	updateLogic := logic.NewUpdateRechargeOrderStatusLogic(ctx, svcCtx)
	paidAt := payload.PaidAt
	if status == model.RechargeStatusSuccess && paidAt <= 0 {
		paidAt = time.Now().Unix()
	}
	_, err := updateLogic.UpdateRechargeOrderStatus(&user.UpdateRechargeOrderStatusRequest{
		OrderNo: payload.OrderNo,
		Status:  int32(status),
		TradeNo: payload.TradeNo,
		Remark:  payload.Remark,
		PaidAt:  paidAt,
	})
	if err != nil {
		logx.Errorf("update recharge order status failed: order_no=%s, status=%d, err=%v", payload.OrderNo, status, err)
		return err
	}
	return nil
}
