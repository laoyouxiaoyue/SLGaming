// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package order

import (
	"SLGaming/back/services/gateway/internal/types"
	"SLGaming/back/services/order/orderclient"
)

// toOrderInfo 将 RPC 的 OrderInfo 转为网关层的 OrderInfo
func toOrderInfo(o *orderclient.OrderInfo) types.OrderInfo {
	if o == nil {
		return types.OrderInfo{}
	}
	return types.OrderInfo{
		Id:           o.Id,
		OrderNo:      o.OrderNo,
		BossId:       o.BossId,
		CompanionId:  o.CompanionId,
		GameName:     o.GameName,
		GameMode:     o.GameMode,
		Duration:     o.DurationMinutes,
		PricePerHour: o.PricePerHour,
		TotalAmount:  o.TotalAmount,
		Status:       o.Status,
		CreatedAt:    o.CreatedAt,
		PaidAt:       o.PaidAt,
		AcceptedAt:   o.AcceptedAt,
		StartAt:      o.StartAt,
		CompletedAt:  o.CompletedAt,
		CancelledAt:  o.CancelledAt,
		Rating:       o.Rating,
		Comment:      o.Comment,
		CancelReason: o.CancelReason,
	}
}
