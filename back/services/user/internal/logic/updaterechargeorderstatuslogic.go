package logic

import (
	"context"
	"errors"
	"time"

	"SLGaming/back/services/user/internal/model"
	"SLGaming/back/services/user/internal/svc"
	"SLGaming/back/services/user/user"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type UpdateRechargeOrderStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateRechargeOrderStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateRechargeOrderStatusLogic {
	return &UpdateRechargeOrderStatusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateRechargeOrderStatusLogic) UpdateRechargeOrderStatus(in *user.UpdateRechargeOrderStatusRequest) (*user.UpdateRechargeOrderStatusResponse, error) {
	if in.GetOrderNo() == "" {
		return nil, status.Error(codes.InvalidArgument, "order_no is required")
	}
	if in.GetStatus() < 0 || in.GetStatus() > 3 {
		return nil, status.Error(codes.InvalidArgument, "invalid status")
	}

	db := l.svcCtx.DB()
	if db == nil {
		return nil, status.Error(codes.Internal, "db not initialized")
	}

	var order model.RechargeOrder
	err := db.Where("order_no = ?", in.GetOrderNo()).First(&order).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "recharge order not found")
		}
		return nil, status.Error(codes.Internal, "query recharge order failed")
	}

	if order.Status == model.RechargeStatusSuccess && int(in.GetStatus()) == model.RechargeStatusSuccess {
		return &user.UpdateRechargeOrderStatusResponse{Success: true}, nil
	}

	updates := map[string]interface{}{
		"status": int(in.GetStatus()),
	}
	if in.GetTradeNo() != "" {
		updates["trade_no"] = in.GetTradeNo()
	}
	if in.GetRemark() != "" {
		updates["remark"] = in.GetRemark()
	}
	if in.GetPaidAt() > 0 {
		paidAt := time.Unix(in.GetPaidAt(), 0)
		updates["paid_at"] = &paidAt
	}

	if err := db.Model(&order).Updates(updates).Error; err != nil {
		return nil, status.Error(codes.Internal, "update recharge order failed")
	}

	return &user.UpdateRechargeOrderStatusResponse{Success: true}, nil
}
