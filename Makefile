.PHONY: default
default: verify build ;

ME ?= MusicBot
PACKAGE_NAME ?= $(NAME)

GIT_COMMIT_HASH=`git log | head -n1 | awk '{print $$2}'`
BUILD_DATE=`date +%Y-%m-%d_%H:%M:%S`
	BUILD_HOST=`hostname`
	GO_VERSION=`go version | awk '{print $$3}'`
	GIT_VERSION_TAG=`git describe --tags --long`

all: deps verify build

help:

deps:
	dep ensure	

verify: fmt lint

fmt:
	        @go fmt $(OUR_PACKAGES) | awk '{if (NF > 0) {if (NR == 1) print "Please run go fmt for:"; print "- "$$1}} END {if (NF > 0) {if (NR > 0) exit 1}}'

lint:
	        @golint ./... | ( ! grep -v -e "^vendor/" -e "be unexported" -e "don't use an underscore in package name" -e "ALL_CAPS" )

build:
		@go build -o MusicBot
