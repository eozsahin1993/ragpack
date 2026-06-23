#!/usr/bin/env bash
set -e

LANCEDB_VERSION="v0.1.4"
LANCEDB_MODULE="$(go env GOPATH)/pkg/mod/github.com/eozsahin1993/lancedb-go@${LANCEDB_VERSION}"

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
[ "$ARCH" = "x86_64" ] && ARCH="amd64"
[ "$ARCH" = "aarch64" ] && ARCH="arm64"
PLATFORM="${OS}_${ARCH}"

LANCEDB_LIB="${LANCEDB_MODULE}/lib/${PLATFORM}/liblancedb_go.a"

if [ ! -f "$LANCEDB_LIB" ]; then
  echo "lancedb native library not found, downloading..."
  chmod -R u+w "$LANCEDB_MODULE"
  (cd "$LANCEDB_MODULE" && bash scripts/download-artifacts.sh "$LANCEDB_VERSION")
fi

export CGO_CFLAGS="-I${LANCEDB_MODULE}/include"

case "$OS" in
  darwin)
    export CGO_LDFLAGS="${LANCEDB_LIB} -framework Security -framework CoreFoundation"
    ;;
  linux)
    export CGO_LDFLAGS="${LANCEDB_LIB} -lpthread -ldl -lm"
    ;;
esac

echo "Starting RagPack dev server (${PLATFORM})..."
go run ./cmd/main.go
