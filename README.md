# 智能论文阅读辅助助手（Paper Assistant）

基于 `Go + Gin + Eino + Vue3` 的前后端分离项目，面向论文阅读场景，提供上传解析、智能问答、术语解释、摘要生成与多论文对比能力（部分能力为占位实现）。

## 项目亮点

- 前后端分离架构，模块边界清晰，便于迭代和协作。
- 后端统一响应规范：`code / message / data / trace_id`。
- 内置 `trace_id` 中间件，支持链路追踪。
- 提供基础认证与论文管理 API，可直接联调。
- AI 能力基于 Eino 接入，支持问答、摘要、术语解释。
- `users`、`papers`、`parse_jobs` 已接入 MySQL 持久化，重启后不会丢失。
- 提供 OpenAPI 文档，方便前后端联调与测试。

## 仓库结构

```text
go项目/
  README.md
  .gitignore
  doc/
    gin-eino-项目选题分析.md
    论文阅读助手-项目总体框架与流程.md
    论文阅读助手-前端模块设计.md
    论文阅读助手-后端模块设计.md
    论文阅读助手-Eino智能体构建方案.md
    论文阅读助手-数据库表结构草案(SQL版).md
  paper-assistant-backend/
    cmd/server/main.go
    internal/
    doc/openapi.yaml
  paper-assistant-frontend/
    src/
    package.json
```

## 技术栈

- 后端：`Go`、`Gin`
- AI 编排：`Eino`、`eino-ext/openai`
- 前端：`Vue3`、`TypeScript`、`Vite`、`Pinia`、`Element Plus`
- 文档：`OpenAPI 3.0.3`、Markdown
- 当前存储：`MySQL`（用户、论文、解析任务） + 本地文件系统 `uploads/`
- 规划中的存储：`Redis`、向量检索（`pgvector/Milvus/ES`）

## 当前实现状态（MVP）

### 已实现

- 用户认证：
  - `POST /api/v1/auth/register`
  - `POST /api/v1/auth/login`
  - `GET /api/v1/auth/me`
- 论文基础：
  - `POST /api/v1/papers/upload`
  - `GET /api/v1/papers`
  - `GET /api/v1/papers/{id}`
  - `GET /api/v1/papers/{id}/parse-jobs/latest`
- AI 能力：
  - `POST /api/v1/papers/{id}/qa`
  - `POST /api/v1/papers/{id}/summary`
  - `POST /api/v1/papers/{id}/term-explain`
  - `POST /api/v1/papers/compare`（当前占位）
- 数据持久化：
  - `users` 表：用户账号、邮箱、角色
  - `papers` 表：论文元数据、文件路径、解析状态
  - `parse_jobs` 表：解析任务状态与进度
- 文件存储：
  - PDF 原文件保存在 `paper-assistant-backend/uploads/`
  - 后端通过 `/api/v1/uploads/*` 提供预览访问
- 前端开发访问：
  - Vite 默认监听 `0.0.0.0`
  - 局域网其他设备可直接访问开发服务器

### 计划中

- 解析任务队列与 Worker
- RAG 检索链路与引用追溯增强
- 多论文对比完整实现
- 更完整的登录态（JWT）与权限校验

## 快速启动

## 1. 启动后端

### 前置条件

- 已安装并启动本地 MySQL
- 已创建数据库与账号，或直接使用代码中的默认本地配置
- 如需 AI 能力，至少设置 `LLM_API_KEY`

### 默认本地 MySQL 配置

如果未显式设置 `MYSQL_DSN`，后端默认使用：

```text
paper_assistant:paper_assistant@tcp(127.0.0.1:3306)/paper_assistant?charset=utf8mb4&parseTime=True&loc=Local
```

后端启动时会自动确保以下表存在：

- `users`
- `papers`
- `parse_jobs`

进入 `paper-assistant-backend` 后运行：

```bash
go mod tidy
go run ./cmd/server
```

默认地址：`http://localhost:8080`

## 2. 启动前端

进入 `paper-assistant-frontend` 后运行：

```bash
npm install
npm run dev
```

本机默认地址：`http://localhost:5173`

由于前端默认监听 `0.0.0.0`，同一局域网内的其他设备也可以直接访问：

```text
http://<你的局域网IP>:5173
```

## 3. 前端代理与后端地址

- 确保前端请求地址指向后端 `http://localhost:8080`。
- 当前 Vite 已内置 `/api -> http://localhost:8080` 代理。
- 如需让其他设备直接调用后端，可访问：

```text
http://<你的局域网IP>:8080
```

## 接口文档

- OpenAPI 文件：`paper-assistant-backend/doc/openapi.yaml`
- 后端说明：`paper-assistant-backend/README.md`

## 环境变量

### MySQL

- `MYSQL_DSN`：MySQL 连接串；不设置时使用本地默认值

### AI

后端支持以下模型配置；当前默认平台为阿里云兼容模式，若你只设置 `LLM_API_KEY`，将优先按阿里云配置发起调用。

- `LLM_PROVIDER`：`aliyun` / `volcengine` / `openai-compatible`
- `LLM_API_KEY`：模型服务 API Key
- `LLM_BASE_URL`：OpenAI 兼容基础地址（不带 `/chat/completions`）
- `LLM_MODEL`：模型名（默认 `qwen-plus`）

阿里云默认基础地址：

```text
https://dashscope.aliyuncs.com/compatible-mode/v1
```

## 文档索引

- 项目总览与流程：`doc/论文阅读助手-项目总体框架与流程.md`
- Eino 智能体方案：`doc/论文阅读助手-Eino智能体构建方案.md`
- 数据库草案（SQL）：`doc/论文阅读助手-数据库表结构草案(SQL版).md`
- 前端模块设计：`doc/论文阅读助手-前端模块设计.md`
- 后端模块设计：`doc/论文阅读助手-后端模块设计.md`

## 后续建议

- 优先补齐 `compare` 接口与引用结构化返回。
- 增加集成测试覆盖主链路：上传 -> 解析 -> 问答 -> 引用返回。
- 为 `papers/parse_jobs` 增加迁移脚本与更多自动化测试。

## 许可证

当前仓库未声明开源许可证；如需开源，建议补充 `MIT` 或 `Apache-2.0`。
