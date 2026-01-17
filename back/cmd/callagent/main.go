package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"SLGaming/back/services/agent/agentclient"

	"github.com/zeromicro/go-zero/zrpc"
)

var (
	endpoint      = flag.String("endpoint", "127.0.0.1:8080", "agent rpc endpoint")
	action        = flag.String("action", "add", "action: add|recommend")
	userID        = flag.Uint64("user_id", 1001, "user id")
	gender        = flag.String("gender", "male", "gender: male/female/other")
	age           = flag.Int("age", 22, "age")
	gameSkill     = flag.String("game", "王者荣耀", "game skill")
	description   = flag.String("desc", "专业王者荣耀陪玩，擅长打野和ADC，段位王者50星，技术过硬，带你上分！性格开朗，沟通能力强，能快速理解你的需求。", "description")
	pricePerHour  = flag.Int64("price", 50, "price per hour")
	rating        = flag.Float64("rating", 4.8, "rating (0-5)")
	recommendText = flag.String("text", "我想要一个王者荣耀的陪玩", "recommend text")
)

func main() {
	flag.Parse()

	client := agentclient.NewAgent(zrpc.MustNewClient(zrpc.RpcClientConf{
		Endpoints: []string{*endpoint},
		NonBlock:  true,
	}))

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	switch *action {
	case "add":
		addCompanion(ctx, client)
	case "recommend":
		recommendCompanion(ctx, client)
	default:
		panic(fmt.Sprintf("unknown action %s, use: add|recommend", *action))
	}
}

func addCompanion(ctx context.Context, cli agentclient.Agent) {
	fmt.Printf("=== 添加陪玩信息到向量数据库 ===\n")
	fmt.Printf("UserID: %d\n", *userID)
	fmt.Printf("Gender: %s\n", *gender)
	fmt.Printf("Age: %d\n", *age)
	fmt.Printf("Game: %s\n", *gameSkill)
	fmt.Printf("Description: %s\n", *description)
	fmt.Printf("PricePerHour: %d\n", *pricePerHour)
	fmt.Printf("Rating: %.2f\n", *rating)
	fmt.Println()

	resp, err := cli.AddCompanionToVectorDB(ctx, &agentclient.AddCompanionToVectorDBRequest{
		UserId:       *userID,
		Gender:       *gender,
		Age:          int32(*age),
		GameSkill:    *gameSkill,
		Description:  *description,
		PricePerHour: *pricePerHour,
		Rating:       *rating,
	})
	if err != nil {
		fmt.Printf("❌ 错误: %v\n", err)
		return
	}

	if resp.Success {
		fmt.Printf("✅ 成功!\n")
		fmt.Printf("CompanionID: %d\n", resp.CompanionId)
		fmt.Printf("Message: %s\n", resp.Message)
	} else {
		fmt.Printf("❌ 失败: %s\n", resp.Message)
	}
}

func recommendCompanion(ctx context.Context, cli agentclient.Agent) {
	fmt.Printf("=== 推荐陪玩 ===\n")
	fmt.Printf("UserInput: %s\n", *recommendText)
	fmt.Printf("UserID: %d\n", *userID)
	fmt.Println()

	resp, err := cli.RecommendCompanion(ctx, &agentclient.RecommendCompanionRequest{
		UserInput: *recommendText,
		UserId:    *userID,
	})
	if err != nil {
		fmt.Printf("❌ 错误: %v\n", err)
		return
	}

	fmt.Printf("✅ 推荐结果:\n")
	fmt.Printf("Explanation: %s\n", resp.Explanation)
	fmt.Printf("Companions (%d):\n", len(resp.Companions))
	for i, comp := range resp.Companions {
		fmt.Printf("\n[%d] UserID: %d\n", i+1, comp.UserId)
		fmt.Printf("    GameSkill: %s\n", comp.GameSkill)
		fmt.Printf("    Gender: %s\n", comp.Gender)
		fmt.Printf("    Age: %d\n", comp.Age)
		fmt.Printf("    Description: %s\n", comp.Description)
		fmt.Printf("    PricePerHour: %d\n", comp.PricePerHour)
		fmt.Printf("    Rating: %.2f\n", comp.Rating)
		fmt.Printf("    Similarity: %.4f\n", comp.Similarity)
	}
}
