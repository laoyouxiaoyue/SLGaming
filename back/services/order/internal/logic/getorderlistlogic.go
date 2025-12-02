package logic

import (
	"context"

	"SLGaming/back/services/order/internal/model"
	"SLGaming/back/services/order/internal/svc"
	"SLGaming/back/services/order/order"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type GetOrderListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetOrderListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOrderListLogic {
	return &GetOrderListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetOrderListLogic) GetOrderList(in *order.GetOrderListRequest) (*order.GetOrderListResponse, error) {
	db := l.svcCtx.DB.WithContext(l.ctx)

	var (
		orders []model.Order
		total  int64
	)

	query := db.Model(&model.Order{})

	if in.GetBossId() != 0 {
		query = query.Where("boss_id = ?", in.GetBossId())
	}
	if in.GetCompanionId() != 0 {
		query = query.Where("companion_id = ?", in.GetCompanionId())
	}
	if in.GetStatus() != 0 {
		query = query.Where("status = ?", in.GetStatus())
	}

	if err := query.Count(&total).Error; err != nil && err != gorm.ErrRecordNotFound {
		l.Errorf("count orders failed: %v", err)
		return nil, status.Error(codes.Internal, "count orders failed")
	}

	page := in.GetPage()
	pageSize := in.GetPageSize()
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize
	if err := query.Order("id DESC").Offset(int(offset)).Limit(int(pageSize)).Find(&orders).Error; err != nil && err != gorm.ErrRecordNotFound {
		l.Errorf("list orders failed: %v", err)
		return nil, status.Error(codes.Internal, "list orders failed")
	}

	var list []*order.OrderInfo
	for i := range orders {
		list = append(list, toOrderInfo(&orders[i]))
	}

	return &order.GetOrderListResponse{
		Orders:   list,
		Total:    int32(total),
		Page:     page,
		PageSize: pageSize,
	}, nil
}
