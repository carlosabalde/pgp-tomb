#!/usr/bin/env bash

set -e

IMAGE_NAME='pgp-tomb:latest'
CONTAINER_NAME='pgp-tomb'

if [ -z "$ROOT" ]; then
    echo 'Missing environment variables!'
    exit 1
fi

if [ ! -z "$1" ]; then
    ACTION="$1"
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
        --env GITHUB_TOKEN="$PGP_TOMB_GITHUB_TOKEN" \
        -v $ROOT:/mnt \
        $IMAGE_NAME
fi

if [ "$(docker ps -a -q -f status=exited -f name=$CONTAINER_NAME)" ]; then
    echo '> Starting container...'
    docker container start $CONTAINER_NAME
fi

if [ "$ACTION" == 'shell' ]; then
    echo '> Connecting to container...'
    docker exec \
        --tty \
        --interactive \
        $CONTAINER_NAME /bin/bash
elif [ "$ACTION" == 'travis' ]; then
    echo '> Running Travis stuff...'
    docker exec \
        $CONTAINER_NAME /bin/bash -c '\
            cd /mnt; \
            make test;'
fi
