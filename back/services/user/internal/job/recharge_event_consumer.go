package job

import (
	"context"
	"encoding/json"
	"time"

	pkgIoc "SLGaming/back/pkg/ioc"
	"SLGaming/back/pkg/rechargemq"
	"SLGaming/back/services/user/internal/helper"
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
		helper.LogInfo(logx.WithContext(ctx), helper.OpMQConsumer, "recharge event consumer not started: rocketmq not configured", nil)
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
		helper.LogError(logx.WithContext(ctx), helper.OpMQConsumer, "init recharge event consumer failed", err, nil)
		return
	}

	// 确保进程退出时关闭 consumer
	go func() {
		<-ctx.Done()
		pkgIoc.ShutdownRocketMQConsumer(consumer)
	}()

	helper.LogSuccess(logx.WithContext(ctx), helper.OpMQConsumer, map[string]interface{}{
		"consumer": "recharge_event",
		"topic":    rechargemq.RechargeEventTopic(),
	})
}

func handleRechargeEvent(ctx context.Context, svcCtx *svc.ServiceContext, msg *primitive.MessageExt) error {
	logger := logx.WithContext(ctx)

	var payload rechargemq.RechargeEventPayload
	if err := json.Unmarshal(msg.Body, &payload); err != nil {
		helper.LogError(logger, helper.OpMQConsumer, "unmarshal RechargeEvent payload failed", err, map[string]interface{}{
			"message_id": msg.MsgId,
			"body":       string(msg.Body),
		})
		return nil // 丢弃这条，避免一直重试
	}

	if payload.OrderNo == "" || payload.UserID == 0 || payload.Amount <= 0 {
		helper.LogError(logger, helper.OpMQConsumer, "invalid RechargeEvent payload", nil, map[string]interface{}{
			"message_id": msg.MsgId,
			"order_no":   payload.OrderNo,
			"user_id":    payload.UserID,
			"amount":     payload.Amount,
		})
		return nil
	}

	tag := msg.GetTags()
	helper.LogInfo(logger, helper.OpMQConsumer, "processing event", map[string]interface{}{
		"tag":      tag,
		"order_no": payload.OrderNo,
		"user_id":  payload.UserID,
		"amount":   payload.Amount,
	})

	switch tag {
	case rechargemq.EventTypeRechargeSuccess():
		return handleRechargeSuccess(ctx, svcCtx, &payload)
	case rechargemq.EventTypeRechargeClosed():
		return handleRechargeStatusUpdate(ctx, svcCtx, &payload, model.RechargeStatusClosed)
	case rechargemq.EventTypeRechargeFailed():
		return handleRechargeStatusUpdate(ctx, svcCtx, &payload, model.RechargeStatusFailed)
	default:
		// 未识别的 tag 忽略
		helper.LogInfo(logger, helper.OpMQConsumer, "skipping unknown tag", map[string]interface{}{
			"tag":        tag,
			"message_id": msg.MsgId,
		})
		return nil
	}
}

func handleRechargeSuccess(ctx context.Context, svcCtx *svc.ServiceContext, payload *rechargemq.RechargeEventPayload) error {
	logger := logx.WithContext(ctx)

	helper.LogInfo(logger, helper.OpMQConsumer, "processing success event", map[string]interface{}{
		"order_no": payload.OrderNo,
		"user_id":  payload.UserID,
		"amount":   payload.Amount,
	})

	// 先入账（幂等）
	rechargeLogic := logic.NewRechargeLogic(ctx, svcCtx)
	if _, err := rechargeLogic.Recharge(&user.RechargeRequest{
		UserId:     payload.UserID,
		Amount:     payload.Amount,
		BizOrderId: payload.OrderNo,
		Remark:     payload.Remark,
	}); err != nil {
		helper.LogError(logger, helper.OpMQConsumer, "recharge wallet failed", err, map[string]interface{}{
			"order_no": payload.OrderNo,
			"user_id":  payload.UserID,
			"amount":   payload.Amount,
		})
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
		helper.LogError(logger, helper.OpMQConsumer, "update order status failed", err, map[string]interface{}{
			"order_no": payload.OrderNo,
		})
		return err
	}

	helper.LogSuccess(logger, helper.OpMQConsumer, map[string]interface{}{
		"event":    "recharge_success",
		"order_no": payload.OrderNo,
		"user_id":  payload.UserID,
		"amount":   payload.Amount,
	})
	return nil
}

func handleRechargeStatusUpdate(ctx context.Context, svcCtx *svc.ServiceContext,
	payload *rechargemq.RechargeEventPayload, status int) error {
	logger := logx.WithContext(ctx)

	helper.LogInfo(logger, helper.OpMQConsumer, "processing status update", map[string]interface{}{
		"order_no": payload.OrderNo,
		"user_id":  payload.UserID,
		"status":   status,
	})

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
		helper.LogError(logger, helper.OpMQConsumer, "update order status failed", err, map[string]interface{}{
			"order_no": payload.OrderNo,
			"status":   status,
		})
		return err
	}

	helper.LogSuccess(logger, helper.OpMQConsumer, map[string]interface{}{
		"event":    "recharge_status_update",
		"order_no": payload.OrderNo,
		"status":   status,
	})
	return nil
}
