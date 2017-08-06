#!/bin/bash

set -e

PACKAGE="github.com/joelanford/scm-bot/app"
APP_NAME="scm-bot"
TAG=$(git tag -l --points-at HEAD)
VERSION=${TAG:-$(git describe --always --dirty --long)}
BUILD_TIME=$(date -u '+%Y-%m-%d %H:%M:%S.%N %z %Z')
USER=${USER:-$USERNAME}
GIT_HASH=$(git rev-parse HEAD)
GIT_BRANCH=${TRAVIS_BRANCH:-$(git rev-parse --abbrev-ref HEAD)}

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
    if [[ ${GIT_BRANCH} == feature/* ]]; then
        echo "Skipping push for feature brach \"${GIT_BRANCH}\""
        return 0
    fi

    docker push ${DOCKER_REPO}:${VERSION}
    if [[ ${GIT_BRANCH} == "develop" ]]; then
        docker tag ${DOCKER_REPO}:${VERSION} ${DOCKER_REPO}:alpha
        docker push ${DOCKER_REPO}:alpha
    elif [[ ${GIT_BRANCH} == release/* ]]; then
        docker tag ${DOCKER_REPO}:${VERSION} ${DOCKER_REPO}:beta
        docker push ${DOCKER_REPO}:beta
    elif [[ ${GIT_BRANCH} == "master" ]]; then
        docker tag ${DOCKER_REPO}:${VERSION} ${DOCKER_REPO}:latest
        docker push ${DOCKER_REPO}:latest
    fi
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
