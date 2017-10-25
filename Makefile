NAME ?= MusicBot
PACKAGE_NAME ?= $(NAME)

GIT_COMMIT_HASH=`git log | head -n1 | awk '{print $$2}'`
BUILD_DATE=`date +%Y-%m-%d_%H:%M:%S`
BUILD_HOST=`hostname`
GO_VERSION=`go version | awk '{print $$3}'`
GIT_VERSION_TAG=`git describe --tags --long`
PACKAGE_VERSION=`git describe --tags --long | cut -c 2-`
LDFLAGS=-X github.com/svenwiltink/go-musicbot/util.GitCommit=${GIT_COMMIT_HASH} \
        -X github.com/svenwiltink/go-musicbot/util.BuildHost=${BUILD_HOST} \
        -X github.com/svenwiltink/go-musicbot/util.BuildDate=${BUILD_DATE} \
        -X github.com/svenwiltink/go-musicbot/util.GoVersion=${GO_VERSION} \
        -X github.com/svenwiltink/go-musicbot/util.VersionTag=${GIT_VERSION_TAG}

BUILD_PLATFORMS ?= -os '!netbsd' -os '!openbsd' -os '!freebsd' -os '!windows'

OUR_PACKAGES=$(shell go list ./... | grep -v '/vendor/')

all: deps verify build

help:

deps:
	go get -u github.com/golang/lint/golint
	go get github.com/mitchellh/gox
	go get -u github.com/Masterminds/glide
	glide install

verify: fmt lint

fmt:
	@go fmt $(OUR_PACKAGES) | awk '{if (NF > 0) {if (NR == 1) print "Please run go fmt for:"; print "- "$$1}} END {if (NF > 0) {if (NR > 0) exit 1}}'

lint:
	@golint ./... | ( ! grep -v -e "^vendor/" -e "be unexported" -e "don't use an underscore in package name" -e "ALL_CAPS" )

build:
	gox $(BUILD_PLATFORMS) \
	        -ldflags="${LDFLAGS}" \
            -output="out/binaries/$(NAME)-{{.OS}}-{{.Arch}}"

deb:
	mkdir -p out/deb
	cp out/binaries/MusicBot-linux-amd64 out/deb/go-musicbot
	fpm -f \
		-s dir -t deb \
		-n go-musicbot \
		-v "${PACKAGE_VERSION}" \
		 --config-files /etc/go-musicbot/conf.json out/deb/=/usr/bin/ conf.json.example=/etc/go-musicbot/conf.json \
