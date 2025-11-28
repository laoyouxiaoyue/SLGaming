package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"SLGaming/back/services/code/codeclient"

	"github.com/zeromicro/go-zero/zrpc"
)

var (
	endpoint = flag.String("endpoint", "127.0.0.1:8085", "code rpc endpoint")
	phone    = flag.String("phone", "13800000009", "phone number")
	purpose  = flag.String("purpose", "login", "code purpose")
)

func main() {
	flag.Parse()

	client := codeclient.NewCode(zrpc.MustNewClient(zrpc.RpcClientConf{
		Endpoints: []string{*endpoint},
		NonBlock:  true,
	}))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := client.SendCode(ctx, &codeclient.SendCodeRequest{
		Phone:   *phone,
		Purpose: *purpose,
	})
	if err != nil {
		panic(err)
	}

	fmt.Printf("SendCode success: requestId=%s expireAt=%d\n", resp.RequestId, resp.ExpireAt)
}
