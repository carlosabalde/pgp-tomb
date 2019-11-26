#!/usr/bin/env bash

set -e

if [ -z "$VERSION" -o -z "$ROOT" ]; then
    echo 'Missing environment variables!'
    exit 1
fi

if [ ! -d './build' ]; then
    echo "Build directory does not exists. Please run 'make build' first."
    exit 1
fi


echo '> Cleaning up...'
cd "$ROOT"
rm -rf dist && mkdir dist

echo
echo '> Packaging...'
for build in $(ls ./build); do
    tar -czf ./dist/$build.tar.gz -C ./build/$build $(ls ./build/$build)
    echo "- $build finished"
done

echo
echo '> Signing...'
for dist in $(ls ./dist); do
    name=${dist%.tar.gz}
    cd "$ROOT/dist"
    shasum -a 256 "$name.tar.gz" > "$name.sha256"
    md5sum "$name.tar.gz" > "$name.md5"
    echo "- $name finished"
done
