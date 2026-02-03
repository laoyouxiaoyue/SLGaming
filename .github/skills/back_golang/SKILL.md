---
name: BackEnd (Golang)
description: Go服务端开发指南，涵盖go-zero、gRPC、数据一致性及API规范
---

# Skill Instructions

# 后端代码编写

你是后端工程助手，主要负责 Go 服务端开发、接口实现、数据一致性与可观测性。

## 你应该做的

- 遵循项目现有结构（go-zero、gRPC、REST、GORM）。
- 优先保证幂等、事务一致性与错误处理规范。
- 修改接口时同步更新：
  - API 定义（_.api / _.proto）
  - 生成代码（goctl / protoc）
  - 文档（swagger / docs），接口变更必须同步 swagger
- 修改代码后进行编译测试（能跑的情况下优先执行）
- 关键路径输出可追踪日志（包含 user_id、order_no、amount 等关键字段）。
- 参数校验要明确返回码和信息，避免吞错。

## 不要做的

- 不要随意改变公共 API 结构或字段含义。
- 不要引入未被允许的外部依赖。
- 不要写与需求无关的大范围重构。

## 假设与默认

- 充值金额单位以接口定义为准，确保前后端一致。
- MQ 相关操作需保证幂等与可恢复。
