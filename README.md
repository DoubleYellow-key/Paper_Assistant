# 智能论文阅读辅助助手（Paper Assistant）

基于 `Go + Gin + Eino + Vue3` 的前后端分离项目，面向论文阅读场景，提供上传解析、智能问答、术语解释、摘要生成与多论文对比能力（部分能力为占位实现）。

## 项目亮点

- 前后端分离架构，模块边界清晰，便于迭代和协作。
- 后端统一响应规范：`code / message / data / trace_id`。
- 内置 `trace_id` 中间件，支持链路追踪。
- 提供基础认证与论文管理 API，可直接联调。
- AI 能力基于 Eino 接入，支持问答、摘要、术语解释。
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

- 后端：`Go 1.22`、`Gin`
- AI 编排：`Eino`、`eino-ext/openai`
- 前端：`Vue3`、`TypeScript`、`Vite`、`Pinia`、`Element Plus`
- 文档：`OpenAPI 3.0.3`、Markdown
- 规划中的存储：`MySQL`、`Redis`、向量检索（`pgvector/Milvus/ES`）

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

### 计划中

- Repository 层接入 MySQL 持久化
- 解析任务队列与 Worker
- RAG 检索链路与引用追溯增强
- 多论文对比完整实现

## 快速启动

## 1. 启动后端

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

默认地址：`http://localhost:5173`

## 3. 前端代理与后端地址

- 确保前端请求地址指向后端 `http://localhost:8080`。
- 如有跨域问题，可在后端增加 CORS 中间件或在 Vite 代理中转发。

## 接口文档

- OpenAPI 文件：`paper-assistant-backend/doc/openapi.yaml`
- 后端说明：`paper-assistant-backend/README.md`

## AI 相关环境变量

后端支持以下模型配置：

- `LLM_PROVIDER`：`volcengine` / `aliyun` / `openai-compatible`
- `LLM_API_KEY`：模型服务 API Key
- `LLM_BASE_URL`：OpenAI 兼容基础地址（不带 `/chat/completions`）
- `LLM_MODEL`：模型名（默认 `qwen-plus`）

## 文档索引

- 项目总览与流程：`doc/论文阅读助手-项目总体框架与流程.md`
- Eino 智能体方案：`doc/论文阅读助手-Eino智能体构建方案.md`
- 数据库草案（SQL）：`doc/论文阅读助手-数据库表结构草案(SQL版).md`
- 前端模块设计：`doc/论文阅读助手-前端模块设计.md`
- 后端模块设计：`doc/论文阅读助手-后端模块设计.md`

## 后续建议

- 优先补齐 `compare` 接口与引用结构化返回。
- 增加集成测试覆盖主链路：上传 -> 解析 -> 问答 -> 引用返回。
- 接入真实数据库后，补充迁移脚本与种子数据。

## 许可证

当前仓库未声明开源许可证；如需开源，建议补充 `MIT` 或 `Apache-2.0`。
