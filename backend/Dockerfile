FROM golang:1.24-bookworm AS builder

WORKDIR /app

# Download Go dependencies (cached unless go.mod/go.sum change)
COPY go.mod go.sum ./
RUN go mod download

# Download LanceDB native library for Linux (cached unless lancedb-go version changes)
RUN LANCEDB_MODULE="$(go env GOPATH)/pkg/mod/github.com/lancedb/lancedb-go@v0.1.2" && \
    chmod -R u+w "$LANCEDB_MODULE" && \
    cd "$LANCEDB_MODULE" && \
    bash scripts/download-artifacts.sh v0.1.2

# Build binary
COPY . .
RUN LANCEDB_MODULE="$(go env GOPATH)/pkg/mod/github.com/lancedb/lancedb-go@v0.1.2" && \
    ARCH=$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/') && \
    CGO_CFLAGS="-I${LANCEDB_MODULE}/include" \
    CGO_LDFLAGS="${LANCEDB_MODULE}/lib/linux_${ARCH}/liblancedb_go.a -lpthread -ldl -lm" \
    go build -o /ragpack ./cmd/main.go

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

COPY --from=builder /ragpack /ragpack

ENV SQLITE_PATH=/data/ragpack.db
ENV LANCEDB_PATH=/data/lancedb

EXPOSE 9000

CMD ["/ragpack"]
