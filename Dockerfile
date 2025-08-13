FROM golang:1.24.6-alpine3.22 AS builder

WORKDIR /src

RUN apk add --no-cache build-base sqlite-dev

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ENV CGO_ENABLED=1 GOOS=linux GOARCH=amd64
RUN go build -ldflags="-s -w" -o /out/main ./cmd/real-time-chat

FROM alpine:3.22

RUN apk add --no-cache ca-certificates sqlite-libs tzdata

WORKDIR /app

COPY --from=builder /out/main /app/main

COPY internal/templates /app/internal/templates
COPY internal/database/migrations /app/internal/database/migrations

ENV PORT=8080 \
    DB_NAME=/app/olha_mensagem.db \
    DB_MIGRATIONS_PATH=/app/internal/database/migrations

EXPOSE 8080
CMD ["/app/main"]