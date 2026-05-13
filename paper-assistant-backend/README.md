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

- `LLM_API_KEY`：模型服务 API Key（必填）

说明：

- `LLM_BASE_URL` 已在代码中固定为 `https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions`
- `LLM_MODEL` 已在代码中固定为 `deepseek-v4-flash`

## MySQL 环境变量

- `MYSQL_DSN`：MySQL 连接串  
  默认值：`root:root@tcp(127.0.0.1:3306)/paper_assistant?charset=utf8mb4&parseTime=true&loc=Local`

说明：

- 服务启动时会自动连接 MySQL 并执行建表迁移（users/papers/parse_jobs/paper_parsed_texts）。
- 未配置 `MYSQL_DSN` 时使用默认值，请确保本地 MySQL 和库名匹配。

## Eino 依赖安装

首次拉起前，请在 `paper-assistant-backend` 目录执行：

```bash
go get github.com/cloudwego/eino@latest
go get github.com/cloudwego/eino-ext/components/model/openai@latest
go get github.com/ledongthuc/pdf@latest
go mod tidy
```
