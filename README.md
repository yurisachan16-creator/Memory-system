# Memory System

基于 Go + Gin + MySQL + Redis 的简易记忆系统，实现了记忆 CRUD、搜索、摘要、缓存和测试骨架。

## 功能概览

- `POST /api/v1/memories`：新增记忆，按 `user_id + content_hash` 去重，重复内容执行合并。
- `GET /api/v1/memories`：分页查询，支持按 `category` 筛选、按 `created_at` 或 `importance` 排序。
- `PUT /api/v1/memories/:id`：更新内容、分类和重要度，并校验归属。
- `DELETE /api/v1/memories/:id`：软删除。
- `GET /api/v1/memories/search`：全文检索 + LIKE 兜底召回，返回 Top 3-5。
- `GET /api/v1/memories/summary`：聚合偏好、目标、背景和最近记忆。

## 项目结构

```text
Memory-system/
├── backend/
│   ├── cmd/server
│   ├── config
│   ├── http
│   ├── internal
│   │   ├── config
│   │   ├── handler
│   │   ├── middleware
│   │   ├── model
│   │   ├── repository
│   │   ├── response
│   │   ├── server
│   │   └── service
│   └── migrations
├── docker-compose.yml
└── README.md
```

## 表结构设计

迁移文件位于 `backend/migrations/000001_create_memories_table.up.sql`，核心字段如下：

| 字段 | 说明 |
| --- | --- |
| `id` | 自增主键 |
| `user_id` | 用户标识 |
| `content` | 记忆正文 |
| `category` | `preference` / `identity` / `goal` / `context` |
| `source` | `chat` / `manual` / `system` |
| `importance` | 1-5，数值越高越重要 |
| `content_hash` | 归一化后内容的 SHA256，用于去重 |
| `is_deleted` | 软删除标记 |
| `created_at` / `updated_at` | 创建和更新时间 |

索引设计：

- `idx_user_category`：列表筛选。
- `idx_user_importance`：按重要度排序。
- `idx_user_created`：按时间排序。
- `idx_content_hash`：重复内容查重。
- `ft_content`：MySQL FULLTEXT 搜索召回。

## 核心设计

### 去重与合并策略

- 写入前先做 `NormalizeContent`，压缩空白并去掉首尾空格。
- 再对归一化后的内容做小写 SHA256，形成 `content_hash`。
- 在同一 `user_id` 下若找到相同 `content_hash`：
  - 不新增记录。
  - 合并为已有记录。
  - `importance` 取更高值。
  - `updated_at` 刷新为当前时间。
- 更新接口如果把内容改成了另一条已有记忆的 hash，会返回冲突，避免两条记录在更新阶段合并成脏状态。

### 删除策略

- 删除采用软删除：`is_deleted = 1`。
- 列表、搜索、摘要、去重查找都只处理 `is_deleted = 0` 的记录。
- 这样既保留审计能力，也避免误删后无法恢复。

### 检索策略

搜索流程分两层：

1. MySQL `MATCH(content) AGAINST(...)` 作为主召回。
2. `LIKE %query%` 作为兜底补召回。

排序采用多因素打分：

```text
final_score = relevance_score * 0.5 + importance_score * 0.3 + recency_score * 0.2
```

- `relevance_score`：关键词命中比例，整句命中会获得更高分。
- `importance_score`：`importance / 5`。
- `recency_score`：按 30 天窗口线性衰减。

最终按 `final_score DESC` 排序，取 Top 3-5 条结果。

### Redis 使用方式

缓存层统一封装在 `backend/internal/repository`：

| Key | 说明 | TTL |
| --- | --- | --- |
| `memories:list:{user_id}:{hash(query_params)}` | 记忆列表缓存 | 5 分钟 |
| `memories:search:{user_id}:{query_hash}` | 搜索结果缓存 | 5 分钟 |
| `memories:summary:{user_id}` | 摘要缓存 | 10 分钟 |
| `memories:dedup:{user_id}:{content_hash}` | 写入去重锁 | 10 秒 |

失效策略：

- `CreateMemory` / `UpdateMemory` / `DeleteMemory` 成功后，统一清理该用户的列表缓存、搜索缓存和摘要缓存。
- 搜索和摘要命中缓存时，接口返回 `cached=true`。
- 去重锁通过 Redis `SETNX` 实现，避免并发请求把同一条内容重复写入。

## 启动方式

### 1. 使用 Docker Compose

```bash
docker-compose up --build
```

默认端口：

- 后端：`http://localhost:8080`
- MySQL：容器内 `3306`
- Redis：容器内 `6379`

### 2. 本地运行后端

先确保本机 MySQL 和 Redis 已启动，然后进入 `backend/`：

```bash
go run ./cmd/server
```

配置来源：

- `backend/config/config.yaml`
- 环境变量覆盖，例如 `MYSQL_HOST`、`MYSQL_PORT`、`REDIS_HOST`、`SERVER_PORT`

## 测试

```bash
go test ./...
```

当前测试覆盖：

- repository：`sqlmock` 覆盖 MySQL 列表、搜索和软删除归属校验。
- service：去重合并、搜索缓存、摘要缓存、列表缓存失效。
- handler：`httptest` 覆盖 CRUD、搜索、摘要接口流程。

示例请求文件：

- `backend/http/memories.http`

## 示例请求

```bash
curl -X POST http://localhost:8080/api/v1/memories \
  -H "Content-Type: application/json" \
  -d '{
    "user_id":"demo-user",
    "content":"User prefers pour-over coffee in the morning",
    "category":"preference",
    "source":"manual",
    "importance":4
  }'
```

```bash
curl "http://localhost:8080/api/v1/memories?user_id=demo-user&page=1&page_size=10"
```

```bash
curl "http://localhost:8080/api/v1/memories/search?user_id=demo-user&query=coffee&limit=5"
```

```bash
curl "http://localhost:8080/api/v1/memories/summary?user_id=demo-user"
```
