package logic

import (
	"context"
	"encoding/json"

	"SLGaming/back/services/user/internal/helper"
	"SLGaming/back/services/user/internal/model"
	"SLGaming/back/services/user/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

// RefundLogic 退款逻辑（幂等）
// 同一笔业务单号（bizOrderId）多次调用，只会实际退款一次，并写入退款成功事件到 Outbox
type RefundLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRefundLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RefundLogic {
	return &RefundLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Refund 为指定用户执行退款（amount 为正数），bizOrderId 用于幂等控制
// orderNo 用于在退款成功事件中携带订单号，便于订单服务更新状态
func (l *RefundLogic) Refund(userID uint64, amount int64, bizOrderID, orderNo, remark string) error {
	if userID == 0 {
		return status.Error(codes.InvalidArgument, "user_id is required")
	}
	if amount <= 0 {
		return status.Error(codes.InvalidArgument, "amount must be positive")
	}
	if bizOrderID == "" {
		return status.Error(codes.InvalidArgument, "biz_order_id is required for refund idempotency")
	}

	// 使用钱包服务统一处理，并在事务中写入 Outbox 事件
	walletService := helper.NewWalletService(l.svcCtx.DB())
	_, err := walletService.UpdateBalance(l.ctx, &helper.WalletUpdateRequest{
		UserID:     userID,
		Amount:     amount,
		Type:       helper.WalletOpRefund,
		BizOrderID: bizOrderID,
		Remark:     remark,
		Logger:     l.Logger,
		AfterTransaction: func(tx *gorm.DB) error {
			// 写入 ORDER_REFUND_SUCCEEDED 事件到 Outbox（即使已退款，重复事件对订单侧是幂等的）
			payload := map[string]any{
				"order_id":     0, // 如需精确到 ID，可在调用处扩展参数
				"order_no":     orderNo,
				"biz_order_id": bizOrderID,
				"user_id":      userID,
				"amount":       amount,
			}
			body, err := json.Marshal(payload)
			if err != nil {
				helper.LogError(l.Logger, helper.OpRefund, "marshal refund outbox payload failed", err, map[string]interface{}{
					"user_id":     userID,
					"biz_order_id": bizOrderID,
				})
				return nil // 不阻塞退款流程
			}
			evt := &model.UserEventOutbox{
				EventType: "ORDER_REFUND_SUCCEEDED",
				Payload:   string(body),
				Status:    "PENDING",
			}
			if err := tx.Create(evt).Error; err != nil {
				helper.LogError(l.Logger, helper.OpRefund, "create user event outbox failed", err, map[string]interface{}{
					"user_id":     userID,
					"biz_order_id": bizOrderID,
				})
				return err
			}
			return nil
		},
	})
	if err != nil {
		return err
	}

	return nil
}
