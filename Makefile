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

build: cuebert

clean:
	rm -rf build/
	rm -f *.zip

.pre-build:
	mkdir -p build/darwin
	mkdir -p build/linux

APP_NAME = cuebert

.pre-cuebert:
	$(eval APP_NAME = cuebert)

cuebert: .pre-build .pre-cuebert
	go build -o build/$(CURRENT_PLATFORM)/cuebert -ldflags ${BUILD_VERSION} ./cuebert

install: .pre-cuebert
	go install -ldflags ${BUILD_VERSION} ./cuebert

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
