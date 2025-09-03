ARG BUILDPLATFORM
ARG TARGETPLATFORM

FROM --platform=$BUILDPLATFORM oven/bun:1.2.20-alpine AS frontend
WORKDIR /olha-mensagem-app/internal/frontend/olha-mensagem-app
COPY internal/frontend/olha-mensagem-app/package.json \
    internal/frontend/olha-mensagem-app/bun.lock \
    internal/frontend/olha-mensagem-app/svelte.config.js \
    internal/frontend/olha-mensagem-app/tsconfig.json \
    internal/frontend/olha-mensagem-app/vite.config.ts \
    ./
RUN --mount=type=cache,target=/root/.cache/bun bun install --frozen-lockfile
COPY internal/frontend/olha-mensagem-app/ .
RUN bun run build

FROM golang:1.25.0-alpine3.22 AS builder
ARG TARGETOS
ARG TARGETARCH
WORKDIR /src
RUN apk add --no-cache build-base sqlite-dev
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download
COPY . .
COPY --from=frontend /olha-mensagem-app/internal/frontend/olha-mensagem-app/build internal/frontend/olha-mensagem-app/build
ENV CGO_ENABLED=1
RUN --mount=type=cache,target=/root/.cache/go-build \
    go build -trimpath -buildvcs=false -ldflags="-s -w" -o /out/main ./cmd/real-time-chat

FROM alpine:3.22
RUN apk add --no-cache ca-certificates sqlite-libs tzdata \
    && addgroup -S app && adduser -S -G app app \
    && mkdir -p /olha-mensagem-app/data && chown app:app /olha-mensagem-app/data
WORKDIR /olha-mensagem-app
COPY --from=builder /out/main /olha-mensagem-app/main
COPY internal/database/migrations /olha-mensagem-app/internal/database/migrations
ENV PORT=8080 \
    DB_NAME=/olha-mensagem-app/data/olha_mensagem.db \
    DB_MIGRATIONS_PATH=/olha-mensagem-app/internal/database/migrations \
    LOG_LEVEL=INFO
EXPOSE 8080
USER app
CMD ["/olha-mensagem-app/main"]