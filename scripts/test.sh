#!/usr/bin/env bash

set -e

if [ -z "$ROOT" ]; then
    echo 'Missing environment variables!'
    exit 1
fi

TEST_OPTIONS="-failfast -race -coverprofile=$ROOT/coverage.txt -covermode=atomic"
TEST_TIMEOUT='2m'
SOURCE_FILES='./...'

if [ -z "$TEST_PATTERN" ]; then
    TEST_PATTERN='.'
fi

go test $TEST_OPTIONS $SOURCE_FILES -run $TEST_PATTERN -timeout=$TEST_TIMEOUT
