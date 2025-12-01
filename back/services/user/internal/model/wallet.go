package model

// 用户钱包与钱包流水模型（帅币）
//
// 设计说明：
// 1. UserWallet：每个用户一条记录，存当前帅币余额（以及预留的冻结金额）
// 2. WalletTransaction：每次余额变动一条流水，便于对账和排查问题

// UserWallet 用户钱包（帅币账户）
type UserWallet struct {
	BaseModel

	// 关联的用户 ID（users.id）
	UserID uint64 `gorm:"not null;index;comment:用户ID" json:"user_id,string"`

	// 帅币可用余额（单位：帅币，整型，避免浮点精度问题）
	Balance int64 `gorm:"not null;default:0;comment:帅币可用余额" json:"balance"`

	// 预留字段：冻结帅币（例如提现、仲裁时暂时冻结）
	FrozenBalance int64 `gorm:"not null;default:0;comment:冻结帅币余额" json:"frozen_balance"`
}

func (w *UserWallet) TableName() string {
	return "user_wallets"
}

// WalletTransaction 钱包流水
type WalletTransaction struct {
	BaseModel

	// 关联的用户 ID（冗余字段，便于按用户查询）
	UserID uint64 `gorm:"not null;index;comment:用户ID" json:"user_id,string"`

	// 关联的钱包 ID（user_wallets.id）
	WalletID uint64 `gorm:"not null;index;comment:钱包ID" json:"wallet_id,string"`

	// 变动金额（正数=增加帅币，例如充值；负数=减少帅币，例如消费）
	ChangeAmount int64 `gorm:"not null;comment:变动金额(正数=增加,负数=减少)" json:"change_amount"`

	// 变动前余额
	BeforeBalance int64 `gorm:"not null;comment:变动前余额" json:"before_balance"`

	// 变动后余额
	AfterBalance int64 `gorm:"not null;comment:变动后余额" json:"after_balance"`

	// 变动类型：RECHARGE(充值) / CONSUME(消费) / REFUND(退款) / ADJUST(人工调整) 等
	Type string `gorm:"size:32;index;not null;comment:变动类型" json:"type"`

	// 业务关联单号：例如订单号、充值单号等，用于追踪来源
	BizOrderID string `gorm:"size:64;index;comment:关联业务订单号" json:"biz_order_id"`

	// 备注信息（可选）
	Remark string `gorm:"size:255;comment:备注" json:"remark"`
}

func (t *WalletTransaction) TableName() string {
	return "wallet_transactions"
}


