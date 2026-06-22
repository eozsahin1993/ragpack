#!/usr/bin/env bash
set -e

LANCEDB_MODULE="$(go env GOPATH)/pkg/mod/github.com/lancedb/lancedb-go@v0.1.2"
LANCEDB_LIB="${LANCEDB_MODULE}/lib/darwin_arm64/liblancedb_go.a"

# Download native lancedb library if not already present
if [ ! -f "$LANCEDB_LIB" ]; then
  echo "lancedb native library not found, downloading..."
  chmod -R u+w "$LANCEDB_MODULE"
  (cd "$LANCEDB_MODULE" && bash scripts/download-artifacts.sh v0.1.2)
fi

export CGO_CFLAGS="-I${LANCEDB_MODULE}/include"
export CGO_LDFLAGS="${LANCEDB_LIB} -framework Security -framework CoreFoundation"

echo "Starting RagPack dev server..."
go run ./cmd/main.go
