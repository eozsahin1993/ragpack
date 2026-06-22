LANCEDB_MODULE := $(shell go env GOPATH)/pkg/mod/github.com/lancedb/lancedb-go@v0.1.2
CGO_CFLAGS     := -I$(LANCEDB_MODULE)/include
CGO_LDFLAGS    := $(LANCEDB_MODULE)/lib/darwin_arm64/liblancedb_go.a -framework Security -framework CoreFoundation

export CGO_CFLAGS
export CGO_LDFLAGS

.PHONY: build run setup

setup:
	chmod -R u+w $(LANCEDB_MODULE)
	cd $(LANCEDB_MODULE) && bash scripts/download-artifacts.sh v0.1.2

build:
	go build ./backend/...

run:
	go run ./backend/cmd/main.go
