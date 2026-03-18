# 项目任务拆分：简易记忆系统

## 项目架构

```
Memory-system/
├── backend/          # Go 后端服务
│   ├── cmd/server/   # 入口
│   ├── internal/
│   │   ├── config/   # 配置管理
│   │   ├── handler/  # Gin 路由处理器
│   │   ├── service/  # 业务逻辑层
│   │   ├── repository/ # 数据访问层 (MySQL + Redis)
│   │   ├── model/    # 数据模型
│   │   └── middleware/ # 中间件（日志、鉴权、错误处理）
│   ├── migrations/   # 数据库迁移文件
│   ├── Dockerfile
│   ├── go.mod
│   └── go.sum
├── frontend/         # 前端界面
│   ├── src/
│   ├── package.json
│   └── Dockerfile
├── docker-compose.yml
└── README.md
```

---

## 后端任务（Backend - Go）

### ✅ B1: 项目脚手架与基础设施
**负责内容：**
- 初始化 Go module (`go mod init`)
- 建立项目目录结构 (cmd/internal/migrations)
- 配置管理 (config.yaml / 环境变量读取)
- Gin 框架初始化与路由注册骨架
- 数据库连接 (MySQL/PostgreSQL) + Redis 连接
- 统一响应格式 (code/message/data)
- 中间件骨架：日志中间件、错误恢复中间件
- `.gitignore` 完善

**交付物：** 可启动的空服务，能连接 DB 和 Redis，路由返回 200

---

### ✅ B2: 数据库设计与迁移
**负责内容：**
- 设计 `memories` 表结构：
  ```sql
  CREATE TABLE memories (
    id          BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id     VARCHAR(64) NOT NULL,
    content     TEXT NOT NULL,
    category    ENUM('preference','identity','goal','context') NOT NULL,
    source      ENUM('chat','manual','system') NOT NULL,
    importance  TINYINT NOT NULL DEFAULT 3 CHECK (importance BETWEEN 1 AND 5),
    content_hash VARCHAR(64) NOT NULL COMMENT '用于去重',
    is_deleted  TINYINT NOT NULL DEFAULT 0 COMMENT '软删除标记',
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_user_category (user_id, category),
    INDEX idx_user_importance (user_id, importance DESC),
    INDEX idx_user_created (user_id, created_at DESC),
    INDEX idx_content_hash (user_id, content_hash),
    FULLTEXT INDEX ft_content (content)
  );
  ```
- 编写迁移文件 (golang-migrate 格式)
- 编写 Model 层结构体与枚举常量

**交付物：** 迁移文件、Model 定义

**依赖：** B1

---

### ✅ B3: 记忆 CRUD 接口
**负责内容：**
- `POST /api/v1/memories` — 新增记忆
  - 参数校验 (content 非空, category/source 枚举, importance 1-5)
  - 去重策略：基于 content_hash (SHA256)，同 user_id 下重复内容执行合并（更新 importance 取较高值，更新 updated_at）
- `GET /api/v1/memories` — 查询记忆列表
  - 查询参数: user_id(必填), category(可选), sort_by(importance/created_at), order(asc/desc), page, page_size
  - 默认按 created_at DESC
  - 分页响应包含 total, page, page_size
- `PUT /api/v1/memories/:id` — 更新记忆
  - 只能修改 content, importance, category
  - 校验 user_id 归属权
- `DELETE /api/v1/memories/:id` — 删除记忆
  - 软删除 (is_deleted = 1)
  - 校验 user_id 归属权

**交付物：** handler + service + repository 三层完整实现

**依赖：** B1, B2

---

### ✅ B4: 记忆检索接口
**负责内容：**
- `GET /api/v1/memories/search?user_id=xxx&query=xxx`
- 检索策略（多层打分）：
  1. MySQL FULLTEXT 全文索引召回候选集
  2. LIKE 模糊匹配补充
  3. 评分公式：`relevance_score * 0.5 + importance * 0.3 + recency_score * 0.2`
  4. 返回 Top 3-5 结果
- Redis 缓存检索结果 (key: `search:{user_id}:{query_hash}`, TTL: 5min)

**交付物：** 搜索接口完整实现，含评分逻辑

**依赖：** B1, B2, B3

---

### ✅ B5: 记忆摘要接口
**负责内容：**
- `GET /api/v1/memories/summary?user_id=xxx`
- 服务端聚合逻辑：
  - `preferences`: category=preference 中 importance >= 3 的记忆
  - `goals`: category=goal 中最新的记忆
  - `background`: category=identity 中 importance 最高的记忆
  - `recent`: 最近 7 天新增的记忆（不限 category）
