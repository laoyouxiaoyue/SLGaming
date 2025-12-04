package logic

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"SLGaming/back/services/user/internal/model"
	"SLGaming/back/services/user/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

	db := l.svcCtx.DB().WithContext(l.ctx)

	var wallet model.UserWallet

	// 在事务中进行余额更新、流水记录和 Outbox 事件写入，保证原子性
	if err := db.Transaction(func(tx *gorm.DB) error {
		// 1. 幂等检查：是否已经存在 REFUND + biz_order_id 的流水
		var existed model.WalletTransaction
		if err := tx.
			Where("type = ? AND biz_order_id = ?", "REFUND", bizOrderID).
			First(&existed).Error; err == nil {
			// 已经退款过了，直接视为成功
			return nil
		} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		// 2. 加锁读取 / 初始化钱包
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ?", userID).
			First(&wallet).Error

		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 没有钱包记录时，为用户创建一个新的钱包
			wallet = model.UserWallet{
				UserID:        userID,
				Balance:       0,
				FrozenBalance: 0,
			}
			if err := tx.Create(&wallet).Error; err != nil {
				return err
			}
		} else if err != nil {
			return err
		}

		before := wallet.Balance
		after := before + amount
		wallet.Balance = after

		if err := tx.Save(&wallet).Error; err != nil {
			return err
		}

		tr := &model.WalletTransaction{
			UserID:        userID,
			WalletID:      wallet.ID,
			ChangeAmount:  amount,
			BeforeBalance: before,
			AfterBalance:  after,
			Type:          "REFUND",
			BizOrderID:    bizOrderID,
			Remark:        remark,
		}

		if err := tx.Create(tr).Error; err != nil {
			// 处理并发情况下的唯一约束冲突（说明已有其它请求完成退款）
			if strings.Contains(err.Error(), "Duplicate entry") {
				// 已有相同 REFUND 流水，说明之前已成功退款，但可能 Outbox 未写入，继续向下执行写入事件
			} else {
				return err
			}
		}

		// 3. 写入 ORDER_REFUND_SUCCEEDED 事件到 Outbox（即使已退款，重复事件对订单侧是幂等的）
		payload := map[string]any{
			"order_id":     0, // 如需精确到 ID，可在调用处扩展参数
			"order_no":     orderNo,
			"biz_order_id": bizOrderID,
			"user_id":      userID,
			"amount":       amount,
		}
		body, err := json.Marshal(payload)
		if err != nil {
			l.Errorf("marshal refund outbox payload failed: %v", err)
			return nil
		}
		evt := &model.UserEventOutbox{
			EventType: "ORDER_REFUND_SUCCEEDED",
			Payload:   string(body),
			Status:    "PENDING",
		}
		if err := tx.Create(evt).Error; err != nil {
			l.Errorf("create user event outbox failed: %v", err)
			return err
		}

		return nil
	}); err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	return nil
}
