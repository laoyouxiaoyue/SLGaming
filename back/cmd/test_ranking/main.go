package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"SLGaming/back/services/user/userclient"

	"github.com/zeromicro/go-zero/zrpc"
)

var (
	endpoint = flag.String("endpoint", "127.0.0.1:8086", "user rpc endpoint")
	action   = flag.String("action", "rating", "action: rating|orders")
	page     = flag.Int("page", 1, "page number")
	pageSize = flag.Int("pageSize", 10, "page size")
)

func main() {
	flag.Parse()

	client := userclient.NewUser(zrpc.MustNewClient(zrpc.RpcClientConf{
		Endpoints: []string{*endpoint},
		NonBlock:  true,
	}))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	switch *action {
	case "rating":
		testRatingRanking(ctx, client)
	case "orders":
		testOrdersRanking(ctx, client)
	default:
		fmt.Printf("unknown action: %s\n", *action)
		fmt.Println("usage: -action=rating|orders")
		os.Exit(1)
	}
}

// testRatingRanking 测试评分排行榜
func testRatingRanking(ctx context.Context, cli userclient.User) {
	fmt.Printf("\n=== 测试评分排行榜 (page=%d, pageSize=%d) ===\n", *page, *pageSize)

	resp, err := cli.GetCompanionRatingRanking(ctx, &userclient.GetCompanionRatingRankingRequest{
		Page:     int32(*page),
		PageSize: int32(*pageSize),
	})
	if err != nil {
		fmt.Printf("GetCompanionRatingRanking failed: %v\n", err)
		os.Exit(1)
	}

	printRankingResult("评分榜", resp.Rankings, resp.Total, resp.Page, resp.PageSize)
}

// testOrdersRanking 测试接单数排行榜
func testOrdersRanking(ctx context.Context, cli userclient.User) {
	fmt.Printf("\n=== 测试接单数排行榜 (page=%d, pageSize=%d) ===\n", *page, *pageSize)

	resp, err := cli.GetCompanionOrdersRanking(ctx, &userclient.GetCompanionOrdersRankingRequest{
		Page:     int32(*page),
		PageSize: int32(*pageSize),
	})
	if err != nil {
		fmt.Printf("GetCompanionOrdersRanking failed: %v\n", err)
		os.Exit(1)
	}

	printRankingResult("接单榜", resp.Rankings, resp.Total, resp.Page, resp.PageSize)
}

// printRankingResult 打印排行榜结果
func printRankingResult(title string, rankings []*userclient.CompanionRankingItem, total, page, pageSize int32) {
	fmt.Printf("\n%s - 共%d人，第%d页，每页%d条\n", title, total, page, pageSize)
	fmt.Println("----------------------------------------")

	if len(rankings) == 0 {
		fmt.Println("暂无数据")
		return
	}

	fmt.Printf("%-6s %-15s %-10s %-8s %-6s\n", "排名", "昵称", "评分", "接单数", "认证")
	fmt.Println("----------------------------------------")

	for _, item := range rankings {
		verified := "否"
		if item.IsVerified {
			verified = "是"
		}
		fmt.Printf("%-6d %-15s %-10.2f %-8d %-6s\n",
			item.Rank,
			truncateString(item.Nickname, 15),
			item.Rating,
			item.TotalOrders,
			verified,
		)
	}

	fmt.Println("----------------------------------------")
	fmt.Printf("共查询到 %d 条数据\n\n", len(rankings))
}

// truncateString 截断字符串
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
