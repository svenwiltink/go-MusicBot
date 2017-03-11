NAME ?= MusicBot
PACKAGE_NAME ?= $(NAME)

BUILD_PLATFORMS ?= -os '!netbsd' -os '!openbsd' -os '!windows'

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
            -output="out/binaries/$(NAME)-{{.OS}}-{{.Arch}}"
