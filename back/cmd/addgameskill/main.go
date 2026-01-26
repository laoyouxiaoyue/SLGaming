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

	client := userclient.NewUser(zrpc.MustNewClient(zrpc.RpcClientConf{
		Endpoints: []string{*endpoint},
		NonBlock:  true,
	}))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 要添加的游戏列表
	games := []struct {
		name        string
		description string
	}{
		{"王者荣耀", "腾讯MOBA手游"},
		{"三角洲行动", "腾讯射击游戏"},
		{"英雄联盟", "Riot Games MOBA游戏"},
		{"无畏契约", "Riot Games战术射击游戏"},
	}

	fmt.Printf("开始添加游戏技能，目标服务器: %s\n", *endpoint)
	fmt.Println("=" + string(make([]byte, 50)) + "=")

	successCount := 0
	failCount := 0

	for _, game := range games {
		fmt.Printf("\n正在添加: %s\n", game.name)

		resp, err := client.CreateGameSkill(ctx, &userclient.CreateGameSkillRequest{
			Name:        game.name,
			Description: game.description,
		})

		if err != nil {
			fmt.Printf("  ❌ 失败: %v\n", err)
			failCount++
		} else {
			fmt.Printf("  ✅ 成功: ID=%d, 名称=%s, 描述=%s\n",
				resp.GetSkill().GetId(),
				resp.GetSkill().GetName(),
				resp.GetSkill().GetDescription())
			successCount++
		}
	}

	fmt.Println("\n" + "=" + string(make([]byte, 50)) + "=")
	fmt.Printf("添加完成: 成功 %d 个, 失败 %d 个\n", successCount, failCount)
}
