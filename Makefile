all: build

.PHONY: build

ifeq ($(GOPATH),)
	PATH := $(HOME)/go/bin:$(PATH)
else
	PATH := $(GOPATH)/bin:$(PATH)
endif

export GO111MODULE=on

BRANCH = $(shell git rev-parse --abbrev-ref HEAD)
REVISION = $(shell git rev-parse HEAD)
REVSHORT = $(shell git rev-parse --short HEAD)
USER = $(shell whoami)
GOVERSION = $(shell go version | awk '{print $$3}')
NOW	= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
SHELL = /bin/sh
VERSION = $(shell git describe --tags --always)

ifneq ($(OS), Windows_NT)
	CURRENT_PLATFORM = linux
	ifeq ($(shell uname), Darwin)
		SHELL := /bin/sh
		CURRENT_PLATFORM = darwin
	endif
else
	CURRENT_PLATFORM = windows
endif

BUILD_VERSION = "\
	-X github.com/johnmikee/cuebert/pkg/version.appName=${APP_NAME} \
	-X github.com/johnmikee/cuebert/pkg/version.version=${VERSION} \
	-X github.com/johnmikee/cuebert/pkg/version.branch=${BRANCH} \
	-X github.com/johnmikee/cuebert/pkg/version.buildUser=${USER} \
	-X github.com/johnmikee/cuebert/pkg/version.buildDate=${NOW} \
	-X github.com/johnmikee/cuebert/pkg/version.revision=${REVISION} \
	-X github.com/johnmikee/cuebert/pkg/version.goVersion=${GOVERSION}"

deps:
	go mod download

test:
	go test -cover ./...

build: cue cuebert

clean:
	rm -rf build/
	rm -f *.zip

.pre-build:
	mkdir -p build/darwin
	mkdir -p build/linux

install-local: \
	install-cue \
	install-cuebert

.pre-cue:
	$(eval APP_NAME = cue)

cue: .pre-build .pre-cue
	go build -o build/$(CURRENT_PLATFORM)/cue -ldflags ${BUILD_VERSION} ./cmd/cue

xp-cue: .pre-build .pre-cue
	GOOS=darwin go build -o build/darwin/cue -ldflags ${BUILD_VERSION} ./cmd/cue
	GOOS=linux CGO_ENABLED=0 go build -o build/linux/cue -ldflags ${BUILD_VERSION} ./cmd/cue

install-cue: .pre-cue
	go install -ldflags ${BUILD_VERSION} ./cmd/cue

APP_NAME = cuebert

.pre-cuebert:
	$(eval APP_NAME = cuebert)

cuebert: .pre-build .pre-cuebert
	go build -o build/$(CURRENT_PLATFORM)/cuebert -ldflags ${BUILD_VERSION} ./cmd/cuebert

install-cuebert: .pre-cuebert
	go install -ldflags ${BUILD_VERSION} ./cmd/cuebert

xp-cuebert: .pre-build .pre-cuebert
	GOOS=darwin go build -o build/darwin/cuebert -ldflags ${BUILD_VERSION} ./cmd/cuebert
	GOOS=linux CGO_ENABLED=0 go build -o build/linux/cuebert -ldflags ${BUILD_VERSION} ./cmd/cuebert

release-zip: xp-cue xp-cuebert
	zip -r cue_${VERSION}.zip build/

docker-cue:
	cp resources/Docker/cue/Dockerfile .
	docker build -t cue --rm .
	rm Dockerfile

run-docker-cue:
	docker run cue cue version

docker-cuebert:
	cp resources/Docker/cuebert/Dockerfile .
	docker build -t cuebert --rm .
	rm Dockerfile

run-docker-cuebert:
	docker run cuebert cuebert -help

clean-postgres-db:
	docker stop pgcue-db-container
	@true
	docker rm pgcue-db-container
	@true

docker-postgres: clean-postgres-db
	cp resources/Docker/postgres/Dockerfile .
	docker build -t pgcue --rm .
	rm Dockerfile

run-docker-postgres: docker-postgres
	docker run -d --name pgcue-db-container -p 5432:5432 pgcue

static-check:
	staticcheck ./...

vet:
	go vet ./...

tidy:
	go mod tidy
