package logic

import (
	"context"
	"time"

	"SLGaming/back/services/order/internal/model"
	"SLGaming/back/services/order/internal/svc"
	"SLGaming/back/services/order/order"
	"SLGaming/back/services/user/userclient"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type AcceptOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAcceptOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AcceptOrderLogic {
	return &AcceptOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *AcceptOrderLogic) AcceptOrder(in *order.AcceptOrderRequest) (*order.AcceptOrderResponse, error) {
	if in.GetOrderId() == 0 || in.GetCompanionId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "order_id and companion_id are required")
	}

	db := l.svcCtx.DB.WithContext(l.ctx)

	var o model.Order
	if err := db.Where("id = ?", in.GetOrderId()).First(&o).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Error(codes.NotFound, "order not found")
		}
		l.Errorf("get order failed: %v", err)
		return nil, status.Error(codes.Internal, "get order failed")
	}

	// 只能由指定陪玩接单，且必须是待接单状态
	if o.CompanionID != in.GetCompanionId() {
		return nil, status.Error(codes.PermissionDenied, "not allowed to accept this order")
	}
	if o.Status != model.OrderStatusPaid && o.Status != model.OrderStatusCreated {
		return nil, status.Error(codes.FailedPrecondition, "order is not pending for accept")
	}

	now := time.Now()
	o.Status = model.OrderStatusAccepted
	o.AcceptedAt = &now

	if err := db.Save(&o).Error; err != nil {
		l.Errorf("accept order failed: %v", err)
		return nil, status.Error(codes.Internal, "accept order failed")
	}

	// 更新陪玩状态为忙碌（2），
	if l.svcCtx.UserRPC != nil {
		_, err := l.svcCtx.UserRPC.UpdateCompanionProfile(l.ctx, &userclient.UpdateCompanionProfileRequest{
			UserId: in.GetCompanionId(),
			Status: 2, // 忙碌
		})
		if err != nil {
			if st, ok := status.FromError(err); ok {
				l.Errorf("update companion status to busy failed: code=%v, msg=%s", st.Code(), st.Message())
			} else {
				l.Errorf("update companion status to busy failed: %v", err)
			}
		}
	}

	return &order.AcceptOrderResponse{
		Order: toOrderInfo(&o),
	}, nil
}
