# paper-assistant-backend

基于 `Go + Gin` 的论文阅读助手后端（MVP）。

## 当前已实现

- 统一响应结构：`code/message/data/trace_id`
- `trace_id` 中间件（支持透传 `X-Trace-Id`）
- 简化鉴权中间件（演示 token：`Bearer uid-<id>`）
- 认证接口：
  - `POST /api/v1/auth/register`
  - `POST /api/v1/auth/login`
  - `GET /api/v1/auth/me`
- 论文基础接口：
  - `POST /api/v1/papers/upload`
  - `GET /api/v1/papers`
  - `GET /api/v1/papers/{id}`
  - `GET /api/v1/papers/{id}/parse-jobs/latest`
- AI 接口占位：
  - `POST /api/v1/papers/{id}/qa`
  - `POST /api/v1/papers/{id}/summary`
  - `POST /api/v1/papers/{id}/term-explain`
  - `POST /api/v1/papers/compare`
- MySQL 持久化：
  - `users`、`papers`、`parse_jobs` 启动时自动建表
  - 用户、论文、解析任务状态重启后不丢失
- 文件访问：
  - 上传文件落盘到 `uploads/`
  - 通过 `/api/v1/uploads/*` 提供预览访问

## 目录

```text
paper-assistant-backend/
  cmd/server/main.go
  internal/api/handler
  internal/api/middleware
  internal/api/router
  internal/model
  internal/service
  internal/pkg/config
  internal/pkg/errors
  internal/pkg/response
```

## 下一步

- 接入 Redis 任务队列与解析 worker
- 接入 Eino AgentService 与 RAG 检索链路
- 将演示 token 升级为标准 JWT

## 运行要求

- 本地可用 MySQL 实例
- 如未设置 `MYSQL_DSN`，默认使用：

```text
paper_assistant:paper_assistant@tcp(127.0.0.1:3306)/paper_assistant?charset=utf8mb4&parseTime=True&loc=Local
```

## LLM 环境变量

- `LLM_PROVIDER`：`aliyun` / `volcengine` / `openai-compatible`
- `LLM_API_KEY`：模型服务 API Key
- `LLM_BASE_URL`：OpenAI 兼容基础地址（不要写到 `/chat/completions`）
- `LLM_MODEL`：模型名（默认 `qwen-plus`）

说明：

- 当前默认平台为阿里云兼容模式
- 默认基础地址为 `https://dashscope.aliyuncs.com/compatible-mode/v1`
- 如果只设置 `LLM_API_KEY`，后端会按阿里云默认配置调用

## Eino 依赖安装

首次拉起前，请在 `paper-assistant-backend` 目录执行：

```bash
go get github.com/cloudwego/eino@latest
go get github.com/cloudwego/eino-ext/components/model/openai@latest
go mod tidy
```
