package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"SLGaming/back/pkg/lock"
	"SLGaming/back/services/order/internal/model"
	"SLGaming/back/services/order/internal/svc"
	"SLGaming/back/services/order/order"
	"SLGaming/back/services/user/userclient"

	"github.com/google/uuid"
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

	// 使用双重分布式锁防止并发创建订单
	// 第一层锁：基于 boss_id 和 companion_id，防止同一老板对同一陪玩并发创建多个订单
	// 第二层锁：基于 companion_id，防止多个老板同时对同一陪玩下单（串行化处理）
	bossCompanionLockKey := fmt.Sprintf("create_order:%d:%d", in.GetBossId(), in.GetCompanionId())
	companionLockKey := fmt.Sprintf("companion_order:%d", in.GetCompanionId())
	lockValue := uuid.New().String()

	// 如果分布式锁未初始化，直接执行（降级处理）
	if l.svcCtx.DistributedLock == nil {
		l.Infof("distributed lock not initialized, skipping lock for order creation")
		return l.doCreateOrder(in)
	}

	// 使用双重分布式锁执行订单创建
	var result *order.CreateOrderResponse
	var createErr error

	lockOptions := &lock.LockOptions{
		TTL:           30,                     // 锁过期时间 30 秒
		RetryInterval: 100 * time.Millisecond, // 重试间隔 100ms
		MaxWaitTime:   5 * time.Second,        // 最大等待时间 5 秒
	}

	// 先获取陪玩级别的锁（防止多个老板同时下单）
	err := l.svcCtx.DistributedLock.WithLock(l.ctx, companionLockKey, lockValue, lockOptions, func() error {
		// 在陪玩锁内，再获取老板-陪玩级别的锁（防止同一老板重复下单）
		bossLockValue := uuid.New().String()
		return l.svcCtx.DistributedLock.WithLock(l.ctx, bossCompanionLockKey, bossLockValue, lockOptions, func() error {
			result, createErr = l.doCreateOrder(in)
			return createErr
		})
	})

	if err != nil {
		if err == context.DeadlineExceeded || err == context.Canceled {
			return nil, status.Error(codes.DeadlineExceeded, "acquire lock timeout, please try again later")
		}
		l.Errorf("create order with lock failed: %v", err)
		return nil, status.Error(codes.Internal, "create order failed")
	}

	return result, createErr
}

// doCreateOrder 执行实际的订单创建逻辑（不加锁）
func (l *CreateOrderLogic) doCreateOrder(in *order.CreateOrderRequest) (*order.CreateOrderResponse, error) {
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

	// 2. 检查老板钱包余额是否足够
	walletResp, err := l.svcCtx.UserRPC.GetWallet(l.ctx, &userclient.GetWalletRequest{
		UserId: in.GetBossId(),
	})
	if err != nil {
		l.Errorf("get boss wallet failed: %v", err)
		return nil, status.Error(codes.Internal, "get wallet failed")
	}
	if walletResp == nil || walletResp.Wallet == nil {
		return nil, status.Error(codes.FailedPrecondition, "wallet not found, please create wallet first")
	}

	currentBalance := walletResp.Wallet.Balance
	if currentBalance < totalAmount {
		l.Infof("create order failed: insufficient balance, boss_id=%d, current_balance=%d, required_amount=%d",
			in.GetBossId(), currentBalance, totalAmount)
		return nil, status.Error(codes.ResourceExhausted,
			"insufficient handsome coins, current balance is insufficient for this order")
	}

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

	// 3. 在一个事务中：创建订单 + 写入 ORDER_PAYMENT_PENDING 事件到 Outbox
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
