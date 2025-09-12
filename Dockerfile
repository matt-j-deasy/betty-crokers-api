# syntax=docker/dockerfile:1.7

########################
# 1) Builder
########################
FROM golang:1.24-alpine AS builder
WORKDIR /app

ENV CGO_ENABLED=0 GO111MODULE=on

# Cache deps
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

# Copy source
COPY . .

# Build the ROOT package (you have main.go at repo root)
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -trimpath -ldflags="-s -w" -o /out/server .

########################
# 2) Runtime (alpine)
########################
FROM alpine:3.20
# Optional: tzdata/ca-certs for outbound TLS and correct timezones
RUN apk add --no-cache ca-certificates tzdata wget
WORKDIR /app
USER nobody

# Environment your code actually reads
ENV LOCAL_PORT=8080 \
    RUN_MODE=local \
    DB_HOST=db \
    DB_PORT=5432 \
    DB_USER=postgres \
    DB_PASSWORD=postgres \
    DB_NAME=app \
    FRONT_END_URL=http://localhost:5173

COPY --from=builder /out/server /app/server

EXPOSE 8080
ENTRYPOINT ["/app/server"]
