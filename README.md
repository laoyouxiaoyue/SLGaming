<div align="center">
	<img src="./assets/images/SLGaming-logo.png" alt="SLGaming" width="220" />
	<h1>SLGaming</h1>
	<p>游戏陪玩平台（go-zero 微服务 + Vue 前端）</p>
	<p>
		<a href="https://github.com/laoyouxiaoyue/SLGaming/graphs/contributors"><img src="https://img.shields.io/github/contributors/laoyouxiaoyue/SLGaming.svg?style=for-the-badge" alt="Contributors" /></a>
		<a href="https://github.com/laoyouxiaoyue/SLGaming/stargazers"><img src="https://img.shields.io/github/stars/laoyouxiaoyue/SLGaming.svg?style=for-the-badge" alt="Stars" /></a>
		<a href="https://github.com/laoyouxiaoyue/SLGaming/network/members"><img src="https://img.shields.io/github/forks/laoyouxiaoyue/SLGaming.svg?style=for-the-badge" alt="Forks" /></a>
		<a href="https://github.com/laoyouxiaoyue/SLGaming/issues"><img src="https://img.shields.io/github/issues/laoyouxiaoyue/SLGaming.svg?style=for-the-badge" alt="Issues" /></a>
	</p>
	<p>
		<a href="https://github.com/laoyouxiaoyue/SLGaming/blob/main/LICENSE"><img src="https://img.shields.io/github/license/laoyouxiaoyue/SLGaming.svg?style=for-the-badge" alt="License" /></a>
		<a href="https://github.com/laoyouxiaoyue/SLGaming/commits/main"><img src="https://img.shields.io/github/last-commit/laoyouxiaoyue/SLGaming.svg?style=for-the-badge" alt="Last Commit" /></a>
		<a href="https://github.com/laoyouxiaoyue/SLGaming"><img src="https://img.shields.io/github/repo-size/laoyouxiaoyue/SLGaming.svg?style=for-the-badge" alt="Repo Size" /></a>
		<a href="https://github.com/laoyouxiaoyue/SLGaming"><img src="https://img.shields.io/github/languages/top/laoyouxiaoyue/SLGaming.svg?style=for-the-badge" alt="Top Language" /></a>
	</p>
	<p>
		<a href="https://nodejs.org/"><img src="https://img.shields.io/node/v/vite.svg?style=for-the-badge" alt="Node Version" /></a>
	</p>
</div>

## 简介


SLGaming 是一套陪玩平台工程，包含用户、陪玩、订单、钱包与充值等核心能力，并通过网关统一对外 API，持续完善中。

演示地址：http://120.26.29.242:9090/

（持续完善中，不代表最终质量）
## ✨ 功能

- 用户注册 / 登录 / 验证码
- 陪玩资料、技能、上下线
- 订单创建 / 支付 / 结算 / 退款
- 钱包与充值（含支付回调）

## 🧩 技术栈

- Go / go-zero / gRPC / REST
- MySQL (GORM) / Redis
- RocketMQ / Consul / Nacos
- JWT / Docker

## 🗂️ 结构

```
assets/               # 资源
└── images/           # 图片
back/                 # 后端（Go）
├── api/              # API 定义（REST）
├── cmd/              # 辅助命令
├── docs/             # 设计/文档
├── pkg/              # 公共包
├── rpc/              # gRPC 定义
├── scripts/          # 脚本与 Swagger
└── services/         # 微服务
	├── gateway/      # API 网关 (REST)
	├── user/         # 用户服务 (gRPC)
	├── order/        # 订单服务 (gRPC)
	├── code/         # 验证码服务 (gRPC)
	└── agent/        # 陪玩服务 (gRPC)
front/                # 前端
└── SLGaming/         # 前端项目（Vite + Vue）
```

## 🚀 快速开始

1. 准备依赖：MySQL、Redis、Consul、Nacos、RocketMQ
2. 配置文件：
	- back/services/gateway/etc/gateway.yaml
	- back/services/user/etc/user.yaml
	- back/services/order/etc/order.yaml
	- back/services/code/etc/code.yaml
3. 启动服务：

```bash
cd back/services/code && go run ./code.go
cd back/services/user && go run ./user.go
cd back/services/order && go run ./order.go
cd back/services/gateway && go run ./gateway.go
```

## 📚 文档

- API：back/api/gateway/*.api
- gRPC：back/rpc/*/*.proto
- Swagger：back/scripts/*/*-api-swagger.json






