#!/usr/bin/env bash

set -e

if [ -z "$ROOT" ]; then
    echo 'Missing environment variables!'
    exit 1
fi

files=$(find "$ROOT" -name '*.go' -not -wholename "$ROOT/vendor/*")

if [ -z "$files" ]; then
    echo "no files found - skipping"
    exit 0
fi

gofmt_files=$(gofmt -l $files)

for file in $gofmt_files; do
    gofmt -w -s "$file"
    goimports -w "$file"
done
