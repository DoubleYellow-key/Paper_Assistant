# 论文阅读助手后端 API 接口文档（联调版）

本文档基于当前后端代码实现整理，适用于前端搭建与前后端联调。

## 1. 基础信息

- 服务前缀：`/api/v1`
- 默认地址：`http://localhost:8080`
- 数据格式：
  - 普通接口：`application/json`
  - 上传接口：`multipart/form-data`

## 2. 统一响应格式

所有接口均返回如下结构：

```json
{
  "code": 0,
  "message": "ok",
  "data": {},
  "trace_id": "a1b2c3d4e5f6"
}
```

- `code`：业务码（`0` 表示成功）
- `message`：提示信息
- `data`：业务数据（失败时通常为空）
- `trace_id`：链路追踪 ID（便于排查问题）

## 3. 通用请求头

### 3.1 鉴权头（受保护接口必填）

```http
Authorization: Bearer uid-<user_id>
```

说明：

- 当前为演示版 token 方案，登录成功后返回 `token`（如 `uid-1`）。
- 受保护接口要求完整格式：`Bearer uid-1`。

### 3.2 Trace 透传（可选）

```http
X-Trace-Id: your-trace-id
```

若不传，后端自动生成并在响应头/响应体中返回。

## 4. 错误码说明

| code | 含义 |
| --- | --- |
| `0` | 成功 |
| `40001` | 请求参数错误 |
| `40101` | 未授权（缺失或非法 token） |
| `40102` | token 过期（预留） |
| `40301` | 无权限（预留） |
| `40401` | 资源不存在 |
| `40901` | 状态冲突（如邮箱已存在） |
| `42901` | 请求限流（预留） |
| `50011` | 解析失败（预留） |
| `50021` | 大模型调用失败 |
| `50000` | 系统内部错误（预留） |

## 5. 数据对象定义

### 5.1 User

```json
{
  "id": 1,
  "username": "alice",
  "email": "alice@example.com",
  "role": "user",
  "created_at": "2026-05-08T12:00:00Z"
}
```

### 5.2 Paper

```json
{
  "id": 1,
  "user_id": 1,
  "title": "Attention Is All You Need",
  "file_name": "paper.pdf",
  "file_path": "/uploads/paper.pdf",
  "file_size": 123456,
  "parse_status": "pending",
  "created_at": "2026-05-08T12:00:00Z",
  "updated_at": "2026-05-08T12:00:00Z"
}
```

### 5.3 ParseJob

```json
{
  "id": 1,
  "paper_id": 1,
  "status": "queued",
  "progress": 0,
  "retry_count": 0,
  "max_retries": 3,
  "created_at": "2026-05-08T12:00:00Z",
  "updated_at": "2026-05-08T12:00:00Z"
}
```

### 5.4 AskResponse（Eino）

```json
{
  "answer": "这是模型回答",
  "citations": [],
  "confidence": "medium"
}
```

## 6. 接口明细

## 6.1 认证模块

### 6.1.1 注册

- 方法：`POST`
- 路径：`/api/v1/auth/register`
- 鉴权：否
- `Content-Type`：`application/json`

请求体：

```json
{
  "username": "alice",
  "email": "alice@example.com",
  "password": "123456"
}
```

成功响应示例：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "user": {
      "id": 1,
      "username": "alice",
      "email": "alice@example.com",
      "role": "user",
      "created_at": "2026-05-08T12:00:00Z"
    }
  },
  "trace_id": "a1b2c3d4e5f6"
}
```

失败场景：

- 参数非法：`40001`
- 邮箱已存在：`40901`

### 6.1.2 登录

- 方法：`POST`
- 路径：`/api/v1/auth/login`
- 鉴权：否
- `Content-Type`：`application/json`

请求体：

```json
{
  "email": "alice@example.com",
  "password": "123456"
}
```

成功响应示例：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "token": "uid-1",
    "user": {
      "id": 1,
      "username": "alice",
      "email": "alice@example.com",
      "role": "user",
      "created_at": "2026-05-08T12:00:00Z"
    }
  },
  "trace_id": "a1b2c3d4e5f6"
}
```

失败场景：

- 参数非法：`40001`
- 账号/密码错误：`40101`

### 6.1.3 当前用户信息

- 方法：`GET`
- 路径：`/api/v1/auth/me`
- 鉴权：是

