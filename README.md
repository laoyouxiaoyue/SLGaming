# SLGaming - 游戏陪玩平台

基于 go-zero 的微服务架构游戏陪玩平台后端系统。

## 技术栈

- **框架**: go-zero v1.9.3
- **语言**: Go 1.23.4
- **数据库**: MySQL (GORM)
- **缓存**: Redis
- **消息队列**: RocketMQ
- **服务发现**: Consul
- **配置中心**: Nacos
- **认证**: JWT
- **容器化**: Docker

## 项目结构

```
back/
├── services/          # 微服务
│   ├── gateway/      # API 网关 (REST)
│   ├── user/         # 用户服务 (gRPC)
│   ├── order/        # 订单服务 (gRPC)
│   └── code/         # 验证码服务 (gRPC)
├── rpc/              # gRPC 定义
├── pkg/              # 公共包
│   ├── ioc/          # 基础设施 (MySQL, Redis, Consul, Nacos, RocketMQ)
│   ├── lock/         # 分布式锁
│   └── snowflake/    # 雪花算法 ID 生成
└── api/              # API 定义
```




