# Olha Mensagem – Real-Time Chat

Lightweight real-time chat application built in Go (Chi + WebSockets) with a SvelteKit frontend embedded into the Go binary. SQLite + migrations via golang-migrate, SQL code generated with sqlc, structured logging, Docker multi-stage build.

</div>

## ✨ Features

- Real-time messaging (broadcast hub) over WebSockets
- User registration & login (bcrypt hashed passwords)
- Simple auth flow (no tokens yet – demo only)
- Embedded SvelteKit build (`embed.FS`) – single self-contained binary
- SQLite with WAL tuning + automatic migrations at startup
- Deterministic SQL layer via `sqlc`
- Structured logging abstraction (pluggable) implemented with `slog`
- Graceful shutdown (context + hub drain)
- Docker & Compose (persistent DB volume)
- Basic unit tests (repository, HTTP handlers, WebSocket user logic)

## 🗂 Project Layout

```
cmd/real-time-chat/main.go        # Application entry point
internal/
  database/                       # DB init + migrations runner
    migrations/                   # SQL migration files
    queries/                      # SQL source for sqlc
  dto/                            # DTO definitions (e.g. UserDTO)
  frontend/                       # Embedded frontend build & source
    embed.go                      # go:embed directive
    olha-mensagem-app/            # SvelteKit app (src + build)
  handlers/                       # HTTP user endpoints
  logger/                         # Logger interface + slog impl
  repository/                     # Generated sqlc code (models, queries)
  server/                         # HTTP server + routes + static serving
  websocket/                      # Hub, client, message & WS handler
.air.toml                         # Air hot-reload config
docker-compose.yml
Dockerfile
sqlc.yml
```

## 🔌 API Overview

Base path: `/api`

### Users
| Method | Path            | Description        | Request Body |
|--------|-----------------|--------------------|--------------|
| POST   | /api/users      | Register new user  | `{ "username": "...", "password": "..." }` |
| POST   | /api/users/login| Login existing user| `{ "username": "...", "password": "..." }` |

Successful responses:
```json
{ "id": 1, "username": "alice" }
```

### WebSocket
Path: `/api/ws/{userId}`  (must be a valid registered user ID)

Outgoing broadcast message shape:
```json
{
  "type": "Chat",
  "userId": 1,
  "username": "alice",
  "content": "hello world",
  "timestamp": "2025-08-28T12:34:56Z",
  "color": "#FF6B6B"
}
```
Client sends plain text frames; server wraps them into structured JSON.

## 🔐 Auth Flow (Demo)
1. Register (stores bcrypt hash)
2. Login returns user DTO (no token/session)
3. Frontend stores user info (e.g. sessionStorage) and opens WS using `id`

> NOTE: Not production-ready. Add JWT / sessions + CSRF + rate limiting for real deployments.

## ⚙️ Environment Variables

| Name                | Default (Docker image)                 | Purpose                        |
|---------------------|----------------------------------------|--------------------------------|
| `PORT`              | `8080`                                 | HTTP listen port               |
| `DB_NAME`           | `/app/data/olha_mensagem.db`           | SQLite database file           |
| `DB_MIGRATIONS_PATH`| `/app/internal/database/migrations`    | Migrations directory           |

Local dev example (optional `.env`):
```
PORT=8080
DB_NAME=./olha_mensagem.db
DB_MIGRATIONS_PATH=./internal/database/migrations
```

## 🛠 Local Development

### Backend
```bash
go mod download
go run ./cmd/real-time-chat
```

Hot reload with Air:
```bash
go install github.com/cosmtrek/air@latest
air
```

### Frontend (SvelteKit) – develop separately
```bash
cd internal/frontend/olha-mensagem-app
npm install   # or bun install
npm run dev   # local dev server
```
Build (assets consumed by Go embed):
```bash
npm run build
```
Rebuild Go binary afterwards to embed updated assets.

## 🧪 Tests
```bash
go test ./...
```
Current coverage targets:
- Repository (SQLite operations)
- HTTP user handlers
- WebSocket user creation logic

## 🗃 SQL Code Generation (sqlc)
Edit queries under `internal/database/queries/`. Then regenerate:
```bash
sqlc generate
```

## 🐳 Docker

Build image:
```bash
docker build -t olha-mensagem-app .
```

Run with Compose (includes persistent volume):
```bash
docker compose up -d
```
Visit: http://localhost:8080

## 🧱 Architecture Summary

- Hub maintains active clients; channels for register/unregister & broadcast
- Each client goroutine reads socket → pushes messages to hub
- Hub fan-outs messages to all clients (broadcast)
- SQLite accessed only via generated repository methods
- Migrations auto-run before server start
- Embedded static files served from SvelteKit build via `embed.FS`
- Graceful shutdown waits and signals hub before exiting

## 🚧 Known Limitations / Future Work

- No JWT / session management
- No message persistence (in-memory only broadcast)
- No rate limiting / flood protection
- No presence indicators (join/leave/system messages)
- Colors may collide (not tracked)
- Limited test coverage (no end-to-end tests yet)

### Potential Enhancements
1. Add auth tokens (JWT or secure session cookies)
2. Persist chat history & pagination endpoints
3. System events (user joined / left) message types
4. Rate limiting & per-connection backpressure
5. Horizontal scaling (external pub/sub – e.g. Redis) for multi-instance broadcast
6. CI pipeline (lint + tests + security scan) if not already configured
7. Add OpenAPI / API docs

## 🧾 Useful Commands
```bash
# Run backend (dev)
air

# Build backend binary
go build -o chat ./cmd/real-time-chat && ./chat

# Run tests
go test -v ./...

# Regenerate SQL code
sqlc generate

# Docker build & run
docker build -t olha-mensagem-app .
docker run -p 8080:8080 olha-mensagem-app
```

## 📦 Dependencies (Core)
- chi (HTTP router)
- gorilla/websocket
- golang-migrate (migrations)
- sqlite3 driver
- sqlc (code generation – dev tool)
- bcrypt (x/crypto)

## 📄 License
MIT – see [LICENSE](LICENSE).

---
Educational / demo project.