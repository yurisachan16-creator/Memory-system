# 记忆系统

> **Language / 语言：** [English](README.md) | 中文

基于 **Go + Gin + MySQL + Redis**（后端）和 **React + TypeScript + Ant Design**（前端）的全栈 AI 记忆管理系统。支持记忆 CRUD、全文检索多因素打分、分类摘要、Redis 多级缓存，以及中英双语切换界面。

---

## 快速启动

使用 Docker Compose 一键启动完整服务栈（后端 + 前端 + MySQL + Redis）：

```bash
git clone https://github.com/yurisachan16-creator/Memory-system.git
cd Memory-system
docker compose up --build
```

### 访问入口

| 服务 | 地址 | 说明 |
|------|------|------|
| **前端界面** | http://localhost:4173 | React 可视化控制台 — 主入口 |
| **后端 API** | http://localhost:8080/api/v1 | REST API 基础路径 |
| **Swagger 文档** | http://localhost:8080/swagger/index.html | 可交互式 API 文档 |

> **使用提示：** 优先打开前端界面。在顶栏输入框中设置 `user_id`，然后在**记忆列表**、**搜索**、**摘要**三个页面间切换。点击右上角的 `中文` / `English` 按钮即可切换语言。

---

## 技术栈

| 层级 | 技术 |
|------|------|
| 后端语言 | Go 1.22 |
| Web 框架 | Gin |
| 数据库 | MySQL 8.4 |
| 缓存 | Redis 7 |
| 前端框架 | React 18 + TypeScript + Vite |
| UI 组件库 | Ant Design 5 |
| 容器化 | Docker + Docker Compose |
| API 文档 | Swagger / swaggo |

---

## 项目结构

```text
Memory-system/
├── backend/
│   ├── cmd/server/          # 服务入口 (main.go)
│   ├── config/              # config.yaml
│   ├── docs/                # Swagger 生成文件
│   ├── http/                # 示例请求 .http 文件
│   ├── internal/
│   │   ├── config/          # 配置加载 (viper)
│   │   ├── handler/         # Gin 路由处理器
│   │   ├── middleware/       # CORS、RequestID、RateLimit、Recovery
│   │   ├── model/           # 结构体与枚举定义
│   │   ├── repository/      # MySQL + Redis 数据访问层
│   │   ├── response/        # 统一响应封装
│   │   ├── server/          # Gin 引擎初始化与路由注册
│   │   └── service/         # 业务逻辑层
│   ├── migrations/          # SQL 迁移文件 (golang-migrate)
│   ├── Dockerfile
│   ├── go.mod
│   └── go.sum
├── frontend/
│   ├── docs/                # 前端设计文档（i18n 等）
│   ├── src/
│   │   ├── api/             # Axios 客户端 + API 封装
│   │   ├── components/      # AppShell、MemoryFormModal 等
│   │   ├── context/         # UserContext、LanguageContext
│   │   ├── i18n/            # 翻译字典（en-US、zh-CN）
│   │   ├── pages/           # 记忆列表、搜索、摘要页
│   │   └── types/           # TypeScript 接口定义
│   ├── Dockerfile
│   ├── package.json
│   └── vitest.config.ts
├── docker-compose.yml
├── README.md                # 英文版
└── README.zh-CN.md          # 本文件（中文版）
```

---

## API 接口

所有接口统一使用 `/api/v1` 前缀。

| 方法 | 路径 | 说明 |
|------|------|------|
| `POST` | `/memories` | 新增记忆（含去重合并） |
| `GET` | `/memories` | 查询记忆列表（筛选、排序、分页） |
| `PUT` | `/memories/:id` | 更新记忆（仅归属用户） |
| `DELETE` | `/memories/:id` | 软删除记忆（仅归属用户） |
| `GET` | `/memories/search` | 全文检索，返回 Top 3–5 条 |
| `GET` | `/memories/summary` | 按分类聚合摘要 |

服务启动后，可在 **http://localhost:8080/swagger/index.html** 查看完整可交互文档。

---

## 数据模型

`memories` 表核心字段（详见 `backend/migrations/`）：

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | BIGINT | 自增主键 |
| `user_id` | VARCHAR(64) | 用户标识 |
| `content` | TEXT | 记忆正文 |
| `category` | ENUM | `preference` / `identity` / `goal` / `context` |
| `source` | ENUM | `chat` / `manual` / `system` |
| `importance` | TINYINT | 1–5，数值越高越重要 |
| `content_hash` | VARCHAR(64) | 归一化内容的 SHA-256，用于去重 |
| `is_deleted` | TINYINT | 软删除标记 |
| `created_at` | DATETIME | 创建时间 |
| `updated_at` | DATETIME | 最后更新时间 |

**索引：** `idx_user_category`、`idx_user_importance`、`idx_user_created`、`idx_content_hash`、`ft_content`（FULLTEXT）

---

## 核心设计

### 去重与合并策略

写入前先对内容做归一化处理（压缩空白、去除首尾空格），再对归一化结果做 SHA-256 得到 `content_hash`。若同一 `user_id` 下已存在相同 `content_hash`：

- **不新增记录。**
- 执行**合并**：`importance` 取较高值，`updated_at` 刷新为当前时间。
- 更新接口若会导致 hash 与另一条记录冲突，直接返回错误，防止数据脏合并。

**为何选择合并而非拒绝？** 合并保留了信息量（较高的 importance），同时保持数据库整洁。静默拒绝会让调用方困惑；而对记忆系统来说重复内容触发硬错误显得过于严苛。

