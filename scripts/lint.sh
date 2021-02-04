#!/usr/bin/env bash

set -e

LINT_VERSION=1.36.0

if [ -z "$ROOT" ]; then
    echo 'Missing environment variables!'
    exit 1
fi

function install {
    curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s v$LINT_VERSION
}

cd "$ROOT"

if [ ! -f "./bin/golangci-lint" ]; then
    echo '> Installing golangci-lint...'
    install
fi

if [[ $("./bin/golangci-lint" --version) != *"$LINT_VERSION"* ]]; then
    echo '> Updating golangci-lint...'
    install
fi

./bin/golangci-lint run --tests=false --enable-all --disable=lll
