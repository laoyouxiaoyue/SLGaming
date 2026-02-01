package user

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

const (
	rechargeStatusPending = 0
	rechargeStatusSuccess = 1
	rechargeStatusFailed  = 2
	rechargeStatusClosed  = 3
)

type rechargeOrder struct {
	OrderNo   string `json:"orderNo"`
	UserId    uint64 `json:"userId"`
	Amount    int64  `json:"amount"`
	Status    int    `json:"status"`
	PayType   string `json:"payType"`
	CreatedAt int64  `json:"createdAt"`
	ExpiresAt int64  `json:"expiresAt"`
}

func rechargeOrderKey(orderNo string) string {
	return fmt.Sprintf("recharge:order:%s", orderNo)
}

func saveRechargeOrder(r *redis.Redis, order *rechargeOrder, ttl time.Duration) error {
	if r == nil {
		return errors.New("redis not configured")
	}
	if order == nil {
		return errors.New("order is nil")
	}
	if ttl <= 0 {
		ttl = 30 * time.Minute
	}
	payload, err := json.Marshal(order)
	if err != nil {
		return err
	}
	return r.Setex(rechargeOrderKey(order.OrderNo), string(payload), int(ttl.Seconds()))
}

func loadRechargeOrder(r *redis.Redis, orderNo string) (*rechargeOrder, error) {
	if r == nil {
		return nil, errors.New("redis not configured")
	}
	val, err := r.Get(rechargeOrderKey(orderNo))
	if err != nil || val == "" {
		return nil, errors.New("order not found")
	}
	var order rechargeOrder
	if err := json.Unmarshal([]byte(val), &order); err != nil {
		return nil, err
	}
	return &order, nil
}