### 删除策略

所有删除均为**软删除**（`is_deleted = 1`）。列表、搜索、摘要和去重查找仅处理 `is_deleted = 0` 的记录。

**为何软删除？** 保留审计能力，支持误删恢复，无需外部备份基础设施。

### 检索策略

搜索采用两层召回流程：

1. **主召回** — `MATCH(content) AGAINST(query IN BOOLEAN MODE)`（MySQL FULLTEXT 全文索引）。
2. **兜底召回** — `LIKE %query%`，覆盖全文索引无法命中的短词和部分匹配。

结果去重后，按多因素公式打分排序：

```
final_score = relevance_score × 0.5 + importance_score × 0.3 + recency_score × 0.2
```

- `relevance_score`：关键词命中比例，整句命中得分更高。
- `importance_score`：`importance / 5`。
- `recency_score`：30 天窗口内线性衰减。

按 `final_score DESC` 取 Top 3–5 条结果返回。

### Redis 缓存

| Key 模式 | 用途 | TTL |
|---------|------|-----|
| `memories:list:{user_id}:{hash(params)}` | 记忆列表缓存 | 5 分钟 |
| `memories:search:{user_id}:{query_hash}` | 搜索结果缓存 | 5 分钟 |
| `memories:summary:{user_id}` | 摘要缓存 | 10 分钟 |
| `memories:dedup:{user_id}:{content_hash}` | 并发写入去重锁（SETNX） | 10 秒 |

任意写入操作（新增 / 更新 / 删除）成功后，统一清除该用户的列表、搜索和摘要缓存。命中缓存时，接口响应中包含 `"cached": true`。

---

## 前端功能

React 前端在 Docker 模式下通过 Nginx 反向代理连接后端，开发模式下通过 Vite devserver proxy 转发。

| 功能 | 说明 |
|------|------|
| **全局用户切换** | 顶栏输入框设置 `user_id`，所有页面共享；支持最近用户快捷切换 |
| **记忆管理** | 表格展示，支持分类筛选、排序、分页，以及新增 / 编辑 / 删除操作 |
| **记忆搜索** | 全文搜索，展示相关度、时效性和综合得分，高亮匹配词 |
| **摘要仪表盘** | 四个分类面板：偏好、目标、背景、最近记忆 |
| **中英双语** | 点击顶栏 `中文` / `English` 按钮即时切换，偏好持久化到 `localStorage` |

---

## 本地开发指南

### 仅运行后端

```bash
# 确保本机 MySQL 和 Redis 已启动，然后：
cd backend
go run ./cmd/server
```

配置优先级（从高到低）：
1. 环境变量：`MYSQL_HOST`、`MYSQL_PORT`、`MYSQL_USER`、`MYSQL_PASSWORD`、`MYSQL_DATABASE`、`REDIS_HOST`、`REDIS_PORT`、`SERVER_PORT`
2. `backend/config/config.yaml`

### 仅运行前端（开发模式）

```bash
cd frontend
npm install
npm run dev
# → http://localhost:5173（Vite devserver，自动代理 API 到 :8080）
```

---

## 测试

### 后端测试

```bash
cd backend
go test ./...
```

覆盖范围：
- **Repository 层** — `sqlmock` 覆盖列表、搜索、软删除归属校验。
- **Service 层** — 去重合并、搜索 / 摘要 / 列表缓存失效逻辑。
- **Handler 层** — `httptest` 覆盖全部 6 个接口端到端流程。

示例请求文件：`backend/http/memories.http`

### 前端测试

```bash
cd frontend
npm install
npm run test:run      # 单次运行
npm test              # 监听模式
npm run test:coverage # 覆盖率报告
```

测试套件：
- `i18n.test.ts` — `translate()` 函数、key 完整性、`{param}` 插值。
- `LanguageContext.test.tsx` — Provider 默认值、切换、localStorage 持久化、参数插值。
- `AppShell.test.tsx` — 切换按钮标签、全局 UI 语言切换、导航项翻译。

---

## 中间件

| 中间件 | 说明 |
|--------|------|
| CORS | 允许前端开发服务器跨域请求 |
| RequestID | 每个请求生成 UUID，写入 `X-Request-Id` 响应头 |
| RateLimit | 令牌桶算法，30 req/s / IP，超限返回 429 |
| Recovery | 捕获 panic，返回 500，不崩溃进程 |

---

## 示例 curl 请求

```bash
# 新增记忆
curl -X POST http://localhost:8080/api/v1/memories \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "demo-user",
    "content": "用户早晨偏好手冲咖啡",
    "category": "preference",
    "source": "manual",
    "importance": 4
  }'

# 查询记忆列表（第 1 页，按分类筛选）
curl "http://localhost:8080/api/v1/memories?user_id=demo-user&category=preference&page=1&page_size=10"

# 搜索
curl "http://localhost:8080/api/v1/memories/search?user_id=demo-user&query=咖啡&limit=5"

# 摘要
curl "http://localhost:8080/api/v1/memories/summary?user_id=demo-user"
```

---

## CI / CD

每次向 `main` 分支推送或提交 Pull Request 时，GitHub Actions 自动执行：

1. **Go** — `go test ./...` + `go build ./...`
2. **前端** — `npm ci` + `npm run build`
3. **Docker** — `docker compose config` 配置校验

详见 `.github/workflows/ci.yml`。
