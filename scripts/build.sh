#!/usr/bin/env bash

set -e

if [ -z "$VERSION" -o -z "$ROOT" ]; then
    echo 'Missing environment variables!'
    exit 1
fi

if [ ! -z "$1" ]; then
    ACTION="$1"
fi

export LD_FLAGS="-X main.version=$VERSION -X main.revision=$(git rev-parse --short HEAD) -s -w"

# Instruct to build statically linked binaries.
export CGO_ENABLED=0

if [ "$ACTION" == 'dev' ]; then
    XC_OS=$(go env GOOS)
    XC_ARCH=$(go env GOARCH)
fi

echo '> Cleaning up...'
cd "$ROOT"
rm -rf build && mkdir build

echo '> Building...'
if [ -d ./cmd ]; then
    for cmd in ./cmd/*; do
        if [ ! -d "$cmd" ]; then
            continue
        fi
        for OS in $XC_OS; do
            for ARCH in $XC_ARCH; do
                if ([ $OS == 'darwin' ] && ([ $ARCH == '386' ] || [ $ARCH == 'arm' ])) ||
                   ([ $OS == 'windows' ] && [ $ARCH == 'arm' ]); then
                    continue
                fi
                GOOS=$OS GOARCH=$ARCH go build -ldflags "$LD_FLAGS" -o build/$OS-$ARCH/${cmd##*/} ./cmd/${cmd##*/}
                echo "- $OS-$ARCH finished"
            done
        done
    done
fi
