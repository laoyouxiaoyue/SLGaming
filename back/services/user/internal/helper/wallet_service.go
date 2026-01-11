package helper

import (
	"context"
	"errors"
	"strings"

	"SLGaming/back/services/user/internal/model"
	"SLGaming/back/services/user/user"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// WalletOperationType 钱包操作类型
type WalletOperationType string

const (
	WalletOpConsume  WalletOperationType = "CONSUME"  // 消费
	WalletOpRecharge WalletOperationType = "RECHARGE" // 充值
	WalletOpRefund   WalletOperationType = "REFUND"   // 退款
)

// AfterTransactionCallback 事务成功后的回调函数
// 在事务内执行，如果返回错误，整个事务会回滚
type AfterTransactionCallback func(tx *gorm.DB) error

// WalletUpdateRequest 钱包更新请求
type WalletUpdateRequest struct {
	UserID     uint64
	Amount     int64  // 金额（正数，实际增减由操作类型决定）
	Type       WalletOperationType
	BizOrderID string // 业务订单号，用于幂等控制
	Remark     string // 备注
	Logger     logx.Logger
	// AfterTransaction 事务成功后的回调（在事务内执行）
	AfterTransaction AfterTransactionCallback
}

// WalletUpdateResult 钱包更新结果
type WalletUpdateResult struct {
	Wallet *model.UserWallet
}

// WalletService 钱包服务
type WalletService struct {
	db *gorm.DB
}

// NewWalletService 创建钱包服务
func NewWalletService(db *gorm.DB) *WalletService {
	return &WalletService{db: db}
}

// UpdateBalance 更新钱包余额（统一处理充值、消费、退款）
func (s *WalletService) UpdateBalance(ctx context.Context, req *WalletUpdateRequest) (*WalletUpdateResult, error) {
	if req.UserID == 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	if req.Amount <= 0 {
		return nil, status.Error(codes.InvalidArgument, "amount must be positive")
	}

	var wallet model.UserWallet

	// 在事务中进行余额更新与流水记录
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. 幂等检查：如果提供了 BizOrderID，检查是否已经存在相同类型的流水
		if req.BizOrderID != "" {
			var existed model.WalletTransaction
			if err := tx.
				Where("type = ? AND biz_order_id = ?", string(req.Type), req.BizOrderID).
				First(&existed).Error; err == nil {
				// 已经操作过了，直接视为成功（幂等）
				if req.Logger != nil {
					LogInfo(req.Logger, LogOperation(req.Type), "idempotent: duplicate biz_order_id", map[string]interface{}{
						"user_id":              req.UserID,
						"biz_order_id":         req.BizOrderID,
						"existing_transaction_id": existed.ID,
					})
				}
				// 获取钱包信息用于返回
				if err := tx.Where("user_id = ?", req.UserID).First(&wallet).Error; err != nil {
					if req.Logger != nil {
						LogError(req.Logger, LogOperation(req.Type), "idempotent check failed to get wallet", err, map[string]interface{}{
							"user_id":     req.UserID,
							"biz_order_id": req.BizOrderID,
						})
					}
					return status.Error(codes.Internal, "failed to get wallet after idempotent check")
				}
				return nil
			} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				if req.Logger != nil {
					LogError(req.Logger, LogOperation(req.Type), "idempotent check failed", err, map[string]interface{}{
						"user_id":     req.UserID,
						"biz_order_id": req.BizOrderID,
					})
				}
				return status.Error(codes.Internal, "failed to check idempotency")
			}
		}

		// 2. 加锁读取钱包记录，避免并发更新问题
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ?", req.UserID).
			First(&wallet).Error

		// 处理钱包不存在的情况
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 消费操作要求钱包必须存在
			if req.Type == WalletOpConsume {
				if req.Logger != nil {
					LogError(req.Logger, LogOperation(req.Type), "wallet not found", nil, map[string]interface{}{
						"user_id": req.UserID,
						"amount":  req.Amount,
					})
				}
				return status.Error(codes.FailedPrecondition, "wallet not found, please create wallet first")
			}
			// 充值和退款操作：如果钱包不存在，创建一个新的钱包
			wallet = model.UserWallet{
				UserID:        req.UserID,
				Balance:       0,
				FrozenBalance: 0,
			}
			if err := tx.Create(&wallet).Error; err != nil {
				if req.Logger != nil {
					LogError(req.Logger, LogOperation(req.Type), "create wallet failed", err, map[string]interface{}{
						"user_id": req.UserID,
					})
				}
				return status.Error(codes.Internal, "failed to create wallet")
			}
		} else if err != nil {
			if req.Logger != nil {
				LogError(req.Logger, LogOperation(req.Type), "read wallet failed", err, map[string]interface{}{
					"user_id": req.UserID,
				})
			}
			return status.Error(codes.Internal, "failed to read wallet")
		}

		// 3. 余额检查（仅消费操作需要）
		if req.Type == WalletOpConsume {
			if wallet.Balance < req.Amount {
				if req.Logger != nil {
					LogWarning(req.Logger, LogOperation(req.Type), "insufficient balance", map[string]interface{}{
						"user_id":        req.UserID,
						"current_balance": wallet.Balance,
						"required_amount": req.Amount,
						"biz_order_id":    req.BizOrderID,
					})
				}
				return status.Error(codes.ResourceExhausted,
					"insufficient handsome coins, current balance is insufficient for this transaction")
			}
		}

		// 4. 计算余额变动
		before := wallet.Balance
		var after int64
		var changeAmount int64

		switch req.Type {
		case WalletOpConsume:
			// 消费：余额减少，变动金额为负数
			after = before - req.Amount
			changeAmount = -req.Amount
		case WalletOpRecharge, WalletOpRefund:
			// 充值和退款：余额增加，变动金额为正数
			after = before + req.Amount
			changeAmount = req.Amount
		default:
			return status.Error(codes.InvalidArgument, "invalid wallet operation type")
		}

		wallet.Balance = after

		// 5. 更新钱包余额
		if err := tx.Save(&wallet).Error; err != nil {
			if req.Logger != nil {
				LogError(req.Logger, LogOperation(req.Type), "update wallet balance failed", err, map[string]interface{}{
					"user_id":   req.UserID,
					"wallet_id": wallet.ID,
				})
			}
			return status.Error(codes.Internal, "failed to update wallet balance")
		}

		// 6. 创建交易流水记录
		tr := &model.WalletTransaction{
			UserID:        req.UserID,
			WalletID:      wallet.ID,
			ChangeAmount:  changeAmount,
			BeforeBalance: before,
			AfterBalance:  after,
			Type:          string(req.Type),
			BizOrderID:    req.BizOrderID,
			Remark:        req.Remark,
		}

		if err := tx.Create(tr).Error; err != nil {
			// 处理并发情况下的唯一约束冲突（说明已有其它请求完成操作）
			if strings.Contains(err.Error(), "Duplicate entry") {
				// 已有相同流水，说明之前已成功操作，直接返回成功
				if req.Logger != nil {
					LogInfo(req.Logger, LogOperation(req.Type), "idempotent: duplicate transaction detected", map[string]interface{}{
						"user_id":     req.UserID,
						"biz_order_id": req.BizOrderID,
					})
				}
				return nil
			}
			if req.Logger != nil {
				LogError(req.Logger, LogOperation(req.Type), "create transaction record failed", err, map[string]interface{}{
					"user_id":     req.UserID,
					"wallet_id":   wallet.ID,
					"biz_order_id": req.BizOrderID,
				})
			}
			return status.Error(codes.Internal, "failed to create transaction record")
		}

		// 7. 执行事务后回调（如果提供）
		if req.AfterTransaction != nil {
			if err := req.AfterTransaction(tx); err != nil {
				if req.Logger != nil {
					LogError(req.Logger, LogOperation(req.Type), "after-transaction callback failed", err, map[string]interface{}{
						"user_id": req.UserID,
					})
				}
				return err
			}
		}

		// 8. 记录成功日志
		if req.Logger != nil {
			LogSuccess(req.Logger, LogOperation(req.Type), map[string]interface{}{
				"user_id":        req.UserID,
				"wallet_id":      wallet.ID,
				"amount":         req.Amount,
				"before_balance": before,
				"after_balance":  after,
				"biz_order_id":   req.BizOrderID,
				"transaction_id": tr.ID,
			})
		}

		return nil
	}); err != nil {
		// 如果是我们在事务中返回的 gRPC status error，直接透传
		if s, ok := status.FromError(err); ok {
			return nil, s.Err()
		}
		// 其他数据库错误
		if req.Logger != nil {
			LogError(req.Logger, LogOperation(req.Type), "transaction failed", err, map[string]interface{}{
				"user_id":     req.UserID,
				"amount":      req.Amount,
				"biz_order_id": req.BizOrderID,
			})
		}
		return nil, status.Error(codes.Internal, "transaction failed, please try again later")
	}

	return &WalletUpdateResult{
		Wallet: &wallet,
	}, nil
}

// ToWalletInfo 转换为 protobuf 的 WalletInfo
func (r *WalletUpdateResult) ToWalletInfo() *user.WalletInfo {
	if r.Wallet == nil {
		return nil
	}
	return &user.WalletInfo{
		UserId:        r.Wallet.UserID,
		Balance:       r.Wallet.Balance,
		FrozenBalance: r.Wallet.FrozenBalance,
	}
}
