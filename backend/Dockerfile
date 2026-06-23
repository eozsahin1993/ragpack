FROM --platform=linux/amd64 golang:1.26-bookworm AS builder

WORKDIR /app

RUN apt-get update && apt-get install -y curl && rm -rf /var/lib/apt/lists/*

# Download Go dependencies (cached unless go.mod/go.sum change)
COPY go.mod go.sum ./
RUN go mod download

# Download LanceDB native library (cached unless lancedb-go version changes)
RUN LANCEDB_MODULE="$(go env GOPATH)/pkg/mod/github.com/lancedb/lancedb-go@v0.1.2" && \
    chmod -R u+w "$LANCEDB_MODULE" && \
    cd "$LANCEDB_MODULE" && \
    bash scripts/download-artifacts.sh v0.1.2 && \
    echo "=== downloaded lancedb libs ===" && \
    find lib -type f

# Build binary — find the .a file regardless of exact platform path
COPY . .
RUN LANCEDB_MODULE="$(go env GOPATH)/pkg/mod/github.com/lancedb/lancedb-go@v0.1.2" && \
    LANCEDB_LIB=$(find "$LANCEDB_MODULE/lib" -name "liblancedb_go.a" | head -1) && \
    if [ -z "$LANCEDB_LIB" ]; then echo "ERROR: liblancedb_go.a not found"; exit 1; fi && \
    echo "Using: $LANCEDB_LIB" && \
    CGO_CFLAGS="-I${LANCEDB_MODULE}/include" \
    CGO_LDFLAGS="$LANCEDB_LIB -lpthread -ldl -lm" \
    go build -o /ragpack ./cmd/main.go

FROM --platform=linux/amd64 debian:bookworm-slim
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

COPY --from=builder /ragpack /ragpack

ENV SQLITE_PATH=/data/ragpack.db
ENV LANCEDB_PATH=/data/lancedb

EXPOSE 9000

CMD ["/ragpack"]
