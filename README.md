<div align="center">
	<img src="./assets/images/SLGaming-logo.png" alt="SLGaming" width="220" />
	<h1>SLGaming</h1>
	<p>游戏陪玩平台 | 微服务架构实战项目</p>
	<p>
		<a href="https://github.com/laoyouxiaoyue/SLGaming/graphs/contributors"><img src="https://img.shields.io/github/contributors/laoyouxiaoyue/SLGaming.svg?style=flat-square" alt="Contributors" /></a>
		<a href="https://github.com/laoyouxiaoyue/SLGaming/stargazers"><img src="https://img.shields.io/github/stars/laoyouxiaoyue/SLGaming.svg?style=flat-square" alt="Stars" /></a>
		<a href="https://github.com/laoyouxiaoyue/SLGaming/network/members"><img src="https://img.shields.io/github/forks/laoyouxiaoyue/SLGaming.svg?style=flat-square" alt="Forks" /></a>
		<a href="https://github.com/laoyouxiaoyue/SLGaming/issues"><img src="https://img.shields.io/github/issues/laoyouxiaoyue/SLGaming.svg?style=flat-square" alt="Issues" /></a>
	</p>
	<p>
		<a href="https://github.com/laoyouxiaoyue/SLGaming/blob/main/LICENSE"><img src="https://img.shields.io/github/license/laoyouxiaoyue/SLGaming.svg?style=flat-square" alt="License" /></a>
		<a href="https://github.com/laoyouxiaoyue/SLGaming/commits/main"><img src="https://img.shields.io/github/last-commit/laoyouxiaoyue/SLGaming.svg?style=flat-square" alt="Last Commit" /></a>
		<img src="https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat-square&logo=go" alt="Go Version" />
		<img src="https://img.shields.io/badge/Vue-3.5+-4FC08D?style=flat-square&logo=vue.js" alt="Vue Version" />
		<img src="https://img.shields.io/badge/Node-20+-339933?style=flat-square&logo=node.js" alt="Node Version" />
	</p>
</div>

## 📖 简介

SLGaming 是一套完整的游戏陪玩平台系统，采用微服务架构设计，包含用户管理、陪玩匹配、订单交易、钱包充值等核心业务能力。项目采用主流云原生技术栈，支持服务注册发现、配置中心、分布式限流、AI智能推荐等企业级特性。

🌐 **在线演示**: http://120.26.29.242:9090/

> ⚠️ 注意：演示环境仅作展示用途，数据会定期清理

---

## ✨ 核心功能

### 用户系统
- 🔐 用户注册 / 登录（支持验证码登录）
- 📝 个人资料管理 / 头像上传
- 🔒 密码修改 / 手机号换绑

### 陪玩系统
- 👤 陪玩资料管理（技能、等级、价格）
- 🎮 游戏技能标签（LOL、王者、吃鸡等）
- 🟢 在线状态切换（接单/休息）
- 🤖 **AI 智能推荐**（基于向量相似度的陪玩匹配）

### 订单系统
- 🛒 订单创建 / 取消 / 接单
- 💰 订单支付（支付宝沙箱）
- ⭐ 订单评价与评分
- 📊 陪玩排行榜（订单量、评分）

### 钱包系统
- 💳 账户余额管理
- 💵 充值功能（支付宝支付）
- 🔔 充值回调通知（RocketMQ异步处理）
- 📈 交易记录查询

### 社交功能
- ❤️ 关注 / 取关
- 👥 粉丝列表 / 关注列表
- 🔍 双向关注查询

---

## 🏗️ 技术架构

### 后端技术栈

| 技术 | 版本 | 用途 |
|------|------|------|
| **Go** | 1.24+ | 后端开发语言 |
| **go-zero** | 1.9.3 | 微服务框架 |
| **gRPC** | 1.78+ | 服务间通信 |
| **GORM** | 1.25+ | ORM 框架 |
| **JWT** | v4 | 身份认证 |
| **RocketMQ** | 5.3.2 | 消息队列 |
| **Milvus** | 2.4+ | 向量数据库 |

### 前端技术栈

| 技术 | 版本 | 用途 |
|------|------|------|
| **Vue.js** | 3.5.26 | 前端框架 |
| **Vite** | 7.3.0 | 构建工具 |
| **Element Plus** | 2.13.1 | UI 组件库 |
| **Pinia** | 3.0.4 | 状态管理 |
| **Vue Router** | 4.6.4 | 路由管理 |
| **Axios** | 1.13.2 | HTTP 客户端 |

