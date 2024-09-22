PROJECT_NAME              = ash
PROJECT_GROUP             = jkf
ARTIFACTORY_PROJECT_GROUP = jkf
GOARCH                    ?= amd64
GO                        = go
BUILD_TYPE                = prod

# build output
BINARY           ?= ${PROJECT_NAME}x
BUILD_IMAGE_NAME ?= ${PROJECT_GROUP}/$PROJECT_NAME}
BUILD_NUMBER     = $(shell git rev-list --count HEAD)
BUILD_RELEASE    ?= "false"
TARGET           ?= us.figge.assh

# workspace
WORKSPACE ?= $(realpath $(dir $(realpath $(firstword $(MAKEFILE_LIST)))))
GOPATH  ?= ${WORKSPACE}/vendor
BIN     = ${GOPATH}/bin
BUILD_DIR = ${WORKSPACE}/build
OUTPUT_DIR ?= ${BUILD_DIR}/bin

VERSION = $(shell git describe --tags --abbrev=0)
COMMIT  = $(shell git rev-parse --short=7 HEAD)
BRANCH  = $(shell git rev-parse --abbrev-ref HEAD)

# Symlink into GOPATH
GITHUB_PATH = github.com:jfigge
CURRENT_DIR = $(shell pwd)
BUILD_DIR_LINK = $(shell readlink ${BUILD_DIR})

export PATH := ${BIN}:$(PATH)
export GOPRIVATE:=*.teradata.com

# Setup the -ldflags option for go build here, interpolate the variable values
FLAGS_PKG=us.figge.auto-ssh/internal/core/config
LDFLAGS = --ldflags "-X ${FLAGS_PKG}.Version=${VERSION} -X ${FLAGS_PKG}.Commit=${COMMIT} -X ${FLAGS_PKG}.Branch=${BRANCH} -X ${FLAGS_PKG}.BuildNumber=${BUILD_NUMBER} -X ${FLAGS_PKG}.Release=${BUILD_RELEASE}"

PKGS= \

all: clean lint test darwin
	

os: all linux windows
	GOARCH=amd64; make darwin linux windows

lint:
	golangci-lint run

fmt:
	go fmt $$(go list ./... | grep -v /internal_vendor/);
	go fmt $$(go list ./... | grep -v /vendor/);
	goimports -local github.com/golangci/golangci-lint -w $$(find . -type f -iname \*.go)

vet:
	go vet $$(go list ./... | grep -v /internal_vendor/);
	go vet $$(go list ./... | grep -v /vendor/)

test: 
	go clean -testcache
	go test -v ./...

test-with-coverage:
	mkdir -p coverage
	go test ./... -coverpkg=./... -covermode=count -coverprofile coverage/coverage.txt
	go tool cover -func=coverage/coverage.txt -o coverage/profile.out
	echo `tail -1 coverage/profile.out`
	gocover-cobertura < coverage/coverage.txt > coverage/cobertura-coverage.xml

linux:
	GOOS=linux GOARCH=${GOARCH} go build ${LDFLAGS} -o ${BINARY}-linux-${GOARCH} ${TARGET};

darwin: OS="Darwin amd64"
darwin:
	GOOS=darwin GOARCH=${GOARCH} go build ${LDFLAGS} -o ${BINARY}-darwin-${GOARCH} ${TARGET};

windows:
	GOOS=windows GOARCH=${GOARCH} go build -buildmode=exe ${LDFLAGS} -o ${BINARY}-windows-${GOARCH}.exe ${TARGET};

clean:
	-rm -rf test
	-rm -rf ash-*


.phony: all os lint fmt vet test test-with-coverage linux darwin windows clean