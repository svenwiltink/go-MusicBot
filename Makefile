NAME ?= MusicBot
PACKAGE_NAME ?= $(NAME)

GIT_COMMIT_HASH=`git log | head -n1 | awk '{print $$2}'`
BUILD_DATE=`date +%Y-%m-%d_%H:%M:%S`
BUILD_HOST=`hostname`
GO_VERSION=`go version | awk '{print $$3}'`
GIT_VERSION_TAG=`git describe --tags --long`
LDFLAGS="-X util.GitCommit=${GIT_COMMIT_HASH} -X util.BuildHost=${BUILD_HOST} -X util.BuildDate=${BUILD_DATE} -X util.GoVersion=${GO_VERSION} -X util.VersionTag=${GIT_VERSION_TAG}"

BUILD_PLATFORMS ?= -os '!netbsd' -os '!openbsd' -os '!freebsd' -os '!windows'

OUR_PACKAGES=$(shell go list ./... | grep -v '/vendor/')

all: deps verify build

help:

deps:
	go get -u github.com/golang/lint/golint
	go get github.com/mitchellh/gox
	go get -u github.com/Masterminds/glide

verify: fmt lint

fmt:
	@go fmt $(OUR_PACKAGES) | awk '{if (NF > 0) {if (NR == 1) print "Please run go fmt for:"; print "- "$$1}} END {if (NF > 0) {if (NR > 0) exit 1}}'

lint:
	@golint ./... | ( ! grep -v -e "^vendor/" -e "be unexported" -e "don't use an underscore in package name" -e "ALL_CAPS" )

build:
	gox $(BUILD_PLATFORMS) \
	        -ldflags="${LDFLAGS}" \
            -output="out/binaries/$(NAME)-{{.OS}}-{{.Arch}}"
