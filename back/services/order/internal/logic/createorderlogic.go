package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"SLGaming/back/pkg/lock"
	"SLGaming/back/services/order/internal/helper"
	orderioc "SLGaming/back/services/order/internal/ioc"
	"SLGaming/back/services/order/internal/model"
	orderMQ "SLGaming/back/services/order/internal/mq"
	"SLGaming/back/services/order/internal/svc"
	"SLGaming/back/services/order/internal/tx"
	"SLGaming/back/services/order/order"
	"SLGaming/back/services/user/userclient"

	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type CreateOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateOrderLogic {
	return &CreateOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// reinitUserRPC 尝试重新初始化 UserRPC
func (l *CreateOrderLogic) reinitUserRPC() error {
	cfg := l.svcCtx.Config
	if cfg.Upstream.UserService == "" {
		helper.LogError(l.Logger, helper.OpCreateOrder, "user service not configured", nil, nil)
		return fmt.Errorf("user service not configured")
	}

	// 使用 Consul 服务发现重新解析用户服务端点
	endpoints, err := orderioc.ResolveServiceEndpoints(cfg.Consul, cfg.Upstream.UserService)
	if err != nil {
		helper.LogError(l.Logger, helper.OpCreateOrder, "resolve user service endpoints failed", err, nil)
		return err
	}

	// 创建新的 RPC 客户端
	client := zrpc.MustNewClient(zrpc.RpcClientConf{
		Endpoints: endpoints,
		NonBlock:  true,
	})

	// 初始化 UserRPC
	l.svcCtx.UserRPC = userclient.NewUser(client)
	helper.LogInfo(l.Logger, helper.OpCreateOrder, "reinit user rpc client success", map[string]interface{}{
		"service":   cfg.Upstream.UserService,
		"endpoints": endpoints,
	})

	return nil
}

func (l *CreateOrderLogic) CreateOrder(in *order.CreateOrderRequest) (*order.CreateOrderResponse, error) {
	helper.LogRequest(l.Logger, helper.OpCreateOrder, map[string]interface{}{
		"boss_id":        in.GetBossId(),
		"companion_id":   in.GetCompanionId(),
		"duration_hours": in.GetDurationHours(),
		"game_name":      in.GetGameName(),
	})

	if in.GetBossId() == 0 || in.GetCompanionId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "boss_id and companion_id are required")
	}
	if in.GetBossId() == in.GetCompanionId() {
		return nil, status.Error(codes.InvalidArgument, "cannot create order for yourself")
	}
	if in.GetDurationHours() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "duration_hours must be positive")
	}

	if l.svcCtx.UserRPC == nil {
		// 尝试重新初始化 UserRPC
		if err := l.reinitUserRPC(); err != nil {
			return nil, status.Error(codes.FailedPrecondition, "user rpc client not initialized")
		}
	}

	// 使用双重分布式锁防止并发创建订单
	// 第一层锁：基于 boss_id 和 companion_id，防止同一老板对同一陪玩并发创建多个订单
	// 第二层锁：基于 companion_id，防止多个老板同时对同一陪玩下单（串行化处理）
	bossCompanionLockKey := fmt.Sprintf("create_order:%d:%d", in.GetBossId(), in.GetCompanionId())
	companionLockKey := fmt.Sprintf("companion_order:%d", in.GetCompanionId())
	lockValue := uuid.New().String()

	// 如果分布式锁未初始化，直接执行（降级处理）
	if l.svcCtx.DistributedLock == nil {
		helper.LogWarning(l.Logger, helper.OpCreateOrder, "distributed lock not initialized, skipping lock", nil)
		return l.doCreateOrder(in)
	}

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
		helper.LogError(l.Logger, helper.OpCreateOrder, "create order with lock failed", err, nil)
		return nil, status.Error(codes.Internal, "create order failed")
	}

	return result, createErr
}

