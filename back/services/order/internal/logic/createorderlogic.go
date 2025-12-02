package logic

import (
	"context"

	"SLGaming/back/services/order/internal/model"
	"SLGaming/back/services/order/internal/svc"
	"SLGaming/back/services/order/order"
	"SLGaming/back/services/user/userclient"

	"github.com/zeromicro/go-zero/core/logx"
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
		Status:          model.OrderStatusPaid,
	}
	o.OrderNo = generateOrderNo(o.BossID)

	// 2. 先创建订单记录，再调用钱包扣款
	if err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(o).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		l.Errorf("create order failed: %v", err)
		return nil, status.Error(codes.Internal, "create order failed")
	}

	// 3. 调用 user 服务扣减老板钱包（消费帅币）
	_, err = l.svcCtx.UserRPC.Consume(l.ctx, &userclient.ConsumeRequest{
		UserId:     in.GetBossId(),
		Amount:     totalAmount,
		BizOrderId: o.OrderNo,
		Remark:     "order payment",
	})
	if err != nil {
		// 如果是业务错误（例如余额不足），直接透传
		if st, ok := status.FromError(err); ok {
			l.Errorf("consume wallet failed: code=%v, msg=%s", st.Code(), st.Message())
			return nil, err
		}
		l.Errorf("consume wallet failed: %v", err)
		return nil, status.Error(codes.Internal, "consume wallet failed")
	}

	return &order.CreateOrderResponse{
		Order: toOrderInfo(o),
	}, nil
}
