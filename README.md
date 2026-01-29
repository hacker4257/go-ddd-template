# go-ddd-template

一个面向 **长期维护** 的 Go 分布式服务模板，强调清晰分层、显式依赖和工程化实践。

A clean, production-oriented Go project template for building maintainable distributed services with MySQL, Redis, and Kafka.



---

## ✨ 特性 / Features

### 中文
- **清晰分层架构**：`api → app → domain ← infra`
- **双进程模型**
  - `server`：HTTP API
  - `worker`：Outbox 投递 + Kafka Consumer
- **MySQL 5.7**
  - 手写 SQL（可控、可优化）
  - 未来可无痛切换 GORM
- **Redis**
  - Cache-Aside（读缓存）
  - Consumer 幂等（SETNX）
- **Kafka**
  - Producer（携带 request_id）
  - Consumer（重试 + DLQ）
- **Outbox Pattern**
  - 业务数据与事件同事务写入
  - Worker 异步投递，避免一致性问题
- **可观测性**
  - server：`/healthz`、`/readyz`、`/metrics`
  - worker：`/healthz`、`/metrics`（默认 `:9091`）

### English
- **Layered architecture**: `api → app → domain ← infra`
- **Two processes**
  - `server`: HTTP API
  - `worker`: Outbox dispatcher + Kafka consumer
- **MySQL 5.7**
  - Hand-written SQL (predictable & optimizable)
  - Can switch to GORM later
- **Redis**
  - Cache-aside for reads
  - Idempotency (SETNX) for consumers
- **Kafka**
  - Producer with request_id headers
  - Consumer with retry & DLQ
- **Outbox pattern**
  - Business data + events in one DB transaction
  - Async delivery by worker
- **Observability**
  - server: `/healthz`, `/readyz`, `/metrics`
  - worker: `/healthz`, `/metrics` (default `:9091`)

---

## 🧱 架构说明 / Architecture

api → app → domain ← infra

````

- **domain**：业务实体、领域错误、接口（port）
- **app**：用例编排、事务边界
- **infra**：MySQL / Redis / Kafka 实现
- **api**：HTTP handler / middleware / router
- **cmd/server**：组装依赖，启动 HTTP
- **cmd/worker**：Outbox dispatcher + Kafka consumer

---

## 🔄 示例业务流 / Example Flow（UserCreated）

1. `POST /users`
2. `app/user.Service.Create`
3. 同一个 DB 事务内：
   - 插入 `users`
   - 插入 `outbox`
4. `worker` 轮询 outbox → 投递 Kafka `user.events`
5. `worker` 消费 `user.events`
   - Redis 幂等校验
   - 写入 `audit_logs`

---

## 🚀 快速开始 / Quick Start

### 依赖 / Prerequisites
- Go 1.21+
- MySQL 5.7
- Redis
- Kafka

---

## 🗄️ 数据库（无 migrations 说明）

### 中文
当前项目 **没有内置 migrations 工具**。  
你可以：

- 手动执行示例 SQL
- 或自行接入：
  - golang-migrate
  - flyway
  - atlas
  - Liquibase

### English
This template **does not include a built-in migration tool**.

You may:
- Execute the provided SQL manually
- Or integrate your own migration solution:
  - golang-migrate
  - flyway
  - atlas
  - Liquibase

### 示例表结构 / Example Tables
- `users`
- `outbox`
- `audit_logs`

（表结构示例见 `docs/sql/` 或项目说明）

---

## ▶️ 运行 / Run

### 启动 worker
```bash
make worker
````

Worker metrics:

* [http://127.0.0.1:9091/metrics](http://127.0.0.1:9091/metrics)

### 启动 server

```bash
make server
```

Server endpoints:

* `GET /healthz`
* `GET /readyz`
* `GET /metrics`
* `POST /users`
* `GET /users/{id}`

---

## 🧪 示例请求 / Example

```bash
curl -X POST http://127.0.0.1:8080/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice","email":"alice@example.com"}'
```

---

## 📁 项目结构 / Project Layout

```
cmd/
  server/        # HTTP server
  worker/        # outbox + kafka consumer + metrics

internal/
  api/http/      # handlers / middleware / router
  app/           # use cases
  domain/        # entities & ports
  infra/         # mysql / redis / kafka
  pkg/           # config / logger / metrics / health

configs/
  config.yaml
```

---

## 📄 License

MIT