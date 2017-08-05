#!/bin/sh

VERSION=$(git describe --always --dirty --long)
BUILDTIME=$(date -u '+%Y-%m-%d %H:%M:%S.%N %z %Z')
USER=${USER:=$USERNAME}
GITHASH=$(git rev-parse HEAD)

PACKAGE="github.com/joelanford/scm-bot/app"
APPNAME="scm-bot"

go build -ldflags " \
    -X '${PACKAGE}.appName=${APPNAME}' \
    -X '${PACKAGE}.version=${VERSION}' \
    -X '${PACKAGE}.buildTime=${BUILDTIME}' \
    -X '${PACKAGE}.buildUser=${USER}' \
    -X '${PACKAGE}.gitHash=${GITHASH}' \
    " -o ${APPNAME}
