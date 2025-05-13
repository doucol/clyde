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
VER ?= $(NEXT)-dev-$(REV)

default: help

clean: ## Clean bin & test cache
	@rm $(OUT) && go clean --testcache

test: ## Run tests
	@go clean --testcache && go test ./...

build: ## Build
	@CGO_ENABLED=0 go build -ldflags "-X ${SRC}/cmd.date=${DATE} -X ${SRC}/cmd.revision=${REV} -X ${SRC}/cmd.version=${VER} -w -s" -a -o ${OUT} main.go

snapshot: ## GoReleaser snapshot
	@goreleaser release --clean --snapshot

devrel: ## dev release v0.0.1-dev-<rev>
	@git tag -f -a $(VER) -m "Release $(VER)"
	@git push origin tag $(VER)
	
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":[^:]*?## "}; {printf "\033[38;5;69m%-30s\033[38;5;38m %s\033[0m\n", $$1, $$2}'
