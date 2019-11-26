#!/usr/bin/env bash

set -e

TEST_PATTERN='.'
TEST_OPTIONS='-v -failfast -race'
TEST_TIMEOUT='2m'
SOURCE_FILES='./...'

go test $TEST_OPTIONS $SOURCE_FILES -run $TEST_PATTERN -timeout=$TEST_TIMEOUT
