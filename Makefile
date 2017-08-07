BUILDVAR_PACKAGE := github.com/joelanford/scm-bot/app
APP_NAME         := scm-bot

DOCKER_REPO  := joelanford
DOCKER_IMAGE := $(DOCKER_REPO)/$(APP_NAME)

TAG        := $(shell git tag -l --points-at HEAD)
VERSION    := $(if $(TAG),$(TAG),$(shell git describe --always --dirty --long))
GIT_HASH   := $(shell git rev-parse HEAD)
GIT_BRANCH := $(if $(TRAVIS_BRANCH),$(TRAVIS_BRANCH),$(shell git rev-parse --abbrev-ref HEAD))

BUILD_TIME := $(shell date -u '+%Y-%m-%d %H:%M:%S.%N %z %Z')
USER       := $(if $(USER),$(USER),$(USERNAME))

.PHONY: all
all: go

ifneq (,$(findstring docker,$(MAKECMDGOALS)))
    GOOS = linux
endif

.PHONY: go
go:
	CGO_ENABLED=0 GOOS=${GOOS} go build -ldflags " \
		-extldflags '-static' \
		-X '${BUILDVAR_PACKAGE}.appName=${APP_NAME}' \
		-X '${BUILDVAR_PACKAGE}.version=${VERSION}' \
		-X '${BUILDVAR_PACKAGE}.buildTime=${BUILD_TIME}' \
		-X '${BUILDVAR_PACKAGE}.buildUser=${USER}' \
		-X '${BUILDVAR_PACKAGE}.gitHash=${GIT_HASH}' \
		" -o ${APP_NAME}

.PHONY: docker-image
docker-image: go
	docker build -f Dockerfile -t ${DOCKER_IMAGE}:${VERSION} .

.PHONY: docker-push
docker-push: docker-image
	@if [[ $(GIT_BRANCH) == feature/* ]]; then \
		echo "Skipping push for feature brach \"${GIT_BRANCH}\""; \
		return 0; \
	fi

	docker push ${DOCKER_IMAGE}:${VERSION}

	@if [[ ${GIT_BRANCH} == "develop" ]]; then \
		echo docker tag ${DOCKER_IMAGE}:${VERSION} ${DOCKER_IMAGE}:alpha; \
		docker tag ${DOCKER_IMAGE}:${VERSION} ${DOCKER_IMAGE}:alpha; \
		echo docker push ${DOCKER_IMAGE}:alpha; \
		docker push ${DOCKER_IMAGE}:alpha; \
	elif [[ ${GIT_BRANCH} == release/* ]]; then \
		echo docker tag ${DOCKER_IMAGE}:${VERSION} ${DOCKER_IMAGE}:beta; \
		docker tag ${DOCKER_IMAGE}:${VERSION} ${DOCKER_IMAGE}:beta; \
		echo docker push ${DOCKER_IMAGE}:beta; \
		docker push ${DOCKER_IMAGE}:beta; \
	elif [[ ${GIT_BRANCH} == "master" ]]; then \
		echo docker tag ${DOCKER_IMAGE}:${VERSION} ${DOCKER_IMAGE}:latest; \
		docker tag ${DOCKER_IMAGE}:${VERSION} ${DOCKER_IMAGE}:latest; \
		echo docker push ${DOCKER_IMAGE}:latest; \
		docker push ${DOCKER_IMAGE}:latest; \
	fi

.PHONY: version
version:
	@echo ${VERSION}	

.PHONY: clean
clean:
	rm -f ${APP_NAME}
	docker rmi -f `docker images | grep ${DOCKER_IMAGE} | awk '{print $3}'` 2>/dev/null || true