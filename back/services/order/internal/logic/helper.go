package logic

import (
	"fmt"
	"time"

	"SLGaming/back/pkg/snowflake"
	"SLGaming/back/services/order/internal/model"
	"SLGaming/back/services/order/order"
)

// toOrderInfo 模型转 RPC 结构
func toOrderInfo(o *model.Order) *order.OrderInfo {
	if o == nil {
		return nil
	}

	var paidAt, acceptedAt, startAt, completedAt, cancelledAt int64
	if o.PaidAt != nil {
		paidAt = o.PaidAt.Unix()
	}
	if o.AcceptedAt != nil {
		acceptedAt = o.AcceptedAt.Unix()
	}
	if o.StartAt != nil {
		startAt = o.StartAt.Unix()
	}
	if o.CompletedAt != nil {
		completedAt = o.CompletedAt.Unix()
	}
	if o.CancelledAt != nil {
		cancelledAt = o.CancelledAt.Unix()
	}

	return &order.OrderInfo{
		Id:              o.ID,
		OrderNo:         o.OrderNo,
		BossId:          o.BossID,
		CompanionId:     o.CompanionID,
		GameName:        o.GameName,
		DurationMinutes: o.DurationMinutes,
		PricePerHour:    o.PricePerHour,
		TotalAmount:     o.TotalAmount,
		Status:          o.Status,
		CreatedAt:       o.CreatedAt.Unix(),
		PaidAt:          paidAt,
		AcceptedAt:      acceptedAt,
		StartAt:         startAt,
		CompletedAt:     completedAt,
		CancelledAt:     cancelledAt,
		Rating:          o.Rating,
		Comment:         o.Comment,
		CancelReason:    o.CancelReason,
	}
}

// generateOrderNo 生成订单号：
// 格式：YYYYMMDD + 6位序列（来自雪花ID后6位），例如 20250101 123456
func generateOrderNo(bossID uint64) string {
	prefix := time.Now().Format("20060102")
	id := snowflake.GenID()
	seq := id % 1_000_000
	return fmt.Sprintf("%s%06d", prefix, seq)
}
