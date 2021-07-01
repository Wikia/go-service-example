.PHONY: build build-alpine clean test help default lint run-local

BIN_NAME=example
GITHUB_REPO=github.com/wikia/go-example-service
BIN_DIR := $(GOPATH)/bin
GOLANGCI_LINT := /usr/local/bin/golangci-lint
CURRENT_DIR := $(shell pwd)

VERSION := $(shell grep "const Version " version/version.go | sed -E 's/.*"(.+)"$$/\1/')
GIT_COMMIT=$(shell git rev-parse HEAD)
GIT_DIRTY=$(shell test -n "`git status --porcelain`" && echo "+CHANGES" || true)
BUILD_DATE=$(shell date '+%Y-%m-%d-%H:%M:%S')
IMAGE_NAME := "artifactory.wikia-inc.com/services/${BIN_NAME}"

default: test

$(GOLANGCI_LINT):
	brew install golangci-lint

help:
	@echo 'Management commands for ${BIN_NAME}:'
	@echo
	@echo 'Usage:'
	@echo '    make build           Compile the project.'
	@echo '    make get-deps        runs dep ensure, mostly used for ci.'
	@echo '    make build-alpine    Compile optimized for alpine linux.'
	@echo '    make package         Build final docker image with just the go binary inside'
	@echo '    make tag             Tag image created by package with latest, git commit and version'
	@echo '    make test            Run tests on a compiled project.'
	@echo '    make push            Push tagged images to registry'
	@echo '    make clean           Clean the directory tree.'
	@echo '    make lint            Run the linter on the source code'
	@echo '    make run-local       Run the server locally with live-reload (using local air binary of docker if not found'
	@echo

lint: $(GOLANGCI_LINT)
	golangci-lint run -E revive -E gosec -E gofmt -E goimports

build:
	@echo "building ${BIN_NAME} ${VERSION}"
	@echo "GOPATH=${GOPATH}"
	go build -ldflags "-X ${GITHUB_REPO}/version.GitCommit=${GIT_COMMIT}${GIT_DIRTY} -X ${GITHUB_REPO}/version.BuildDate=${BUILD_DATE}" -o bin/${BIN_NAME} cmd/server/main.go

get-deps:
	go mod install

build-alpine:
	@echo "building ${BIN_NAME} ${VERSION}"
	@echo "GOPATH=${GOPATH}"
	go build -ldflags '-w -linkmode external -extldflags "-static" -X ${GITHUB_REPO}/version.GitCommit=${GIT_COMMIT}${GIT_DIRTY} -X ${GITHUB_REPO}/version.BuildDate=${BUILD_DATE}' -o bin/${BIN_NAME} cmd/server/main.go

package:
	@echo "building image ${BIN_NAME} ${VERSION} $(GIT_COMMIT)"
	docker build --build-arg VERSION=${VERSION} --build-arg GIT_COMMIT=$(GIT_COMMIT) -t $(IMAGE_NAME):local cmd/server/main.go

tag:
	@echo "Tagging: ${VERSION}"
	docker tag $(IMAGE_NAME):local $(IMAGE_NAME):${VERSION}

push: tag
	@echo "Pushing docker image to registry: ${VERSION}"
	docker push $(IMAGE_NAME):${VERSION}

clean:
	@test ! -e bin/${BIN_NAME} || rm bin/${BIN_NAME}

test:
	go test ./...

run-local:
ifeq (, $(shell which air))
	@echo "Running server using docker air image"
	@docker run -it --rm -w "/example" -e "air_wd=/example" -v ${CURRENT_DIR}:/example -p 3000:3000 -p 4000:4000 cosmtrek/air
else
	@echo "Running server using local air binary"
	@air
endif
