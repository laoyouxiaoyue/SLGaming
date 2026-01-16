# Agent Proto 生成说明

## 生成命令

```bash
# 在 back/rpc/agent 目录下执行
goctl rpc protoc agent.proto --go_out=../../services/agent --go-grpc_out=../../services/agent --zrpc_out=../../services/agent
```

## 安装 goctl

### Windows (使用 Go 安装)
```bash
go install github.com/zeromicro/go-zero/tools/goctl@latest
```

### 或者使用 protoc 直接生成

如果已安装 protoc 和插件，可以使用：

```bash
# 在 back/rpc/agent 目录下执行
protoc --go_out=../../services/agent --go_opt=paths=source_relative \
       --go-grpc_out=../../services/agent --go-grpc_opt=paths=source_relative \
       agent.proto
```

## 生成的文件

生成后会在 `back/services/agent/agent/` 目录下生成：
- `agent.pb.go` - 消息定义
- `agent_grpc.pb.go` - gRPC 服务定义
