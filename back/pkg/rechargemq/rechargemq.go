package rechargemq

// 充值相关事件 Topic/Tags 定义
const (
	rechargeEventTopic       = "recharge_events"
	eventTypeRechargeSuccess = "RECHARGE_SUCCEEDED"
	eventTypeRechargeFailed  = "RECHARGE_FAILED"
	eventTypeRechargeClosed  = "RECHARGE_CLOSED"
)

// RechargeEventTopic 返回充值事件主题
func RechargeEventTopic() string {
	return rechargeEventTopic
}

// EventTypeRechargeSuccess 返回充值成功事件类型
func EventTypeRechargeSuccess() string {
	return eventTypeRechargeSuccess
}

// EventTypeRechargeFailed 返回充值失败事件类型
func EventTypeRechargeFailed() string {
	return eventTypeRechargeFailed
}

// EventTypeRechargeClosed 返回充值关闭事件类型
func EventTypeRechargeClosed() string {
	return eventTypeRechargeClosed
}

// RechargeEventPayload 充值事件负载
// 用于通知用户服务执行入账与订单状态更新
// PaidAt 使用 Unix 时间戳（秒）
type RechargeEventPayload struct {
	OrderNo string `json:"order_no"`
	UserID  uint64 `json:"user_id"`
	Amount  int64  `json:"amount"`
	Status  int    `json:"status"`
	PayType string `json:"pay_type,omitempty"`
	TradeNo string `json:"trade_no,omitempty"`
	PaidAt  int64  `json:"paid_at,omitempty"`
	Remark  string `json:"remark,omitempty"`
}
