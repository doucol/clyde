REV := $(shell git rev-parse --short HEAD)
SEC := $(shell date +%s)
ifeq ($(shell uname), Darwin)
  DATE := $(shell TZ=UTC date -j -f "%s" ${SEC} +"%Y-%m-%dT%H:%M:%SZ")
else
  DATE := $(shell date -u -d @${SEC} +"%Y-%m-%dT%H:%M:%SZ")
endif

NEXT ?= v0.3

APP := clyde
OUT ?= bin/$(APP)
SRC := github.com/doucol/$(APP)
VER ?= $(NEXT)-0.dev-$(REV)

.PHONY: default
default: help

.PHONY: clean
clean: ## Clean bin & test cache
	@rm $(OUT) 2>/dev/null || true
	@go clean --testcache

.PHONY: test
test: ## Run tests
	@go clean --testcache && go test ./...

.PHONY: build
build: ## Build
	@CGO_ENABLED=0 go build -ldflags "-X ${SRC}/cmd.date=${DATE} -X ${SRC}/cmd.revision=${REV} -X ${SRC}/cmd.version=${VER} -w -s" -a -o ${OUT} main.go

.PHONY: snapshot
snapshot: ## GoReleaser snapshot
	@goreleaser release --clean --snapshot

.PHONY: devrel
devrel: ## dev release v0.0.1-dev-<rev>
	@git tag -f -a $(VER) -m "Release $(VER)"
	@git push origin tag $(VER)
	
.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":[^:]*?## "}; {printf "\033[38;5;69m%-30s\033[38;5;38m %s\033[0m\n", $$1, $$2}'
