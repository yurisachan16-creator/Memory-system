# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go backend implementation of a simplified AI Memory System, based on the specification in `Go 后端工程师测试题：简易记忆系统.pdf`.

## Technology Stack

- **Language:** Go
- **Framework:** Gin (or similar Go web framework)
- **Database:** MySQL or PostgreSQL
- **Cache:** Redis
- **Containerization:** Docker + Docker Compose

## API Endpoints

所有接口统一使用 `/api/v1` 前缀。

| Method | Path | Description |
|--------|------|-------------|
| POST | /api/v1/memories | Create a memory |
| GET | /api/v1/memories | List memories (filter by user_id, category; sort by importance/time; paginated) |
| PUT | /api/v1/memories/:id | Update a memory (owner only) |
| DELETE | /api/v1/memories/:id | Delete a memory (owner only) |
| GET | /api/v1/memories/search | Search memories (user_id + query → top 3-5 relevant results) |
| GET | /api/v1/memories/summary | Summarize memories by category + importance + recency |

## Data Model

Memory fields: `id`, `user_id`, `content`, `category`, `source`, `importance`, `created_at`, `updated_at`

- `category`: `preference` | `identity` | `goal` | `context`
- `source`: `chat` | `manual` | `system`
- `importance`: integer 1–5

## Frontend Convention

- **user_id 交互方式**：全局顶栏切换。使用 React Context 维护当前 user_id，所有页面共享，不需要每个页面单独输入。顶栏提供输入框 + 最近用户下拉快捷切换。
- **前端技术栈**：React + TypeScript + Vite + Ant Design
- **API 代理**：开发模式下 Vite proxy 转发 `/api` 到后端 `localhost:8080`

## Key Design Decisions to Document

When implementing, the README must explain:
1. **Deduplication strategy** — reject / merge / overwrite duplicate content, and why
2. **Delete strategy** — hard delete vs soft delete, and why
3. **Search strategy** — keyword matching, inverted index, scoring, or full-text search (no vector DB)
4. **Redis usage** — caching layers, search result caching, deduplication assistance

## Common Commands (to be set up)

```bash
# Run the service
go run ./cmd/...

# Run with Docker Compose
docker-compose up --build

# Run database migrations
# (depends on migration tool chosen, e.g. golang-migrate)

# Run tests
go test ./...

# Run a single test
go test ./path/to/package -run TestName
```

## Project Structure (recommended)

```
.
├── cmd/server/         # Entrypoint
├── internal/
│   ├── handler/        # Gin route handlers
│   ├── service/        # Business logic
│   ├── repository/     # DB + Redis access
│   └── model/          # Structs and types
├── migrations/         # SQL migration files
├── docker-compose.yml
├── Dockerfile
└── README.md
```
