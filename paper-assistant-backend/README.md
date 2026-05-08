# paper-assistant-backend

基于 `Go + Gin` 的论文阅读助手后端骨架（MVP）。

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

- 接入 MySQL 与 repository 层
- 接入 Redis 任务队列与解析 worker
- 接入 Eino AgentService 与 RAG 检索链路

## LLM 环境变量

- `LLM_PROVIDER`：`volcengine` / `aliyun` / `openai-compatible`
- `LLM_API_KEY`：模型服务 API Key
- `LLM_BASE_URL`：OpenAI 兼容基础地址（不要写到 `/chat/completions`）
- `LLM_MODEL`：模型名（默认 `qwen-plus`）

## Eino 依赖安装

首次拉起前，请在 `paper-assistant-backend` 目录执行：

```bash
go get github.com/cloudwego/eino@latest
go get github.com/cloudwego/eino-ext/components/model/openai@latest
go mod tidy
```