### 基础设施

| 服务 | 版本 | 端口 | 用途 |
|------|------|------|------|
| **MySQL** | 8.0.44 | 3306 | 主数据库 |
| **Redis** | 7-alpine | 6379 | 缓存 / Token存储 |
| **Nacos** | v3.1.0 | 8848 | 配置中心 / 服务注册 |
| **Consul** | latest | 8500 | 服务发现 |
| **RocketMQ** | 5.3.2 | 9876 | 消息队列 |

---

## 🗂️ 项目结构

```
SLGaming/
├── 📁 assets/                    # 静态资源
│   └── images/                   # 图片资源
│
├── 📁 back/                      # 后端服务（Go）
│   ├── 📁 api/                   # REST API 定义（go-zero .api 文件）
│   │   └── gateway/              # 网关 API 定义
│   │
│   ├── 📁 cmd/                   # 命令行工具
│   │   ├── addgameskill/         # 添加游戏技能
│   │   ├── batchregistercompanion/ # 批量注册陪玩
│   │   ├── callagent/            # 调用 agent 服务测试
│   │   ├── callcode/             # 调用 code 服务测试
│   │   ├── calluser/             # 调用 user 服务测试
│   │   └── test_create_order/    # 订单创建测试
│   │
│   ├── 📁 pkg/                   # 公共包
│   │   ├── ioc/                  # 依赖注入容器
│   │   ├── lock/                 # 分布式锁
│   │   ├── snowflake/            # 雪花算法 ID 生成
│   │   └── avatarmq/             # 头像处理 MQ
│   │
│   ├── 📁 rpc/                   # gRPC Proto 定义
│   │   ├── agent/                # Agent 服务协议
│   │   ├── code/                 # Code 服务协议
│   │   ├── order/                # Order 服务协议
│   │   └── user/                 # User 服务协议
│   │
│   ├── 📁 scripts/               # 脚本和文档
│   │   ├── agent/                # Agent 服务脚本
│   │   ├── code/                 # Code 服务脚本
│   │   ├── gateway/              # Gateway 脚本
│   │   ├── order/                # Order 服务脚本
│   │   └── user/                 # User 服务脚本
│   │
│   ├── 📁 services/              # 🎯 微服务目录
│   │   ├── 📁 gateway/           # API 网关（REST 入口）
│   │   │   ├── etc/gateway.yaml  # 网关配置
│   │   │   ├── internal/         # 内部实现
│   │   │   │   ├── handler/      # HTTP 处理器
│   │   │   │   ├── logic/        # 业务逻辑
│   │   │   │   ├── middleware/   # 中间件
│   │   │   │   └── jwt/          # JWT 实现
│   │   │   └── Dockerfile        # 容器构建
│   │   │
│   │   ├── 📁 user/              # 用户服务（gRPC）
│   │   │   ├── etc/user.yaml     # 服务配置
│   │   │   ├── internal/
│   │   │   │   ├── logic/        # 32个业务逻辑文件
│   │   │   │   ├── model/        # 数据模型
│   │   │   │   ├── cache/        # 缓存层
│   │   │   │   └── job/          # 定时任务
│   │   │   └── Dockerfile
│   │   │
│   │   ├── 📁 order/             # 订单服务（gRPC）
│   │   │   ├── etc/order.yaml    # 服务配置
│   │   │   ├── internal/
│   │   │   │   ├── logic/        # 订单业务逻辑
│   │   │   │   └── tx/           # 事务管理
│   │   │   └── Dockerfile
│   │   │
│   │   ├── 📁 code/              # 验证码服务（gRPC）
│   │   │   ├── etc/code.yaml     # 服务配置
│   │   │   ├── etc/code-templates.yaml # 短信模板
│   │   │   └── Dockerfile
│   │   │
│   │   └── 📁 agent/             # AI 智能服务（gRPC）
│   │       ├── etc/agent.yaml    # 服务配置
│   │       ├── data/             # 向量数据
│   │       └── internal/
│   │           ├── embedder/     # 向量嵌入
│   │           └── llm/          # 大模型接口
│   │
│   ├── go.mod                    # Go 模块定义
│   └── go.sum                    # 依赖锁定
│
├── 📁 front/                     # 前端项目
│   └── 📁 SLGaming/              # Vue.js 项目
│       ├── src/
│       │   ├── api/              # API 接口封装
│       │   ├── components/       # 公共组件
│       │   ├── router/           # 路由配置
│       │   ├── stores/           # Pinia 状态管理
│       │   ├── utils/            # 工具函数
│       │   └── views/            # 页面视图
│       ├── package.json          # 依赖配置
│       ├── vite.config.js        # Vite 配置
│       ├── nginx.conf            # Nginx 部署配置
│       └── Dockerfile            # 容器构建
│
├── 📁 .github/
│   └── workflows/                # GitHub Actions CI/CD
│       ├── go.yml                # Go 后端 CI
│       └── vue.yml               # Vue 前端 CI
│
├── docker-compose.yaml           # 基础设施编排
├── Jenkinsfile                   # Jenkins CI/CD 流水线
└── README.md                     # 项目说明
```

