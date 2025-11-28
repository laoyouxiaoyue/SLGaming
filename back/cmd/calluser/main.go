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
	endpoint = flag.String("endpoint", "127.0.0.1:8086", "user rpc endpoint")
	action   = flag.String("action", "login", "action: register|login|get")
	phone    = flag.String("phone", "13800000009", "phone number")
	password = flag.String("password", "123456", "password")
	nickname = flag.String("nickname", "测试用户", "nickname for register")
	userID   = flag.Uint64("id", 0, "user id for get action")
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
	case "register":
		register(ctx, client)
	case "login":
		login(ctx, client)
	case "get":
		getUser(ctx, client)
	default:
		panic(fmt.Sprintf("unknown action %s", *action))
	}
}

func register(ctx context.Context, cli userclient.User) {
	resp, err := cli.Register(ctx, &userclient.RegisterRequest{
		Phone:    *phone,
		Password: *password,
		Nickname: *nickname,
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("Register success: id=%d uid=%d\n", resp.GetId(), resp.GetUid())
}

func login(ctx context.Context, cli userclient.User) {
	resp, err := cli.Login(ctx, &userclient.LoginRequest{
		Phone:    *phone,
		Password: *password,
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("Login success: id=%d uid=%d\n", resp.GetId(), resp.GetUid())
}

func getUser(ctx context.Context, cli userclient.User) {
	resp, err := cli.GetUser(ctx, &userclient.GetUserRequest{
		Id: *userID,
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("GetUser success: %+v\n", resp.GetUser())
}
