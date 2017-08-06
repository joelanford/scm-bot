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
    [[ ${GIT_BRANCH} != feature/* ]] && docker push ${REPO}:${VERSION}
    [[ ${GIT_BRANCH} == "master" ]] && docker tag ${REPO}:${VERSION} ${REPO}:latest && docker push ${REPO}:latest
    [[ ${GIT_BRANCH} == release/* ]] && docker tag ${REPO}:${VERSION} ${REPO}:beta && docker push ${REPO}:beta
    [[ ${GIT_BRANCH} == "develop" ]] && docker tag ${REPO}:${VERSION} ${REPO}:alpha && docker push ${REPO}:alpha
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
