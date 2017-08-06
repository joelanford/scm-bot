#!/bin/bash

set -e

PACKAGE="github.com/joelanford/scm-bot/app"
APP_NAME="scm-bot"
VERSION=$(git describe --always --dirty --long)
BUILD_TIME=$(date -u '+%Y-%m-%d %H:%M:%S.%N %z %Z')
USER=${USER:=$USERNAME}
GIT_HASH=$(git rev-parse HEAD)
GIT_BRANCH=${TRAVIS_BRANCH:=$(git rev-parse --abbrev-ref HEAD)}

DOCKER_REPO=joelanford/scm-bot

function buildGo() {
    export CGO_ENABLED=0
    go build -ldflags " \
        -extldflags '-static' \
        -X '${PACKAGE}.appName=${APP_NAME}' \
        -X '${PACKAGE}.version=${VERSION}' \
        -X '${PACKAGE}.buildTime=${BUILD_TIME}' \
        -X '${PACKAGE}.buildUser=${USER}' \
        -X '${PACKAGE}.gitHash=${GIT_HASH}' \
        " -o ${APP_NAME}
}

function buildDocker() {
    docker build -f Dockerfile -t $DOCKER_REPO:${VERSION} .
}

function pushDocker() {
    [[ ${GIT_BRANCH} != feature/* ]] && docker push ${DOCKER_REPO}:${VERSION} || return 0
    [[ ${GIT_BRANCH} == "master" ]] && docker tag ${DOCKER_REPO}:${VERSION} ${DOCKER_REPO}:latest && docker push ${DOCKER_REPO}:latest
    [[ ${GIT_BRANCH} == release/* ]] && docker tag ${DOCKER_REPO}:${VERSION} ${DOCKER_REPO}:beta && docker push ${DOCKER_REPO}:beta
    [[ ${GIT_BRANCH} == "develop" ]] && docker tag ${DOCKER_REPO}:${VERSION} ${DOCKER_REPO}:alpha && docker push ${DOCKER_REPO}:alpha
}

if [[ $1 == "version" ]]; then
    echo ${VERSION}
elif [[ $1 == "go" || -z "$1" ]]; then
    buildGo
elif [[ $1 == "docker" ]]; then
    buildGo
    buildDocker
    if [[ $2 == "push" ]]; then
        pushDocker
    fi
fi
