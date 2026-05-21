# paper-assistant-backend

基于 `Go + Gin` 的论文阅读助手后端（MVP）。

## 当前已实现

- 统一响应结构：`code/message/data/trace_id`
- `trace_id` 中间件（支持透传 `X-Trace-Id`）
- 简化鉴权中间件（演示 token：`Authorization: Bearer uid-<id>`）
- 认证接口：
  - `POST /api/v1/auth/register`
  - `POST /api/v1/auth/login`
  - `GET /api/v1/auth/me`
- 论文基础接口：
  - `POST /api/v1/papers/upload`
  - `GET /api/v1/papers`
  - `GET /api/v1/papers/{id}`
  - `GET /api/v1/papers/{id}/parse-jobs/latest`
- AI 能力：
  - `POST /api/v1/papers/{id}/qa`（RAG：首次提问自动抽取 PDF→切分→Embedding 入库；返回回答 + citations）
  - `POST /api/v1/papers/{id}/summary`
  - `POST /api/v1/papers/{id}/term-explain`
  - `POST /api/v1/papers/{id}/translate`
  - `GET /api/v1/papers/{id}/translations/latest`
  - `POST /api/v1/papers/compare`
- MySQL 持久化：
  - `users`、`papers`、`parse_jobs`、`paper_translations` 启动时自动建表
  - 用户、论文、解析任务状态重启后不丢失
- 文件访问：
  - 上传文件落盘到 `uploads/`
  - 通过 `/api/v1/uploads/*` 提供预览访问
- 本地向量库：
  - 默认落盘到 `vectordb/`（可用 `VECTOR_DB_PATH` 指定）

## 鉴权说明

- 当前为演示版 token：登录成功后返回 `token`，格式为 `uid-<数字>`。
- 需要鉴权的接口均要求请求头：`Authorization: Bearer uid-<数字>`。

## 目录

```text
paper-assistant-backend/
  cmd/server/main.go
  internal/api/handler
  internal/api/middleware
  internal/api/router
  internal/model
  internal/service
  internal/repository
  internal/pkg/config
  internal/pkg/errors
  internal/pkg/response
  doc/
```

## 下一步

- 接入 Redis 任务队列与解析 worker
- 引用结构化（页码/段落定位等）与多论文对比
- 将演示 token 升级为标准 JWT 与权限校验

## 运行要求

- Go（见 go.mod 的 `go` 版本声明）
- 本地可用 MySQL 实例
- 如未设置 `MYSQL_DSN`，默认使用：

```text
paper_assistant:paper_assistant@tcp(127.0.0.1:3306)/paper_assistant?charset=utf8mb4&parseTime=True&loc=Local
```

## 运行

在 `paper-assistant-backend` 目录执行：

```bash
go mod tidy
go run ./cmd/server
```

默认监听 `:8080`，可通过 `HTTP_ADDR` 修改。

## 环境变量

### 服务地址

- `HTTP_ADDR`：HTTP 监听地址（默认 `:8080`）

### MySQL

- `MYSQL_DSN`：MySQL 连接串（不设置时使用上面的默认 DSN）

### LLM（摘要/术语解释/翻译等）

后端通过 Eino 使用 OpenAI-compatible 协议。`LLM_PROVIDER` 主要用于给出默认的 `LLM_BASE_URL`。

- `LLM_PROVIDER`：`aliyun` / `volcengine` / `openai-compatible`（默认 `aliyun`）
- `LLM_API_KEY`：API Key（默认空；为空时 AI 能力将不可用）
- `LLM_BASE_URL`：OpenAI-compatible Base URL（不要写到 `/chat/completions`；末尾 `/` 会被自动规整）
- `LLM_MODEL`：模型名（默认 `qwen-plus`）

### Embedding（RAG）

Embedding 默认复用 LLM 配置（APIKey/BaseURL），可按需覆盖。

- `EMBEDDING_API_KEY`：Embedding API Key（默认同 `LLM_API_KEY`）
- `EMBEDDING_BASE_URL`：Embedding Base URL（默认同 `LLM_BASE_URL`）
- `EMBEDDING_MODEL`：Embedding 模型名（默认 `text-embedding-v3`）

### 向量库

- `VECTOR_DB_PATH`：向量库持久化目录（默认 `vectordb`）

## 接口文档

- OpenAPI：`doc/openapi.yaml`
- 中文说明：`doc/API-接口文档.md`

## 常见问题

- 启动时报 MySQL 连接失败：检查 `MYSQL_DSN`、库/账号是否存在、MySQL 是否允许本机连接。
- AI 接口返回 `missing llm api key`：设置 `LLM_API_KEY`（如 Embedding 单独计费，再设置 `EMBEDDING_API_KEY`）。
- 首次对某论文提问较慢：会进行 PDF 抽取、切分与向量化入库；可删除 `vectordb/` 以重新构建索引。
