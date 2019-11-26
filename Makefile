SHELL = /bin/bash

ROOT := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))

PROJECT := pgp-tomb
VERSION := $(shell cat VERSION)
XC_OS := linux darwin
XC_ARCH := 386 amd64 arm
GO111MODULE := on

export ROOT
export PROJECT
export VERSION
export XC_OS
export XC_ARCH
export GO111MODULE

.PHONY: all
all: help

.PHONY: help
help:
	@echo 'make docker - launch Docker shell'
	@echo 'make build - build $(PROJECT) for follwing OS-ARCH constilations: $(XC_OS) / $(XC_ARCH) '
	@echo 'make build-dev - build $(PROJECT) for OS-ARCH set by GOOS and GOARCH env variables'
	@echo 'make fmt - run gofmt & goimports'
	@echo 'make lint - run golangci-lint'
	@echo 'make test - run tests'
	@echo 'make dist - build & create packages with hashsums'
	@echo 'make mod - run assorted modules stuff'

.PHONY: docker
docker:
	@scripts/docker.sh

.PHONY: build
build:
	@scripts/build.sh

.PHONY: build-dev
build-dev:
	@scripts/build.sh dev

.PHONY: dist
dist:
	@scripts/dist.sh

.PHONY: fmt
fmt:
	@scripts/fmt.sh

.PHONY: lint
lint:
	@scripts/lint.sh

.PHONY: test
test:
	@scripts/test.sh

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
