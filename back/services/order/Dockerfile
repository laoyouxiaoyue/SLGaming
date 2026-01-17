# 构建阶段
FROM golang:1.25.5-alpine AS builder

WORKDIR /build

# 安装必要的构建工具
RUN apk add --no-cache git make

# 复制 go mod 文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY services/ ./services/
COPY pkg/ ./pkg/
COPY rpc/ ./rpc/

# 解析所有导入并下载缺失的依赖
RUN go mod tidy

# 构建应用
WORKDIR /build/services/order
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o order \
    order.go

# 运行阶段
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
ENV TZ=Asia/Shanghai

WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /build/services/order/order .
# 复制配置文件
COPY --from=builder /build/services/order/etc ./etc

# 暴露端口
EXPOSE 8087

# 运行应用
CMD ["./order", "-f", "etc/order.yaml"]
