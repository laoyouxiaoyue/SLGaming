package main

import (
	"flag"
	"fmt"

	"SLGaming/back/services/code/code"
	"SLGaming/back/services/code/internal/ioc"
	"SLGaming/back/services/code/internal/server"
	"SLGaming/back/services/code/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	configFile    = flag.String("f", "etc/code.yaml", "the config file")
	templatesFile = flag.String("t", "etc/code-templates.yaml", "the templates file")
)

func main() {
	flag.Parse()

	c := ioc.LoadConfig(*configFile, *templatesFile)
	ctx := svc.NewServiceContext(c)

	registrar, err := ioc.RegisterConsul(c.Consul, c.ListenOn)
	if err != nil {
		logx.Errorf("consul register failed: %v", err)
	} else if registrar != nil {
		defer registrar.Deregister()
	}

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		code.RegisterCodeServer(grpcServer, server.NewCodeServer(ctx))
		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	logx.Info(c.Template)
	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
