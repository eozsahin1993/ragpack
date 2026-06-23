# syntax=docker/dockerfile:1
FROM --platform=linux/amd64 golang:1.26-bookworm AS builder

WORKDIR /app

RUN apt-get update && apt-get install -y curl && rm -rf /var/lib/apt/lists/*

# Download Go deps + LanceDB native lib (BuildKit cache survives --no-cache rebuilds)
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download && \
    LANCEDB_MODULE="/go/pkg/mod/github.com/lancedb/lancedb-go@v0.1.2" && \
    chmod -R u+w "$LANCEDB_MODULE" && \
    cd "$LANCEDB_MODULE" && \
    bash scripts/download-artifacts.sh v0.1.2

# Build binary (Go build cache also persisted)
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    ARCH=$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/') && \
    LANCEDB_LIB="/go/pkg/mod/github.com/lancedb/lancedb-go@v0.1.2/lib/linux_${ARCH}/liblancedb_go.a" && \
    if [ ! -f "$LANCEDB_LIB" ]; then echo "ERROR: liblancedb_go.a not found at $LANCEDB_LIB"; exit 1; fi && \
    echo "Using: $LANCEDB_LIB" && \
    CGO_CFLAGS="-I/go/pkg/mod/github.com/lancedb/lancedb-go@v0.1.2/include" \
    CGO_LDFLAGS="$LANCEDB_LIB -lpthread -ldl -lm" \
    go build -o /ragpack ./cmd/main.go

FROM --platform=linux/amd64 debian:bookworm-slim
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

COPY --from=builder /ragpack /ragpack

ENV SQLITE_PATH=/data/ragpack.db
ENV LANCEDB_PATH=/data/lancedb

EXPOSE 9000

CMD ["/ragpack"]
