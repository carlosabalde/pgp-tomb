#!/usr/bin/env bash

set -e

IMAGE_NAME='pgp-tomb-golang:latest'
CONTAINER_NAME='pgp-tomb-golang'

if [ -z "$ROOT" ]; then
    echo 'Missing environment variables!'
    exit 1
fi

cd "$ROOT"

if [ ! "$(docker images -q $IMAGE_NAME 2> /dev/null)" ]; then
    echo '> Building image...'
    docker build -t $IMAGE_NAME .
fi

if [ ! "$(docker ps -a -q -f name=$CONTAINER_NAME)" ]; then \
    echo '> Building container...'
    docker run \
        --detach \
        --name $CONTAINER_NAME \
        -v $ROOT:/mnt \
        $IMAGE_NAME
fi

if [ "$(docker ps -a -q -f status=exited -f name=$CONTAINER_NAME)" ]; then
    echo '> Starting container...'
    docker container start $CONTAINER_NAME
fi

echo '> Connecting to container...'
docker exec \
    --tty \
    --interactive \
    $CONTAINER_NAME /bin/bash
