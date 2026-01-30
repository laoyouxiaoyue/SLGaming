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
	endpoint = flag.String("endpoint", "120.26.29.242:8086", "user rpc endpoint")
)

func main() {
	flag.Parse()

	client := userclient.NewUser(zrpc.MustNewClient(zrpc.RpcClientConf{
		Endpoints: []string{*endpoint},
		NonBlock:  true,
	}))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	fmt.Printf("查询游戏技能列表，目标服务器: %s\n", *endpoint)
	fmt.Println("=" + string(make([]byte, 50)) + "=")

	resp, err := client.ListGameSkills(ctx, &userclient.ListGameSkillsRequest{})
	if err != nil {
		fmt.Printf("❌ 查询失败: %v\n", err)
		return
	}

	skills := resp.GetSkills()
	fmt.Printf("\n共找到 %d 个游戏技能:\n\n", len(skills))

	for i, skill := range skills {
		fmt.Printf("%d. ID: %d\n", i+1, skill.GetId())
		fmt.Printf("   名称: %s\n", skill.GetName())
		if skill.GetDescription() != "" {
			fmt.Printf("   描述: %s\n", skill.GetDescription())
		}
		fmt.Println()
	}
}
