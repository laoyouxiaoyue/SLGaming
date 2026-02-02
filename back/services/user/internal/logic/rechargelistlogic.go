package logic

import (
	"context"
	"time"

	"SLGaming/back/services/user/internal/helper"
	"SLGaming/back/services/user/internal/model"
	"SLGaming/back/services/user/internal/svc"
	"SLGaming/back/services/user/user"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type RechargeListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRechargeListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RechargeListLogic {
	return &RechargeListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RechargeListLogic) RechargeList(in *user.RechargeListRequest) (*user.RechargeListResponse, error) {
	userID := in.GetUserId()
	if userID == 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	db := l.svcCtx.DB().WithContext(l.ctx)
	query := db.Model(&model.RechargeOrder{}).Where("user_id = ?", userID)

	statusFilter := int(in.GetStatus())
	if statusFilter >= 0 {
		if statusFilter > model.RechargeStatusClosed {
			return nil, status.Error(codes.InvalidArgument, "invalid status")
		}
		query = query.Where("status = ?", statusFilter)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	pagination := helper.NormalizePaginationWithDefault(in.GetPage(), in.GetPageSize(), 20)
	offset := (pagination.Page - 1) * pagination.PageSize

	var orders []model.RechargeOrder
	if err := query.Order("created_at DESC").Offset(offset).Limit(pagination.PageSize).Find(&orders).Error; err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	respOrders := make([]*user.RechargeOrderInfo, 0, len(orders))
	for i := range orders {
		paidAt := int64(0)
		if orders[i].PaidAt != nil {
			paidAt = orders[i].PaidAt.Unix()
		}
		createdAt := int64(0)
		if !orders[i].CreatedAt.IsZero() {
			createdAt = orders[i].CreatedAt.Unix()
		} else {
			createdAt = time.Now().Unix()
		}

		respOrders = append(respOrders, &user.RechargeOrderInfo{
			OrderNo:   orders[i].OrderNo,
			Status:    int32(orders[i].Status),
			Amount:    orders[i].Amount,
			PayType:   orders[i].PayType,
			TradeNo:   orders[i].TradeNo,
			PaidAt:    paidAt,
			CreatedAt: createdAt,
		})
	}

	return &user.RechargeListResponse{
		Orders:   respOrders,
		Total:    int32(total),
		Page:     int32(pagination.Page),
		PageSize: int32(pagination.PageSize),
	}, nil
}
