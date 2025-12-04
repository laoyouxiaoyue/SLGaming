package logic

import (
	"context"
	"encoding/json"

	"SLGaming/back/services/order/internal/model"
	"SLGaming/back/services/order/internal/svc"
	"SLGaming/back/services/order/order"
	"SLGaming/back/services/user/userclient"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

// orderPaymentPendingEventPayload 订单支付待处理事件负载结构
type orderPaymentPendingEventPayload struct {
	OrderID    uint64 `json:"order_id"`
	OrderNo    string `json:"order_no"`
	BossID     uint64 `json:"boss_id"`
	Amount     int64  `json:"amount"`
	BizOrderID string `json:"biz_order_id"`
}

type CreateOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateOrderLogic {
	return &CreateOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateOrderLogic) CreateOrder(in *order.CreateOrderRequest) (*order.CreateOrderResponse, error) {
	if in.GetBossId() == 0 || in.GetCompanionId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "boss_id and companion_id are required")
	}
	if in.GetDurationMinutes() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "duration_minutes must be positive")
	}

	if l.svcCtx.UserRPC == nil {
		return nil, status.Error(codes.FailedPrecondition, "user rpc client not initialized")
	}

	db := l.svcCtx.DB.WithContext(l.ctx)

	// 1. 查询陪玩当前价格
	cpResp, err := l.svcCtx.UserRPC.GetCompanionProfile(l.ctx, &userclient.GetCompanionProfileRequest{
		UserId: in.GetCompanionId(),
	})
	if err != nil {
		l.Errorf("get companion profile failed: %v", err)
		return nil, status.Error(codes.Internal, "get companion profile failed")
	}
	if cpResp == nil || cpResp.Profile == nil {
		return nil, status.Error(codes.FailedPrecondition, "companion profile not found")
	}

	pricePerHour := cpResp.Profile.PricePerHour
	if pricePerHour <= 0 {
		return nil, status.Error(codes.FailedPrecondition, "invalid companion price")
	}

	// 金额按照时长一次性计算（分钟 -> 小时），这里简单用向上取整
	durationMinutes := in.GetDurationMinutes()
	hours := (durationMinutes + 59) / 60
	totalAmount := pricePerHour * int64(hours)

	o := &model.Order{
		BossID:          in.GetBossId(),
		CompanionID:     in.GetCompanionId(),
		GameName:        in.GetGameName(),
		GameMode:        in.GetGameMode(),
		DurationMinutes: in.GetDurationMinutes(),
		PricePerHour:    pricePerHour,
		TotalAmount:     totalAmount,
		Status:          model.OrderStatusCreated, // 先创建为待支付状态
	}
	o.OrderNo = generateOrderNo(o.BossID)

	// 2. 在一个事务中：创建订单 + 写入 ORDER_PAYMENT_PENDING 事件到 Outbox
	// 使用 Outbox 模式确保订单创建和支付事件的原子性
	if err := db.Transaction(func(tx *gorm.DB) error {
		// 创建订单记录
		if err := tx.Create(o).Error; err != nil {
			return err
		}

		// 构造支付事件负载
		payload := orderPaymentPendingEventPayload{
			OrderID:    o.ID,
			OrderNo:    o.OrderNo,
			BossID:     o.BossID,
			Amount:     o.TotalAmount,
			BizOrderID: o.OrderNo,
		}
		payloadJSON, err := json.Marshal(payload)
		if err != nil {
			l.Errorf("marshal payment pending payload failed: %v", err)
			return status.Error(codes.Internal, "marshal payment event failed")
		}

		// 写入 ORDER_PAYMENT_PENDING 事件到 outbox
		evt := &model.OrderEventOutbox{
			EventType: "ORDER_PAYMENT_PENDING",
			Payload:   string(payloadJSON),
			Status:    "PENDING",
		}

		if err := tx.Create(evt).Error; err != nil {
			l.Errorf("create payment pending event outbox failed: %v", err)
			return status.Error(codes.Internal, "create payment event failed")
		}

		return nil
	}); err != nil {
		l.Errorf("create order with payment event failed: %v", err)
		return nil, err
	}

	return &order.CreateOrderResponse{
		Order: toOrderInfo(o),
	}, nil
}
