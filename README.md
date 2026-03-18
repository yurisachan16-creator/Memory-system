# Memory System

> **Language / 语言：** English | [中文](README.zh-CN.md)

A full-stack AI memory management system built with **Go + Gin + MySQL + Redis** on the backend and **React + TypeScript + Ant Design** on the frontend. Supports memory CRUD, full-text search with multi-factor scoring, category-based summaries, Redis caching, and a bilingual (EN/ZH) UI.

---

## Quick Start

The fastest way to run the entire stack (backend + frontend + MySQL + Redis) is Docker Compose:

```bash
git clone https://github.com/yurisachan16-creator/Memory-system.git
cd Memory-system
docker compose up --build
```

### Access Points

| Service | URL | Description |
|---------|-----|-------------|
| **Frontend UI** | http://localhost:4173 | React dashboard — main entry point |
| **Backend API** | http://localhost:8080/api/v1 | REST API base URL |
| **Swagger Docs** | http://localhost:8080/swagger/index.html | Interactive API documentation |

> **Tip:** Open the Frontend UI first. Use the top bar to set an active `user_id`, then navigate between **Memories**, **Search**, and **Summary** pages. Click the `中文` button (top-right) to switch to Chinese.

---

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Backend language | Go 1.22 |
| Web framework | Gin |
| Database | MySQL 8.4 |
| Cache | Redis 7 |
| Frontend | React 18 + TypeScript + Vite |
| UI library | Ant Design 5 |
| Containerization | Docker + Docker Compose |
| API docs | Swagger / swaggo |

---

## Project Structure

```text
Memory-system/
├── backend/
│   ├── cmd/server/          # Entrypoint (main.go)
│   ├── config/              # config.yaml
│   ├── docs/                # Swagger generated files
│   ├── http/                # Example .http request files
│   ├── internal/
│   │   ├── config/          # Config loading (viper)
│   │   ├── handler/         # Gin route handlers
│   │   ├── middleware/       # CORS, RequestID, RateLimit, Recovery
│   │   ├── model/           # Structs and enums
│   │   ├── repository/      # MySQL + Redis data access
│   │   ├── response/        # Unified response wrapper
│   │   ├── server/          # Gin engine setup and route registration
│   │   └── service/         # Business logic
│   ├── migrations/          # SQL migration files (golang-migrate)
│   ├── Dockerfile
│   ├── go.mod
│   └── go.sum
├── frontend/
│   ├── docs/                # Frontend design documents (i18n, etc.)
│   ├── src/
│   │   ├── api/             # Axios client + API wrappers
│   │   ├── components/      # AppShell, MemoryFormModal, etc.
│   │   ├── context/         # UserContext, LanguageContext
│   │   ├── i18n/            # Translation dictionaries (en-US, zh-CN)
│   │   ├── pages/           # MemoryPage, SearchPage, SummaryPage
│   │   └── types/           # TypeScript interfaces
│   ├── Dockerfile
│   ├── package.json
│   └── vitest.config.ts
├── docker-compose.yml
├── README.md                # This file (English)
└── README.zh-CN.md          # Chinese version
```

---

## API Endpoints