---

## 🚀 快速开始

### 环境要求

- **Go**: 1.24+
- **Node.js**: 20+ 
- **Docker & Docker Compose**
- **Make** (可选)

### 1. 启动基础设施

```bash
# 使用 Docker Compose 一键启动所有依赖服务
docker-compose up -d

# 查看服务状态
docker-compose ps
```

启动的服务：
- MySQL (3306)
- Redis (6379)
- Nacos (8848)
- Consul (8500)
- RocketMQ Namesrv (9876)
- RocketMQ Dashboard (8082)

### 2. 数据库初始化

```bash
# 连接 MySQL 创建数据库
mysql -h 127.0.0.1 -P 3306 -u root -p

# 创建数据库
CREATE DATABASE IF NOT EXISTS SLGaming CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

### 3. 配置服务

修改各服务的配置文件（将 IP 改为你的实际地址）：

```bash
# 后端配置文件列表
back/services/gateway/etc/gateway.yaml    # 网关（端口8888）
back/services/user/etc/user.yaml          # 用户服务（端口8086）
back/services/order/etc/order.yaml        # 订单服务（端口8087）
back/services/code/etc/code.yaml          # 验证码服务（端口8085）
back/services/agent/etc/agent.yaml        # Agent服务（端口8090）
```

### 4. 启动后端服务

**方式一：直接运行（开发模式）**

```bash
cd back/

# 安装依赖
go mod download

# 启动各个服务（需要多个终端窗口）
cd services/code && go run code.go
cd services/user && go run user.go  
cd services/order && go run order.go
cd services/agent && go run agent.go
cd services/gateway && go run gateway.go
```

**方式二：使用 Docker 运行**

```bash
# 构建镜像
cd back/services/gateway && docker build -t slgaming-gateway .
cd back/services/user && docker build -t slgaming-user .
# ... 其他服务

# 运行容器
docker run -d -p 8888:8888 slgaming-gateway
```

### 5. 启动前端

```bash
cd front/SLGaming/

# 安装依赖
npm install

# 开发模式启动
npm run dev

