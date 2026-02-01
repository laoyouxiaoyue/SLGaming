package logic

import (
	"context"
	"errors"

	"SLGaming/back/services/user/internal/model"
	"SLGaming/back/services/user/internal/svc"
	"SLGaming/back/services/user/user"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type CreateRechargeOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateRechargeOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateRechargeOrderLogic {
	return &CreateRechargeOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateRechargeOrderLogic) CreateRechargeOrder(in *user.CreateRechargeOrderRequest) (*user.CreateRechargeOrderResponse, error) {
	if in.GetUserId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	if in.GetOrderNo() == "" {
		return nil, status.Error(codes.InvalidArgument, "order_no is required")
	}
	if in.GetAmount() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "amount must be positive")
	}

	db := l.svcCtx.DB()
	if db == nil {
		return nil, status.Error(codes.Internal, "db not initialized")
	}

	payType := in.GetPayType()
	if payType == "" {
		payType = "alipay"
	}

	var existing model.RechargeOrder
	err := db.Where("order_no = ?", in.GetOrderNo()).First(&existing).Error
	if err == nil {
		updates := map[string]interface{}{
			"user_id":  in.GetUserId(),
			"amount":   in.GetAmount(),
			"status":   int(in.GetStatus()),
			"pay_type": payType,
			"remark":   in.GetRemark(),
		}
		if err := db.Model(&existing).Updates(updates).Error; err != nil {
			return nil, status.Error(codes.Internal, "update recharge order failed")
		}
		return &user.CreateRechargeOrderResponse{Success: true}, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, status.Error(codes.Internal, "query recharge order failed")
	}

	order := &model.RechargeOrder{
		UserID:  in.GetUserId(),
		OrderNo: in.GetOrderNo(),
		Amount:  in.GetAmount(),
		Status:  int(in.GetStatus()),
		PayType: payType,
		Remark:  in.GetRemark(),
	}
	if err := db.Create(order).Error; err != nil {
		return nil, status.Error(codes.Internal, "create recharge order failed")
	}

	return &user.CreateRechargeOrderResponse{Success: true}, nil
}
