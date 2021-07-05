.PHONY: build build-alpine clean test help default lint run-local bump-version

BIN_NAME = example
GITHUB_REPO = github.com/wikia/go-example-service
BIN_DIR := $(GOPATH)/bin
GOLANGCI_LINT := /usr/local/bin/golangci-lint
GORELEASER := /usr/local/bin/goreleaser
CURRENT_DIR := $(shell pwd)

VERSION := $(shell git describe --tags --exact-match 2>/dev/null || echo "unknown")
GIT_COMMIT = $(shell git rev-parse --short HEAD)
GIT_DIRTY = $(shell test -n "`git status --porcelain`" && echo "+CHANGES" || true)
BUILD_DATE = $(shell date '+%Y-%m-%d-%H:%M:%S')
IMAGE_NAME := "artifactory.wikia-inc.com/services/${BIN_NAME}"

default: test

$(GOLANGCI_LINT):
	brew install golangci-lint

$(GORELEASER):
	brew install goreleaser/tap/goreleaser

help:
	@echo 'Management commands for ${BIN_NAME}:'
	@echo
	@echo 'Usage:'
	@echo '    make build           	Compile the project.'
	@echo '    make build-docker    	Build all releases and docker images but will not publish them.'
	@echo '    make release         	Build all releases and docker image and pushes them out'
	@echo '    make get-deps        	Runs `go mod install`, mostly used for ci.'
	@echo '    make build-alpine    	Compile optimized for alpine linux.'
	@echo '    make test            	Run tests on a compiled project.'
	@echo '    make clean           	Clean the directory tree.'
	@echo '    make lint            	Run the linter on the source code'
	@echo '    make openapi-generate 	Will generate server/client code using OpenAPI schema'
	@echo '    make run-local       	Run the server locally with live-reload (using local air binary of docker if not found'
	@echo

lint: $(GOLANGCI_LINT) lint-cmd

lint-ci: lint-cmd

lint-cmd:
	golangci-lint run -E revive -E gosec -E gofmt -E goimports
build:
	@echo "building ${BIN_NAME}@${VERSION}"
	go build -ldflags "-X main.commit=${GIT_COMMIT}${GIT_DIRTY} -X main.date=${BUILD_DATE} -X main.version=${VERSION}" -o bin/${BIN_NAME} cmd/main.go

get-deps:
	go mod install

build-alpine:
	@echo "building ${BIN_NAME}@${VERSION}"
	go build -ldflags '-w -linkmode external -extldflags "-static" -X main.commit=${GIT_COMMIT}${GIT_DIRTY} -X main.date=${BUILD_DATE} -X main.version=${VERSION}' -o bin/${BIN_NAME} cmd/main.go

build-docker: $(GORELEASER)
	goreleaser --snapshot --skip-publish --rm-dist

release: $(GORELEASER)
	goreleaser --rm-dist

clean:
	@test ! -e bin/${BIN_NAME} || rm bin/${BIN_NAME}
	@docker compose down

test:
	go test ./...

openapi-generate:
	@oapi-codegen -config ./cmd/openapi/server.cfg.yaml ./cmd/openapi/schema.yaml
	@oapi-codegen -config ./cmd/openapi/types.cfg.yaml ./cmd/openapi/schema.yaml

run-local:
	@echo "Running server using docker air image"
	@docker compose up --remove-orphans
