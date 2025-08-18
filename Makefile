APP_NAME=heheswitch
PKG=./cmd/server
DIST=dist
DATE:=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)
GIT_SHA:=$(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
LDFLAGS=-s -w -X main.buildDate=$(DATE) -X main.commit=$(GIT_SHA)

.PHONY: help build build-arm64 build-linux-arm64 run clean

help:
	@echo "Targets:"
	@echo "  build                - build local (current OS/ARCH)"
	@echo "  build-arm64          - build linux arm64 binary (alias build-linux-arm64)"
	@echo "  run                  - run locally"
	@echo "  clean                - remove dist directory"

build:
	CGO_ENABLED=0 go build -ldflags '$(LDFLAGS)' -o $(DIST)/$(APP_NAME) $(PKG)

build-arm64 build-linux-arm64:
	@mkdir -p $(DIST)
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags '$(LDFLAGS)' -o $(DIST)/$(APP_NAME)-linux-arm64 $(PKG)

run: build
	./$(DIST)/$(APP_NAME)

clean:
	rm -rf $(DIST)
