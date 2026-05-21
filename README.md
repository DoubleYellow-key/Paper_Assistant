# 智能论文阅读辅助助手（Paper Assistant）

基于 `Go + Gin + Eino + Vue3` 的前后端分离项目，面向论文阅读场景，提供论文上传、（轻量）解析状态管理、基于论文内容的知识问答（RAG）、术语解释、摘要生成与翻译等能力。

## 项目亮点

- 前后端分离架构，模块边界清晰，便于迭代和协作。
- 后端统一响应规范：`code / message / data / trace_id`。
- 内置 `trace_id` 中间件，支持链路追踪。
- 提供基础认证与论文管理 API，可直接联调。
- AI 能力通过 Eino 接入 OpenAI-compatible 模型，支持摘要、术语解释、翻译等。
- 提供基于 `chromem-go` 的本地持久化向量库：首次对某篇论文提问时自动抽取 PDF 文本、切分、Embedding 入库，后续问答走相似检索（RAG）。
- `users`、`papers`、`parse_jobs`、`paper_translations` 已接入 MySQL 持久化，重启后不会丢失。
- 提供 OpenAPI 文档，方便前后端联调与测试。

## 仓库结构

```text
Paper_Assistant/
  README.md
  paper-assistant-backend/        # Go + Gin 后端
    cmd/server/main.go            # 启动入口
    internal/                     # handler/service/repository 等
    doc/                          # OpenAPI 与接口说明
  paper-assistant-frontend/       # Vue3 + Vite 前端
    src/
```

## 技术栈

- 后端：`Go`、`Gin`
- AI 编排：`Eino`、`eino-ext/openai`
- 前端：`Vue3`、`TypeScript`、`Vite`、`Pinia`、`Element Plus`
- 文档：`OpenAPI 3.0.3`、Markdown
- 当前存储：`MySQL`（用户、论文、解析任务、翻译结果） + 本地文件系统 `uploads/` + 本地向量库目录（默认 `vectordb/`）

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
  - `POST /api/v1/papers/{id}/qa`（RAG：向量检索 + 引用片段）
  - `POST /api/v1/papers/{id}/summary`
  - `POST /api/v1/papers/{id}/term-explain`
  - `POST /api/v1/papers/{id}/translate`
  - `GET /api/v1/papers/{id}/translations/latest`
  - `POST /api/v1/papers/compare`（当前占位）
- 数据持久化：
  - `users` 表：用户账号、邮箱、角色
  - `papers` 表：论文元数据、文件路径、解析状态
  - `parse_jobs` 表：解析任务状态与进度
  - `paper_translations` 表：翻译结果缓存（按论文 + 目标语言唯一）
- 文件存储：
  - PDF 原文件保存在 `paper-assistant-backend/uploads/`
  - 后端通过 `/api/v1/uploads/*` 提供预览访问
- 前端开发访问：
  - Vite 默认监听 `0.0.0.0`
  - 局域网其他设备可直接访问开发服务器

### 计划中

- 解析任务队列与 Worker（当前上传后 parse_job 直接标记为 completed）
- RAG 的引用结构化（页码/段落定位等）
- 多论文对比完整实现
- 标准 JWT 与权限校验

## 快速启动

## 1. 启动后端

### 前置条件

- 已安装并启动本地 MySQL
- 已创建数据库与账号，或直接使用默认本地配置
- 如需 AI 能力（摘要/术语解释/翻译/RAG），至少设置 `LLM_API_KEY`（Embedding 默认复用该 Key）

### 默认本地 MySQL 配置

如果未显式设置 `MYSQL_DSN`，后端默认使用：

```text
paper_assistant:paper_assistant@tcp(127.0.0.1:3306)/paper_assistant?charset=utf8mb4&parseTime=True&loc=Local
```

后端启动时会自动确保以下表存在：

- `users`
- `papers`
- `parse_jobs`
- `paper_translations`

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
- 中文接口说明：`paper-assistant-backend/doc/API-接口文档.md`
- 后端说明：`paper-assistant-backend/README.md`

## 鉴权说明

- 当前为演示版 token：登录成功后返回 `token`，格式为 `uid-<数字>`。
- 调用需要鉴权的接口时，请在请求头携带：`Authorization: Bearer uid-<数字>`。

## 环境变量

### 服务地址

- `HTTP_ADDR`：HTTP 监听地址（默认 `:8080`）

### MySQL

- `MYSQL_DSN`：MySQL 连接串（默认见上文）

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

## 常见问题

- 启动时报 MySQL 连接失败：检查 `MYSQL_DSN`、库/账号是否存在、MySQL 是否允许本机连接。
- AI 接口返回 `missing llm api key`：设置 `LLM_API_KEY`（如 Embedding 单独计费，再设置 `EMBEDDING_API_KEY`）。
- 首次对某论文提问较慢：会进行 PDF 抽取、切分与向量化入库；可删除 `vectordb/` 以重新构建索引。

## 后续建议

- 优先补齐 `compare` 接口与引用结构化返回。
- 增加集成测试覆盖主链路：上传 -> 解析 -> 问答 -> 引用返回。
- 为 `papers/parse_jobs` 增加迁移脚本与更多自动化测试。

## 许可证

当前仓库未声明开源许可证；如需开源，建议补充 `MIT` 或 `Apache-2.0`。