All endpoints use the `/api/v1` prefix.

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/memories` | Create a memory (with deduplication) |
| `GET` | `/memories` | List memories (filter, sort, paginate) |
| `PUT` | `/memories/:id` | Update a memory (owner only) |
| `DELETE` | `/memories/:id` | Soft-delete a memory (owner only) |
| `GET` | `/memories/search` | Full-text search, returns top 3–5 results |
| `GET` | `/memories/summary` | Aggregate summary by category |

Full interactive documentation is available at **http://localhost:8080/swagger/index.html** when the service is running.

---

## Data Model

The `memories` table schema (see `backend/migrations/`):

| Field | Type | Description |
|-------|------|-------------|
| `id` | BIGINT | Auto-increment primary key |
| `user_id` | VARCHAR(64) | User identifier |
| `content` | TEXT | Memory body |
| `category` | ENUM | `preference` / `identity` / `goal` / `context` |
| `source` | ENUM | `chat` / `manual` / `system` |
| `importance` | TINYINT | 1–5 (higher = more important) |
| `content_hash` | VARCHAR(64) | SHA-256 of normalised content — used for deduplication |
| `is_deleted` | TINYINT | Soft-delete flag |
| `created_at` | DATETIME | Creation timestamp |
| `updated_at` | DATETIME | Last update timestamp |

**Indexes:** `idx_user_category`, `idx_user_importance`, `idx_user_created`, `idx_content_hash`, `ft_content` (FULLTEXT)

---

## Core Design

### Deduplication Strategy

Before inserting, content is normalised (whitespace collapsed, trimmed) and hashed with SHA-256 to produce `content_hash`. If a record with the same `user_id + content_hash` already exists:

- No new row is inserted.
- The existing record is **merged**: `importance` is updated to the higher value, `updated_at` is refreshed.
- The update endpoint rejects changes that would produce a hash collision with another existing memory.

**Why merge instead of reject?** Merging preserves information (higher importance) while keeping the database clean — a silent reject would confuse callers, and a hard error on duplicate data is unnecessarily strict for a memory system.

### Delete Strategy

All deletes are **soft deletes** (`is_deleted = 1`). List, search, summary, and deduplication queries filter to `is_deleted = 0`.

**Why soft delete?** Preserves audit history and allows recovery from accidental deletes without external backup infrastructure.

### Search Strategy

Search runs a two-layer retrieval pipeline:

1. **Primary recall** — `MATCH(content) AGAINST(query IN BOOLEAN MODE)` via MySQL FULLTEXT index.
2. **Fallback recall** — `LIKE %query%` to catch short terms and partial matches that FULLTEXT misses.

Results are de-duplicated and ranked by a multi-factor score:

```
final_score = relevance_score × 0.5 + importance_score × 0.3 + recency_score × 0.2
```

- `relevance_score` — keyword hit ratio; full-sentence matches score higher.
- `importance_score` — `importance / 5`.
- `recency_score` — linear decay over a 30-day window.

Top 3–5 results by `final_score DESC` are returned.

### Redis Caching

| Key Pattern | Purpose | TTL |
|-------------|---------|-----|
| `memories:list:{user_id}:{hash(params)}` | List query cache | 5 min |
| `memories:search:{user_id}:{query_hash}` | Search result cache | 5 min |
| `memories:summary:{user_id}` | Summary cache | 10 min |
| `memories:dedup:{user_id}:{content_hash}` | Concurrent write dedup lock (SETNX) | 10 s |

Cache entries for a user are invalidated atomically on any write (create / update / delete). When a cached result is served, the API sets `"cached": true` in the response.

---

## Frontend Features

The React frontend connects to the backend API via Nginx reverse proxy (Docker) or Vite dev-server proxy (local).

| Feature | Description |
|---------|-------------|
| **Global user switcher** | Top-bar input — sets `user_id` for all pages; recent users remembered |
| **Memory Management** | Table view with category filter, sort, pagination, create/edit/delete |
| **Search** | Full-text query with relevance, recency, and final-score display; matched terms highlighted |
| **Summary Dashboard** | Four category buckets — Preferences, Goals, Background, Recent |
| **Bilingual UI** | Click `中文` / `English` in the header to toggle; preference persisted to `localStorage` |

---

## Development Guide

### Run backend locally

```bash
# Ensure MySQL and Redis are running locally, then:
cd backend
go run ./cmd/server
```

Configuration sources (in priority order):
1. Environment variables: `MYSQL_HOST`, `MYSQL_PORT`, `MYSQL_USER`, `MYSQL_PASSWORD`, `MYSQL_DATABASE`, `REDIS_HOST`, `REDIS_PORT`, `SERVER_PORT`
2. `backend/config/config.yaml`

### Run frontend locally (dev mode)

```bash
cd frontend
npm install
npm run dev
# → http://localhost:5173 (Vite dev server with API proxy to :8080)
```

---

## Testing

### Backend

```bash
cd backend
go test ./...
```

Coverage areas:
- **Repository layer** — `sqlmock` covering list, search, and soft-delete ownership checks.
- **Service layer** — deduplication merge, search/summary/list cache invalidation.
- **Handler layer** — `httptest` covering all 6 API endpoints end-to-end.

Example request file: `backend/http/memories.http`

### Frontend

```bash
cd frontend
npm install
npm run test:run      # single run
npm test              # watch mode
npm run test:coverage # coverage report
```

Test suites:
- `i18n.test.ts` — `translate()` function, key completeness, param interpolation.
- `LanguageContext.test.tsx` — provider defaults, toggle, localStorage persistence, param interpolation.
- `AppShell.test.tsx` — toggle button label, full UI language switch, nav items.

---

## Middleware

| Middleware | Description |
|-----------|-------------|
| CORS | Allows cross-origin requests from the frontend dev server |
| RequestID | Generates a UUID per request; returned in `X-Request-Id` response header |
| RateLimit | 30 req/s per IP via token bucket; returns 429 on excess |
| Recovery | Catches panics, returns 500 without crashing the process |

---

## Example curl Requests

```bash
# Create a memory
curl -X POST http://localhost:8080/api/v1/memories \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "demo-user",
    "content": "User prefers pour-over coffee in the morning",
    "category": "preference",
    "source": "manual",
    "importance": 4
  }'

# List memories (page 1, filtered by category)
curl "http://localhost:8080/api/v1/memories?user_id=demo-user&category=preference&page=1&page_size=10"

# Search
curl "http://localhost:8080/api/v1/memories/search?user_id=demo-user&query=coffee&limit=5"

# Summary
curl "http://localhost:8080/api/v1/memories/summary?user_id=demo-user"
```

---

## CI / CD

GitHub Actions runs on every push and pull request to `main`:

1. **Go** — `go test ./...` + `go build ./...`
2. **Frontend** — `npm ci` + `npm run build`
3. **Docker** — `docker compose config` validation

See `.github/workflows/ci.yml` for the full pipeline definition.