成功响应：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "user": {
      "id": 1,
      "username": "alice",
      "email": "alice@example.com",
      "role": "user",
      "created_at": "2026-05-08T12:00:00Z"
    }
  },
  "trace_id": "a1b2c3d4e5f6"
}
```

---

## 6.2 论文模块

### 6.2.1 上传论文

- 方法：`POST`
- 路径：`/api/v1/papers/upload`
- 鉴权：是
- `Content-Type`：`multipart/form-data`

表单字段：

- `file`：文件（必填）
- `title`：标题（可选，为空则使用文件名）

成功响应：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "paper": {
      "id": 1,
      "user_id": 1,
      "title": "paper.pdf",
      "file_name": "paper.pdf",
      "file_path": "/uploads/paper.pdf",
      "file_size": 123456,
      "parse_status": "pending",
      "created_at": "2026-05-08T12:00:00Z",
      "updated_at": "2026-05-08T12:00:00Z"
    },
    "parse_job": {
      "id": 1,
      "paper_id": 1,
      "status": "queued",
      "progress": 0,
      "retry_count": 0,
      "max_retries": 3,
      "created_at": "2026-05-08T12:00:00Z",
      "updated_at": "2026-05-08T12:00:00Z"
    }
  },
  "trace_id": "a1b2c3d4e5f6"
}
```

失败场景：

- 未传文件：`40001`
- 未鉴权：`40101`

### 6.2.2 获取论文列表

- 方法：`GET`
- 路径：`/api/v1/papers`
- 鉴权：是

成功响应：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "items": [
      {
        "id": 1,
        "user_id": 1,
        "title": "paper.pdf",
        "file_name": "paper.pdf",
        "file_path": "/uploads/paper.pdf",
        "file_size": 123456,
        "parse_status": "pending",
        "created_at": "2026-05-08T12:00:00Z",
        "updated_at": "2026-05-08T12:00:00Z"
      }
    ]
  },
  "trace_id": "a1b2c3d4e5f6"
}
```

### 6.2.3 获取论文详情

- 方法：`GET`
- 路径：`/api/v1/papers/{id}`
- 鉴权：是

路径参数：

- `id`：论文 ID（`uint64`）

失败场景：

- `id` 非法：`40001`
- 论文不存在：`40401`

### 6.2.4 获取最新解析任务状态

- 方法：`GET`
- 路径：`/api/v1/papers/{id}/parse-jobs/latest`
- 鉴权：是

路径参数：

- `id`：论文 ID（`uint64`）

成功响应：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "parse_job": {
      "id": 1,
      "paper_id": 1,
      "status": "queued",
      "progress": 0,
      "retry_count": 0,
      "max_retries": 3,
      "created_at": "2026-05-08T12:00:00Z",
      "updated_at": "2026-05-08T12:00:00Z"
    }
  },
  "trace_id": "a1b2c3d4e5f6"
}
```

---

## 6.3 AI 模块（Eino）

### 6.3.1 论文问答

- 方法：`POST`
- 路径：`/api/v1/papers/{id}/qa`
- 鉴权：是
- `Content-Type`：`application/json`

请求体：

```json
{
  "query": "这篇论文的核心贡献是什么？"
}
```

成功响应（`data` 直接为 AskResponse）：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "answer": "模型生成的回答",
    "confidence": "medium"
  },
  "trace_id": "a1b2c3d4e5f6"
}
```

失败场景：

- `agent service unavailable`：`50021`（HTTP 503）
- 缺少 API Key：`50021`（HTTP 503）
- 模型调用失败：`50021`（HTTP 502）

### 6.3.2 论文摘要

- 方法：`POST`
- 路径：`/api/v1/papers/{id}/summary`
- 鉴权：是
- 请求体同 `qa`：`{"query":"..."}`
- 响应同 `qa`

### 6.3.3 术语解释

- 方法：`POST`
- 路径：`/api/v1/papers/{id}/term-explain`
- 鉴权：是
- 请求体同 `qa`：`{"query":"..."}`
- 响应同 `qa`

### 6.3.4 多论文对比（当前占位）

- 方法：`POST`
- 路径：`/api/v1/papers/compare`
- 鉴权：是

当前实现返回：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "message": "TODO: compare endpoint"
  },
  "trace_id": "a1b2c3d4e5f6"
}
```

说明：

- 该接口已预留路由，业务逻辑待实现。

## 7. 前后端联调建议顺序

1. 完成注册/登录，拿到 `token`
2. 在前端请求拦截器统一注入 `Authorization: Bearer <token>`
3. 打通论文上传与论文列表
4. 上传后轮询 `parse-jobs/latest`
5. 打通 `qa/summary/term-explain`
6. 最后接入 `compare`（待后端实现）

## 8. 前端实现注意事项

- 对所有响应统一按 `code === 0` 判断成功，不要只看 HTTP 状态码。
- 建议把 `trace_id` 记录到前端日志，联调时可快速定位后端日志。
- 上传接口必须使用 `multipart/form-data`，字段名必须是 `file`。
- AI 接口若返回 `50021`，优先检查：
  - `LLM_API_KEY` 是否正确
  - `LLM_BASE_URL` 是否为兼容基础地址（不带 `/chat/completions`）
  - 模型名是否可用

## 9. 版本说明

- 文档版本：`v0.1`
- 对应代码阶段：后端骨架 + Eino 基础调用 + 内存数据存储