// doCreateOrder 执行实际的订单创建逻辑（不加锁）
func (l *CreateOrderLogic) doCreateOrder(in *order.CreateOrderRequest) (*order.CreateOrderResponse, error) {
	// 1. 查询陪玩当前价格
	cpResp, err := l.svcCtx.UserRPC.GetCompanionProfile(l.ctx, &userclient.GetCompanionProfileRequest{
		UserId: in.GetCompanionId(),
	})
	if err != nil {
		helper.LogError(l.Logger, helper.OpCreateOrder, "get companion profile failed", err, map[string]interface{}{
			"companion_id": in.GetCompanionId(),
		})
		return nil, status.Error(codes.Internal, "get companion profile failed")
	}
	if cpResp == nil || cpResp.Profile == nil {
		return nil, status.Error(codes.FailedPrecondition, "companion profile not found")
	}

	pricePerHour := cpResp.Profile.PricePerHour
	if pricePerHour <= 0 {
		return nil, status.Error(codes.FailedPrecondition, "invalid companion price")
	}

	// 金额按照小时计算
	durationHours := in.GetDurationHours()
	totalAmount := pricePerHour * int64(durationHours)

	// 2. 检查老板钱包余额是否足够
	walletResp, err := l.svcCtx.UserRPC.GetWallet(l.ctx, &userclient.GetWalletRequest{
		UserId: in.GetBossId(),
	})
	if err != nil {
		helper.LogError(l.Logger, helper.OpCreateOrder, "get boss wallet failed", err, map[string]interface{}{
			"boss_id": in.GetBossId(),
		})
		return nil, status.Error(codes.Internal, "get wallet failed")
	}
	if walletResp == nil || walletResp.Wallet == nil {
		return nil, status.Error(codes.FailedPrecondition, "wallet not found, please create wallet first")
	}

	currentBalance := walletResp.Wallet.Balance
	if currentBalance < totalAmount {
		helper.LogWarning(l.Logger, helper.OpCreateOrder, "insufficient balance", map[string]interface{}{
			"boss_id":         in.GetBossId(),
			"current_balance": currentBalance,
			"required_amount": totalAmount,
		})
		return nil, status.Error(codes.ResourceExhausted,
			"insufficient handsome coins, current balance is insufficient for this order")
	}

	orderNo := generateOrderNo(in.GetBossId())

	// 3. 使用 RocketMQ 事务消息发送 ORDER_PAYMENT_PENDING，并在本地事务中创建订单
	if l.svcCtx.OrderEventTxProducer == nil {
		// 不提供降级方案，事务 Producer 必须初始化成功
		return nil, status.Error(codes.FailedPrecondition, "order transaction producer not initialized")
	}

	payload := &tx.OrderPaymentPendingPayload{
		OrderNo:       orderNo,
		BossID:        in.GetBossId(),
		Amount:        totalAmount,
		BizOrderID:    orderNo,
		CompanionID:   in.GetCompanionId(),
		GameName:      in.GetGameName(),
		DurationHours: durationHours,
		PricePerHour:  pricePerHour,
	}

	// 构造事务消息
	msgBody, err := json.Marshal(payload)
	if err != nil {
		helper.LogError(l.Logger, helper.OpCreateOrder, "marshal payment pending payload failed", err, nil)
		return nil, status.Error(codes.Internal, "marshal payment event failed")
	}
	msg := primitive.NewMessage(orderMQ.OrderEventTopic(), msgBody)
	msg.WithTag(orderMQ.EventTypePaymentPending())

	// 发送 RocketMQ 事务消息
	// 注意：SendMessageInTransaction 会同步执行 ExecuteLocalTransaction（即 ExecuteCreateOrderTx）
	// 如果 ExecuteCreateOrderTx 返回 error，本地事务会回滚，消息也会回滚
	// 如果返回 nil error，说明半消息发送成功，但本地事务是否成功需要通过查询订单确认
	txRes, err := l.svcCtx.OrderEventTxProducer.SendMessageInTransaction(l.ctx, msg)
	if err != nil {
		// SendMessageInTransaction 返回 error 可能有两种情况：
		// 1. 发送半消息失败（网络问题、Broker 不可用等）
		// 2. 本地事务执行失败（ExecuteCreateOrderTx 返回 error）
		// 无论哪种情况，订单都不会创建，直接返回错误
		helper.LogError(l.Logger, helper.OpCreateOrder, "send transactional message failed", err, map[string]interface{}{
			"order_no": orderNo,
			"result":   fmt.Sprintf("%+v", txRes),
		})
		return nil, status.Error(codes.Internal, "create order failed: transaction message send failed")
	}

	// 此时本地事务（ExecuteOrderTx -> ExecuteCreateOrderTx）已经执行完成
	// 但 SendMessageInTransaction 返回 nil error 只表示半消息发送成功，
	// 本地事务是否成功需要通过查询订单确认（因为 ExecuteLocalTransaction 可能返回 RollbackMessageState）
	db := l.svcCtx.DB.WithContext(l.ctx)
	var o model.Order
	if err := db.Where("order_no = ?", orderNo).First(&o).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// 订单不存在，说明本地事务回滚了（ExecuteCreateOrderTx 返回了 error）
			// 可能的原因：数据库错误、数据校验失败、幂等检查发现重复等
			helper.LogError(l.Logger, helper.OpCreateOrder, "create order transaction rolled back", err, map[string]interface{}{
				"order_no": orderNo,
				"result":   fmt.Sprintf("%+v", txRes),
			})
			return nil, status.Error(codes.Internal, "create order failed: local transaction rolled back")
		}
		// 数据库查询错误（非记录不存在）
		helper.LogError(l.Logger, helper.OpCreateOrder, "query order after transactional message failed", err, map[string]interface{}{
			"order_no": orderNo,
		})
		return nil, status.Error(codes.Internal, "create order failed: query order error")
	}

	helper.LogSuccess(l.Logger, helper.OpCreateOrder, map[string]interface{}{
		"order_id":     o.ID,
		"order_no":     o.OrderNo,
		"boss_id":      o.BossID,
		"companion_id": o.CompanionID,
		"total_amount": o.TotalAmount,
	})

	return &order.CreateOrderResponse{Order: toOrderInfo(&o)}, nil
}
