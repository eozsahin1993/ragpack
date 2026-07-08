#!/usr/bin/env bash
set -e

LANCEDB_VERSION="v1.0.4"
LANCEDB_MODULE="$(go env GOPATH)/pkg/mod/github.com/eozsahin1993/lancedb-go@${LANCEDB_VERSION}"
ARCH=$(go env GOARCH)

CGO_CFLAGS="-I${LANCEDB_MODULE}/include" \
CGO_LDFLAGS="${LANCEDB_MODULE}/lib/linux_${ARCH}/liblancedb_go.a -lpthread -ldl -lm" \
go build -o ./tmp/ragpack ./cmd
