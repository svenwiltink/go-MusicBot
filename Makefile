.PHONY: default
default: verify build ;

ME ?= github.com/svenwiltink/go-musicbot
PACKAGE_NAME ?= $(NAME)
NAME ?= MusicBot

GIT_COMMIT_HASH=`git log | head -n1 | awk '{print $$2}'`
BUILD_DATE=`date +%Y-%m-%d_%H:%M:%S`
BUILD_HOST=`hostname`
GO_VERSION=`go version | awk '{print $$3}'`
GIT_VERSION_TAG=`git describe --tags --long`

BUILD_PLATFORMS ?= -os '!netbsd' -os '!openbsd' -os '!freebsd'

all: deps verify build test

help:

deps:
	go get -u golang.org/x/lint/golint
	go get -u github.com/mitchellh/gox
	go mod vendor	

verify: fmt lint

fmt:
			  @go fmt ./... | awk '{if (NF > 0) {if (NR == 1) print "Please run go fmt for:"; print "- "$$1}} END {if (NF > 0) {if (NR > 0) exit 1}}'

lint:
			  @golint ./... | ( ! grep -v -e "^vendor/" -e "be unexported" -e "don't use an underscore in package name" -e "ALL_CAPS" )

build:
		  @mkdir -p out/binaries
	gox $(BUILD_PLATFORMS) \
		-ldflags " \
		 	-X github.com/svenwiltink/go-musicbot/pkg/bot.Version=${GIT_VERSION_TAG} \
			-X github.com/svenwiltink/go-musicbot/pkg/bot.GoVersion=${GO_VERSION} \
			-X github.com/svenwiltink/go-musicbot/pkg/bot.BuildDate=${BUILD_DATE}" \
		-output="out/binaries/$(NAME)-{{.OS}}-{{.Arch}}" github.com/svenwiltink/go-musicbot/cmd/go-musicbot
test:
	go test -v 'github.com/svenwiltink/go-musicbot/...'

deb:
	./deb-build.sh
	rm -rf pkg_root

clean:
	rm -rf out
	rm -rf pkg_root
