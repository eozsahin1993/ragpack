# syntax=docker/dockerfile:1
FROM ghcr.io/eozsahin1993/ragpack-builder:latest AS builder

ARG TARGETARCH
ARG VERSION=dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build \
    LANCEDB_MODULE="$(go env GOPATH)/pkg/mod/github.com/eozsahin1993/lancedb-go@v0.1.4" && \
    CGO_CFLAGS="-I$LANCEDB_MODULE/include" \
    CGO_LDFLAGS="$LANCEDB_MODULE/lib/linux_${TARGETARCH}/liblancedb_go.a -lpthread -ldl -lm" \
    go build -ldflags "-X ragpack/pkg/api.Version=${VERSION}" -o /ragpack ./cmd

FROM ubuntu:24.04
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

COPY --from=builder /ragpack /ragpack

ENV SQLITE_PATH=/data/ragpack.db
ENV LANCEDB_PATH=/data/lancedb

EXPOSE 9000

CMD ["/ragpack"]
