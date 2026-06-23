# syntax=docker/dockerfile:1
FROM ghcr.io/eozsahin1993/ragpack-builder:latest AS builder

WORKDIR /app

# Download app dependencies alongside lancedb-go already in base image
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build \
    LANCEDB_MODULE="$(go env GOPATH)/pkg/mod/github.com/eozsahin1993/lancedb-go@v0.1.3" && \
    CGO_CFLAGS="-I$LANCEDB_MODULE/include" \
    CGO_LDFLAGS="$LANCEDB_MODULE/lib/linux_amd64/liblancedb_go.a -lpthread -ldl -lm" \
    go build -o /ragpack ./cmd/main.go

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

COPY --from=builder /ragpack /ragpack

ENV SQLITE_PATH=/data/ragpack.db
ENV LANCEDB_PATH=/data/lancedb

EXPOSE 9000

CMD ["/ragpack"]
