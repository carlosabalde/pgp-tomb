#!/usr/bin/env bash

set -e

if [ -z "$ROOT" ]; then
    echo 'Missing environment variables!'
    exit 1
fi

TEST_PATTERN='.'
TEST_OPTIONS="-v -failfast -race -coverprofile=$ROOT/coverage.txt -covermode=atomic"
TEST_TIMEOUT='2m'
SOURCE_FILES='./...'

go test $TEST_OPTIONS $SOURCE_FILES -run $TEST_PATTERN -timeout=$TEST_TIMEOUT
