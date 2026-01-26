package main

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"time"

	"SLGaming/back/services/user/userclient"

	"github.com/zeromicro/go-zero/zrpc"
)

var (
	endpoint = flag.String("endpoint", "120.26.29.194:8086", "user rpc endpoint")
	count    = flag.Int("count", 10, "要注册的陪玩数量")
)

// 游戏技能列表
var gameSkills = []string{
	"王者荣耀",
	"三角洲行动",
	"英雄联盟",
	"无畏契约",
}

// 陪玩昵称前缀
var nicknamePrefixes = []string{
	"游戏高手", "陪玩小", "专业陪", "游戏达人", "陪玩师",
	"游戏陪", "专业陪玩", "游戏大神", "陪玩小", "游戏专家",
}

func main() {
	flag.Parse()

	client := userclient.NewUser(zrpc.MustNewClient(zrpc.RpcClientConf{
		Endpoints: []string{*endpoint},
		NonBlock:  true,
	}))

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fmt.Printf("开始批量注册陪玩，目标服务器: %s\n", *endpoint)
	fmt.Printf("计划注册数量: %d\n", *count)
	fmt.Println("=" + string(make([]byte, 50)) + "=")

	successCount := 0
	failCount := 0
	registeredUsers := make([]*CompanionInfo, 0)

	// 生成随机种子
	rand.Seed(time.Now().UnixNano())

	// 使用时间戳作为起始偏移，避免重复
	baseOffset := int(time.Now().Unix() % 9000) // 0-8999

	for i := 0; i < *count; i++ {
		// 生成手机号（1380000xxxx，从2000开始避免与已有数据冲突）
		phone := fmt.Sprintf("1380000%04d", 2000+baseOffset+i)
		nickname := fmt.Sprintf("%s%d", nicknamePrefixes[i%len(nicknamePrefixes)], i+1)
		password := "123456"

		fmt.Printf("\n[%d/%d] 正在注册: %s (%s)\n", i+1, *count, nickname, phone)

		// 步骤1: 注册用户（role=2 表示陪玩）
		registerResp, err := client.Register(ctx, &userclient.RegisterRequest{
			Phone:    phone,
			Password: password,
			Nickname: nickname,
			Role:     2, // 陪玩角色
		})

		if err != nil {
			// 如果手机号已存在，尝试下一个
			if err.Error() == "rpc error: code = AlreadyExists desc = phone already registered" {
				fmt.Printf("  ⚠ 手机号已存在，跳过\n")
				failCount++
				continue
			}
			fmt.Printf("  ❌ 注册失败: %v\n", err)
			failCount++
			continue
		}

		userID := registerResp.GetId()
		fmt.Printf("  ✓ 用户注册成功: ID=%d\n", userID)

		// 步骤2: 设置陪玩信息
		// 随机选择游戏技能
		gameSkill := gameSkills[rand.Intn(len(gameSkills))]
		// 随机价格（50-200帅币/小时）
		pricePerHour := int64(50 + rand.Intn(151))
		// 设置为在线状态（1=在线）
		status := int32(1)

		updateResp, err := client.UpdateCompanionProfile(ctx, &userclient.UpdateCompanionProfileRequest{
			UserId:       userID,
			GameSkill:    gameSkill,
			PricePerHour: pricePerHour,
			Status:       status,
		})

		if err != nil {
			fmt.Printf("  ⚠ 用户已注册但设置陪玩信息失败: %v\n", err)
			failCount++
			continue
		}

		profile := updateResp.GetProfile()
		fmt.Printf("  ✅ 陪玩信息设置成功: 游戏=%s, 价格=%d帅币/小时, 状态=%d\n",
			profile.GetGameSkill(), profile.GetPricePerHour(), profile.GetStatus())

		successCount++
		registeredUsers = append(registeredUsers, &CompanionInfo{
			UserID:       userID,
			Phone:        phone,
			Nickname:     nickname,
			GameSkill:    profile.GetGameSkill(),
			PricePerHour: profile.GetPricePerHour(),
			Status:       profile.GetStatus(),
		})

		// 避免请求过快
		time.Sleep(100 * time.Millisecond)
	}

	// 输出汇总信息
	fmt.Println("\n" + "=" + string(make([]byte, 50)) + "=")
	fmt.Printf("批量注册完成: 成功 %d 个, 失败 %d 个\n\n", successCount, failCount)

	if successCount > 0 {
		fmt.Println("已注册的陪玩列表:")
		fmt.Println("-" + string(make([]byte, 50)) + "-")
		for i, info := range registeredUsers {
			statusText := []string{"离线", "在线", "忙碌"}[info.Status]
			fmt.Printf("%d. %s (ID:%d, 手机:%s)\n", i+1, info.Nickname, info.UserID, info.Phone)
			fmt.Printf("   游戏: %s | 价格: %d帅币/小时 | 状态: %s\n",
				info.GameSkill, info.PricePerHour, statusText)
		}
	}
}

type CompanionInfo struct {
	UserID       uint64
	Phone        string
	Nickname     string
	GameSkill    string
	PricePerHour int64
	Status       int32
}
