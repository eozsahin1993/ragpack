# syntax=docker/dockerfile:1
FROM ghcr.io/eozsahin1993/ragpack-builder:latest AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    LANCEDB_MODULE="$(go env GOPATH)/pkg/mod/github.com/lancedb/lancedb-go@v0.1.2" && \
    LANCEDB_LIB="$LANCEDB_MODULE/lib/linux_amd64/liblancedb_go.a" && \
    if [ ! -f "$LANCEDB_LIB" ]; then echo "ERROR: liblancedb_go.a not found"; exit 1; fi && \
    CGO_CFLAGS="-I$LANCEDB_MODULE/include" \
    CGO_LDFLAGS="$LANCEDB_LIB -lpthread -ldl -lm" \
    go build -o /ragpack ./cmd/main.go

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

COPY --from=builder /ragpack /ragpack

ENV SQLITE_PATH=/data/ragpack.db
ENV LANCEDB_PATH=/data/lancedb

EXPOSE 9000

CMD ["/ragpack"]
