package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"SLGaming/back/services/user/userclient"

	"github.com/zeromicro/go-zero/zrpc"
)

var (
	endpoint = flag.String("endpoint", "120.26.29.194:8086", "user rpc endpoint")
)

func main() {
	flag.Parse()

	fmt.Println("测试 Gateway 动态服务发现功能")
	fmt.Println("=" + string(make([]byte, 50)) + "=")
	fmt.Println("\n说明：")
	fmt.Println("1. 如果先启动 gateway，后启动微服务，gateway 应该能自动发现服务")
	fmt.Println("2. 如果微服务重启，gateway 应该能自动更新端点")
	fmt.Println("\n测试步骤：")
	fmt.Println("1. 先启动 gateway（微服务未启动）")
	fmt.Println("2. 等待几秒后启动 user-rpc 服务")
	fmt.Println("3. 观察 gateway 日志，应该能看到服务发现和端点更新的日志")
	fmt.Println("\n当前测试：直接调用 user-rpc 服务")
	fmt.Println("-" + string(make([]byte, 50)) + "-")

	client := userclient.NewUser(zrpc.MustNewClient(zrpc.RpcClientConf{
		Endpoints: []string{*endpoint},
		NonBlock:  true,
	}))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 测试调用 ListGameSkills
	fmt.Printf("\n调用 ListGameSkills 接口...\n")
	resp, err := client.ListGameSkills(ctx, &userclient.ListGameSkillsRequest{})
	if err != nil {
		fmt.Printf("❌ 调用失败: %v\n", err)
		fmt.Println("\n可能的原因：")
		fmt.Println("1. user-rpc 服务未启动")
		fmt.Println("2. 服务地址不正确")
		fmt.Println("3. 网络连接问题")
		return
	}

	skills := resp.GetSkills()
	fmt.Printf("✅ 调用成功！找到 %d 个游戏技能:\n\n", len(skills))
	for i, skill := range skills {
		fmt.Printf("%d. %s (ID: %d)\n", i+1, skill.GetName(), skill.GetId())
	}
}
