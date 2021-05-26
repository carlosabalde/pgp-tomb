SHELL := /bin/bash

ROOT := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
UMASK := 022

PROJECT := pgp-tomb
VERSION := 0.3.9

DOCKER_IMAGE_NAME := $(PROJECT):latest
DOCKER_CONTAINER_NAME := $(PROJECT)

BUILD_LD_FLAGS := \
    -X github.com/carlosabalde/pgp-tomb/internal/core/config.version=$(VERSION) \
    -X github.com/carlosabalde/pgp-tomb/internal/core/config.revision=$(shell git rev-parse --short HEAD) \
    -s -w
BUILD_OSS := linux darwin windows
BUILD_ARCHS := 386 amd64 arm
BUILD_DEV_OS := $(shell go env GOOS)
BUILD_DEV_ARCH := $(shell go env GOARCH)

TEST_PATTERN ?= .

GO111MODULE := on
export GO111MODULE

.PHONY: all
all: help

.PHONY: help
help:
	@( \
		echo 'Host targets:'; \
		echo '  shell - launch Docker shell.'; \
		echo '  travis - run Travis stuff in Docker container.'; \
		echo; \
		echo 'Container targets:'; \
		echo '  build - build for the following OS-architecture pairs: $(BUILD_OSS) / $(BUILD_ARCHS).'; \
		echo '  build-dev - build for the current OS ($(BUILD_DEV_OS)) & architecture ($(BUILD_DEV_ARCH)).'; \
		echo '  fmt - run goimports.'; \
		echo '  lint - run golangci-lint.'; \
		echo '  test - run tests.'; \
		echo '  mod - run assorted modules stuff.'; \
		echo '  dist - build & create .tar.gz packages with hashsums.'; \
		echo '  mrproper - clean up temporary files.'; \
	)

.PHONY: docker
docker:
	@( \
		set -e; \
		\
		if [ ! "$$(docker images -q $(DOCKER_IMAGE_NAME) 2> /dev/null)" ]; then \
			echo '> Building Docker image...'; \
			cd '$(ROOT)'; \
			docker build -t $(DOCKER_IMAGE_NAME) .; \
		fi; \
		\
		if [ ! "$$(docker ps -a -q -f name=$(DOCKER_CONTAINER_NAME))" ]; then \
			echo '> Building Docker container...'; \
			docker run \
				--detach \
				--name $(DOCKER_CONTAINER_NAME) \
				--env GITHUB_TOKEN="$$PGP_TOMB_GITHUB_TOKEN" \
				-v $(ROOT):/mnt \
				$(DOCKER_IMAGE_NAME); \
		fi; \
		\
		if [ "$$(docker ps -a -q -f status=exited -f name=$(DOCKER_CONTAINER_NAME))" ]; then \
			echo '> Starting Docker container...'; \
			docker container start $(DOCKER_CONTAINER_NAME); \
		fi; \
	)

.PHONY: shell
shell: docker
	@( \
		set -e; \
		\
		echo '> Launching shell in Docker container...'; \
		docker exec \
			--tty \
			--interactive \
			--workdir /mnt \
			$(DOCKER_CONTAINER_NAME) /bin/bash; \
	)

.PHONY: travis
travis: docker
	@( \
		set -e; \
		\
		echo '> Running Travis stuff in Docker container...'; \
		docker exec \
			--workdir /mnt \
			$(DOCKER_CONTAINER_NAME) /bin/bash -c ' \
				set -e; \
				make test'; \
	)

.PHONY: build
build: mrproper
	@( \
		set -e; \
		\
		cd '$(ROOT)'; \
		mkdir -p build; \
		\
		echo '> Building...'; \
		export CGO_ENABLED=0; \
		for CMD in $$(ls ./cmd); do \
			for OS in $(BUILD_OSS); do \
				for ARCH in $(BUILD_ARCHS); do \
					if ([ $$OS == 'darwin' ] && ([ $$ARCH == '386' ] || [ $$ARCH == 'arm' ])) || \
					   ([ $$OS == 'windows' ] && [ $$ARCH == 'arm' ]); then \
						continue; \
					fi; \
					\
					SUFFIX=''; \
					if [ $$OS == 'windows' ]; then \
						SUFFIX='.exe'; \
					fi; \
					\
					echo "  - $$CMD ($$OS-$$ARCH)"; \
					GOOS=$$OS GOARCH=$$ARCH go build \
						-ldflags '$(BUILD_LD_FLAGS)' \
						-o build/$$OS-$$ARCH/$$CMD$$SUFFIX \
						./cmd/$$CMD/; \
				done; \
			done; \
		done; \
	)

.PHONY: build-dev
build-dev: BUILD_OSS := $(BUILD_DEV_OS)
build-dev: BUILD_ARCHS := $(BUILD_DEV_ARCH)
build-dev: build

.PHONY: fmt
fmt:
	@( \
		set -e; \
		\
		if [ ! -f '/go/bin/goimports' ]; then \
			echo '> Installing goimports...'; \
			go install golang.org/x/tools/cmd/goimports@v0.1.0; \
		fi; \
		\
		echo '> Running goimports...'; \
		FILES=$$(find '$(ROOT)' -name '*.go' -not -wholename '$(ROOT)/vendor/*'); \
		for FILE in $$(/go/bin/goimports -l $$FILES); do \
			echo "- $$FILE"; \
			/go/bin/goimports -w "$$FILE"; \
		done; \
	)

.PHONY: lint
lint:
	@( \
		set -e; \
		\
		if [ ! -f '/go/bin/golangci-lint' ]; then \
			echo '> Installing golangci-lint...'; \
			curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b /go/bin v1.38.0; \
		fi; \
		\
		echo '> Running golangci-lint...'; \
		/go/bin/golangci-lint run '$(ROOT)/...'; \
	)

.PHONY: test
test:
	@( \
		set -e; \
		\
		echo '> Running tests...'; \
		go test \
			-v -failfast -race \
			-coverprofile=$(ROOT)/coverage.txt \
			-covermode=atomic \
			'$(ROOT)/...' -run '$(TEST_PATTERN)' -timeout=2m; \
	)

.PHONY: mod
mod:
	@( \
		set -e; \
		\
		echo '> Adding missing and removing unused modules...'; \
		go mod tidy; \
		\
		echo '> Making vendored copy of dependencies...'; \
		go mod vendor; \
		\
		echo '> Printing module requirement graph...'; \
		go mod graph; \
	)

.PHONY: dist
dist: build
	@( \
		set -e; \
		\
		cd '$(ROOT)'; \
		mkdir -p dist; \
		\
		echo '> Packaging...'; \
		for BUILD in $$(ls ./build); do \
			echo "  - $$BUILD"; \
			tar -czf ./dist/$(PROJECT)-$(VERSION).$$BUILD.tar.gz -C ./build/$$BUILD $$(ls ./build/$$BUILD); \
		done; \
		\
		echo '> Signing...'; \
		for DIST in $$(ls ./dist); do \
			echo "  - $$DIST"; \
			NAME=$${DIST%.tar.gz}; \
			shasum -a 256 "$(ROOT)/dist/$$NAME.tar.gz" > "$(ROOT)/dist/$$NAME.sha256"; \
			md5sum "$(ROOT)/dist/$$NAME.tar.gz" > "$(ROOT)/dist/$$NAME.md5"; \
		done; \
	)

.PHONY: mrproper
mrproper:
	@( \
		echo '> Cleaning up...'; \
		rm -rf '$(ROOT)/build' '$(ROOT)/dist'; \
		git clean -f -x -d '$(ROOT)'; \
	)
