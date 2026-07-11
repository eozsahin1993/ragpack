#!/usr/bin/env bash
set -e

LANCEDB_VERSION="v1.0.4"
LANCEDB_MODULE="$(go env GOPATH)/pkg/mod/github.com/eozsahin1993/lancedb-go@${LANCEDB_VERSION}"

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
[ "$ARCH" = "x86_64" ] && ARCH="amd64"
[ "$ARCH" = "aarch64" ] && ARCH="arm64"
PLATFORM="${OS}_${ARCH}"

LANCEDB_LIB="${LANCEDB_MODULE}/lib/${PLATFORM}/liblancedb_go.a"

# Redirected to stderr: --print-env's stdout must be only the two `export`
# lines below, since callers do `source <(./dev.sh --print-env)` — on a cold
# cache (nothing downloaded yet), download-artifacts.sh's progress/log
# output would otherwise land on stdout too and get parsed as shell commands.
if [ ! -f "$LANCEDB_LIB" ]; then
  echo "lancedb native library not found, downloading..." >&2
  chmod -R u+w "$LANCEDB_MODULE"
  (cd "$LANCEDB_MODULE" && bash scripts/download-artifacts.sh "$LANCEDB_VERSION") >&2
fi

export CGO_CFLAGS="-I${LANCEDB_MODULE}/include"

case "$OS" in
  darwin)
    export CGO_LDFLAGS="${LANCEDB_LIB} -lbz2 -framework Security -framework CoreFoundation"
    ;;
  linux)
    export CGO_LDFLAGS="${LANCEDB_LIB} -lpthread -ldl -lm -lbz2"
    ;;
esac

# --print-env: emit `export` lines instead of running the server, so other
# commands (integration tests, CI) can pick up the same CGO env — e.g.
# `source <(./dev.sh --print-env)`.
if [ "$1" = "--print-env" ]; then
  echo "export CGO_CFLAGS=\"${CGO_CFLAGS}\""
  echo "export CGO_LDFLAGS=\"${CGO_LDFLAGS}\""
  exit 0
fi

echo "Starting RagPack dev server (${PLATFORM})..."
go run ./cmd/main.go