- Redis 缓存摘要结果 (key: `summary:{user_id}`, TTL: 10min)
- 当用户记忆发生写入/更新/删除时，清除对应缓存

**交付物：** 摘要接口完整实现

**依赖：** B1, B2, B3

---

### ✅ B6: Redis 缓存层设计
**负责内容：**
- 记忆列表缓存: `memories:list:{user_id}:{hash(query_params)}` TTL 5min
- 搜索结果缓存: `memories:search:{user_id}:{query_hash}` TTL 5min
- 摘要缓存: `memories:summary:{user_id}` TTL 10min
- 写入去重锁: `memories:dedup:{user_id}:{content_hash}` TTL 10s (防并发重复写入)
- 缓存失效：写入/更新/删除操作后清除该用户相关缓存
- 封装 Redis 操作到 repository/cache 层

**交付物：** 缓存层实现，集成到各接口

**依赖：** B1, B3, B4, B5

---

### ✅ B7: Docker 化
**负责内容：**
- 后端 `Dockerfile` (多阶段构建)
- `docker-compose.yml` 编排：Go 服务 + MySQL + Redis
- 数据库初始化脚本自动执行
- 环境变量配置

**交付物：** `docker-compose up` 一键启动

**依赖：** B1, B2

---

### ✅ B8: 单元测试与接口测试
**负责内容：**
- Repository 层单元测试 (使用 sqlmock)
- Service 层单元测试
- Handler 层接口测试 (httptest)
- 提供示例请求 (curl 命令或 .http 文件)

**交付物：** 测试文件，`go test ./...` 可通过

**依赖：** B3, B4, B5

---

### ✅ B9: README 文档
**负责内容：**
- 项目整体设计说明
- 表结构设计与说明
- Redis 使用方式 (key 设计、失效策略)
- 检索策略说明 (召回 + 排序)
- 去重/合并策略说明
- 启动与测试说明
- 示例请求

**交付物：** 完整 README.md

**依赖：** 全部后端任务

---

## 前端任务（Frontend）

### ✅ F1: 前端项目初始化
**负责内容：**
- 使用 React + TypeScript + Vite 初始化项目
- 安装依赖：Ant Design (UI 库)、axios (HTTP)、react-router
- 配置 API 代理 (proxy 到后端)
- 建立目录结构: pages / components / api / types
- 前端 Dockerfile

**交付物：** 可运行的前端空项目

---

### ✅ F2: 记忆管理页面
**负责内容：**
- 记忆列表页：表格展示，支持分页、按 category 筛选、按 importance/时间排序
- 新增记忆：表单弹窗 (content, category, source, importance)
- 编辑记忆：表单弹窗，修改 content/importance/category
- 删除记忆：确认弹窗后删除
- 参数校验与错误提示

**交付物：** 记忆 CRUD 完整页面

**依赖：** F1, B3

---

### ✅ F3: 记忆检索页面
**负责内容：**
- 搜索框：输入 query 进行检索
- 结果列表：展示 3-5 条最相关记忆，高亮匹配关键词
- 支持 user_id 选择/切换

**交付物：** 搜索页面

**依赖：** F1, B4

---

### ✅ F4: 记忆摘要页面
**负责内容：**
- Dashboard 样式展示用户记忆摘要
- 分区展示：用户偏好、当前目标、重要背景、最近记忆
- 卡片式布局

**交付物：** 摘要 Dashboard 页面

**依赖：** F1, B5

---

## 任务依赖关系与执行顺序

```
阶段 1（可并行）：B1 + F1
    ↓
阶段 2（可并行）：B2 + B7
    ↓
阶段 3（可并行）：B3 + B4 + B5
    ↓
阶段 4（可并行）：B6 + B8 + F2 + F3 + F4
    ↓
阶段 5：B9（README 文档）
```

## 任务分配建议

| Agent | 任务 | 说明 |
|-------|------|------|
| Agent 1 (后端基础) | B1 → B2 → B7 | 项目脚手架 + 数据库 + Docker |
| Agent 2 (后端核心) | B3 → B4 → B5 | 三个核心 API 模块 |
| Agent 3 (后端进阶) | B6 → B8 → B9 | Redis 缓存 + 测试 + 文档 |
| Agent 4 (前端) | F1 → F2 → F3 → F4 | 全部前端工作 |
