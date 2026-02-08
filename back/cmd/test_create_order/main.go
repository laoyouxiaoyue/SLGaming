package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"SLGaming/back/services/order/orderclient"

	"github.com/zeromicro/go-zero/zrpc"
)

func main() {
	// 这里请根据实际情况填写 order rpc 服务地址
	orderRpcAddr := os.Getenv("ORDER_RPC")
	if orderRpcAddr == "" {
		orderRpcAddr = "127.0.0.1:8087" // 默认本地端口
	}

	cli := zrpc.MustNewClient(zrpc.RpcClientConf{
		Endpoints: []string{orderRpcAddr},
		NonBlock:  true,
	})
	orderClient := orderclient.NewOrder(cli)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := orderClient.CreateOrder(ctx, &orderclient.CreateOrderRequest{
		CompanionId:   2017119996664090624,
		GameName:      "英雄联盟",
		DurationHours: 1,
		// BossId 通常由网关鉴权注入，这里本地测试可写死
		BossId: 1996102769315942400,
	})
	if err != nil {
		fmt.Printf("CreateOrder failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("CreateOrder success: %+v\n", resp.Order)

	time.Sleep(5 * time.Second)
	// 取消订单测试
	// 注意：这里只是演示，实际应确保订单ID和操作人ID有效
	cancelResp, err := orderClient.CancelOrder(ctx, &orderclient.CancelOrderRequest{
		OrderId:    resp.Order.Id,     // 使用刚创建的订单ID
		OperatorId: resp.Order.BossId, // 假设老板取消
		Reason:     "测试取消",
	})
	if err != nil {
		fmt.Printf("CancelOrder failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("CancelOrder success: %+v\n", cancelResp.Order)
}
