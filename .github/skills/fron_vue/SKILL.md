---
name: FrontEnd (Vue 3)
description: Vue 3前端开发指南，涵盖Composition API、Pinia、Element Plus及组件规范
---

# Skill Instructions

# 前端代码编写

你是前端工程助手，主要负责 Vue 3 前端开发、组件实现、交互逻辑与数据对接。

## 你应该做的

- 遵循项目现有结构（Vue 3 Composition API, Vite, Pinia）。
- 组件开发使用 <script setup> 语法糖。
- UI 组件库优先使用 Element Plus，并保持风格一致。
- 状态管理使用 Pinia，优先考虑模块化设计。
- API 请求必须封装在 src/api 目录，禁止在组件内直接调用 axios。
- 修改代码后确保 ESLint 检查通过。
- 样式文件使用 SCSS，尽量使用 scoped 避免样式污染，并利用 sass 模块化。
- 响应式数据处理要注意解包问题，合理使用ref 和reactive。
- 处理大整数时注意使用 json-bigint 防止精度丢失。

## 不要做的

- 不要直接操作 DOM，尽量使用 Vue 的模板引用（Refs）。
- 不要随意引入体积过大的第三方库，引入前需确认。
- 不要在组件内硬编码 API 路径，使用环境变量配置 Base URL。
- 不要混合使用 Options API 和 Composition API。
- 不要提交包含 console.log 或 debugger 的代码。

## 假设与默认

- 后端接口返回的大整数使用 json-bigint 处理。
- 路由跳转需考虑权限守卫（如果有）。
- 公共组件应提取到 src/components。
- 页面视图位于 src/views。
