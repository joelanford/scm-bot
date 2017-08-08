SHELL := /bin/bash

BUILDVAR_PACKAGE := github.com/joelanford/scm-bot/app
APP_NAME         := scm-bot
GOOS             := 

DOCKER_REPO  := joelanford
DOCKER_IMAGE := $(DOCKER_REPO)/$(APP_NAME)

TAG        ?= $(shell git tag -l --points-at HEAD)
VERSION    := $(if $(TAG),$(TAG),$(shell git describe --always --dirty --long))
GIT_HASH   ?= $(shell git rev-parse HEAD)
GIT_BRANCH ?= $(shell git rev-parse --abbrev-ref HEAD)

BUILD_TIME := $(shell date -u '+%Y-%m-%d %H:%M:%S.%N %z %Z')
USER       := $(if $(USER),$(USER),$(USERNAME))

.PHONY: all
all: go

.PHONY: build
build:
	CGO_ENABLED=0 GOOS=${GOOS} go build -ldflags " \
		-extldflags '-static' \
		-X '${BUILDVAR_PACKAGE}.appName=${APP_NAME}' \
		-X '${BUILDVAR_PACKAGE}.version=${VERSION}' \
		-X '${BUILDVAR_PACKAGE}.buildTime=${BUILD_TIME}' \
		-X '${BUILDVAR_PACKAGE}.buildUser=${USER}' \
		-X '${BUILDVAR_PACKAGE}.gitHash=${GIT_HASH}' \
		" -o ${APP_NAME}

.PHONY: test
test:
	CGO_ENABLED=0 GOOS=${GOOS} go test `go list ./... | grep -v vendor`

.PHONY: image
image: GOOS := linux
image:
	docker build -f Dockerfile -t ${DOCKER_IMAGE}:${VERSION} .

.PHONY: push
push:
	@if [[ ${GIT_BRANCH} == feature/* ]]; then \
		echo "Skipping push for feature brach \"${GIT_BRANCH}\""; \
	else \
		docker push ${DOCKER_IMAGE}:${VERSION}; \
		EXTRA_IMAGE_TAG=""; \
		if [[ ${GIT_BRANCH} == "develop" ]]; then \
			EXTRA_IMAGE_TAG=alpha; \
		elif [[ ${GIT_BRANCH} == release/* ]]; then \
			EXTRA_IMAGE_TAG=beta; \
		elif [[ ${GIT_BRANCH} == "master" ]]; then \
			EXTRA_IMAGE_TAG=latest; \
		fi; \
		if [[ -n "$${EXTRA_IMAGE_TAG}" ]]; then \
			echo docker tag ${DOCKER_IMAGE}:${VERSION} ${DOCKER_IMAGE}:$${EXTRA_IMAGE_TAG}; \
			docker tag ${DOCKER_IMAGE}:${VERSION} ${DOCKER_IMAGE}:$${EXTRA_IMAGE_TAG}; \
			echo docker push ${DOCKER_IMAGE}:$${EXTRA_IMAGE_TAG}; \
			docker push ${DOCKER_IMAGE}:$${EXTRA_IMAGE_TAG}; \
		fi; \
	fi

.PHONY: version
version:
	@echo ${VERSION}	

.PHONY: clean
clean:
	rm -f ${APP_NAME}
	docker rmi -f `docker images | grep ${DOCKER_IMAGE} | awk '{print $$3}'` 2>/dev/null || true