# 构建生产版本
npm run build
```

### 6. 访问系统

- 🌐 **前端页面**: http://localhost:5173/ (开发) / http://localhost:9090/ (生产)
- 📘 **API 文档**: http://localhost:8888/swagger/
- 🔍 **Nacos 控制台**: http://localhost:8848/nacos/ (nacos/nacos)
- 🔍 **Consul 控制台**: http://localhost:8500/
- 📊 **RocketMQ Dashboard**: http://localhost:8082/

---

## 📡 API 接口概览

### 网关服务端口：8888

| 模块 | 接口路径 | 说明 |
|------|----------|------|
| **用户** | `POST /api/user/register` | 用户注册 |
| | `POST /api/user/login` | 用户登录 |
| | `POST /api/user/login/code` | 验证码登录 |
| | `GET /api/user` | 获取用户信息 |
| | `PUT /api/user` | 更新用户信息 |
| | `POST /api/user/upload/avatar` | 上传头像 |
| | `GET /api/user/wallet` | 获取钱包信息 |
| **陪玩** | `POST /api/user/companion/apply` | 申请成为陪玩 |
| | `GET /api/user/companion` | 获取陪玩资料 |
| | `PUT /api/user/companion` | 更新陪玩资料 |
| | `POST /api/user/companion/status` | 更新在线状态 |
| | `GET /api/user/companion/list` | 陪玩列表 |
| | `GET /api/game/skills` | 游戏技能列表 |
| | `GET /api/agent/recommend` | AI 智能推荐 |
| **订单** | `POST /api/order` | 创建订单 |
| | `GET /api/order/:id` | 获取订单详情 |
| | `GET /api/order` | 订单列表 |
| | `POST /api/order/:id/accept` | 接单 |
| | `POST /api/order/:id/cancel` | 取消订单 |
| | `POST /api/order/:id/complete` | 完成订单 |
| | `POST /api/order/:id/rate` | 评价订单 |
| **充值** | `POST /api/user/recharge` | 创建充值订单 |
| | `GET /api/user/recharge` | 充值记录 |
| | `POST /api/user/recharge/alipay/notify` | 支付宝回调 |
| **关注** | `POST /api/follow/:userId` | 关注用户 |
| | `DELETE /api/follow/:userId` | 取消关注 |
| | `GET /api/follow/following` | 关注列表 |
| | `GET /api/follow/followers` | 粉丝列表 |
| | `GET /api/follow/mutual` | 双向关注 |
| **验证码** | `POST /api/code/send` | 发送验证码 |
| | `POST /api/code/verify` | 验证验证码 |

---

## 📊 项目统计

| 指标 | 数值 |
|------|------|
| **后端代码行数** | ~31,500 行 (Go) |
| **前端代码行数** | ~6,100 行 (Vue/JS) |
| **微服务数量** | 5 个 |
| **Go 文件数** | 244 个 |
| **Vue 组件数** | 26 个 |
| **配置文件** | 10+ 个 |
| **Docker 镜像** | 6 个 |

---

## 🔧 开发工具

### 常用命令

```bash
# 格式化代码
cd back && gofmt -w .

# 运行测试
cd back && go test ./...

# 生成 API 文档（go-zero）
cd back/api/gateway && goctl api go -api gateway.api -dir ../..

# 生成 gRPC 代码
cd back/rpc/user && goctl rpc protoc user.proto --go_out=.. --go-grpc_out=..

# 前端代码检查
cd front/SLGaming && npm run lint
```

---

## 📝 配置说明

### 核心配置项

#### Gateway 配置 (`back/services/gateway/etc/gateway.yaml`)

```yaml
# 服务端口
Port: 8888

# JWT 配置
JWT:
  SecretKey: "your-secret-key"           # 修改生产密钥
  AccessTokenDuration: 600s              # Access Token 有效期
  RefreshTokenDuration: 1209600s         # Refresh Token 有效期

# 限流配置
RateLimit:
  Enabled: true
  GlobalQPS: 2000
  Routes:
    - Path: "/api/code/send"
      PerIPQPS: 5                        # 每秒每IP限制

# 支付宝配置
Alipay:
  AppID: "your-app-id"
  PrivateKey: "your-private-key"
  IsProduction: false                   # 沙箱/生产环境切换
```

#### Agent 服务配置 (`back/services/agent/etc/agent.yaml`)

```yaml
# LLM 配置
LLM:
  APIKey: "your-dashscope-api-key"      # 阿里云 DashScope
  Model: "text-embedding-v4"            # 向量嵌入模型
  ChatModel: "qwen-plus"                # 对话模型

# Milvus 配置
Milvus:
  Address: "127.0.0.1:19530"
  Username: "root"
  Password: "Milvus"
```

---

## 🤝 参与贡献

我们欢迎所有形式的贡献，包括但不限于：

- 🐛 提交 Bug 报告
- 💡 提出新功能建议
- 📝 改进文档
- 🔧 提交代码修复
- 🎨 优化 UI/UX

### 贡献流程

1. Fork 本仓库
2. 创建你的特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交你的修改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开一个 Pull Request

---

## 📄 开源协议

本项目基于 [MIT](LICENSE) 协议开源。

---

## 🙏 致谢

- [go-zero](https://github.com/zeromicro/go-zero) - 优秀的 Go 微服务框架
- [Element Plus](https://element-plus.org/) - 强大的 Vue 3 组件库
- [Milvus](https://milvus.io/) - 开源向量数据库
- [RocketMQ](https://rocketmq.apache.org/) - 分布式消息队列

---

<div align="center">
	<p>Made with ❤️ by SLGaming Team</p>
	<p>⭐ Star 我们，支持项目发展！</p>
</div>
