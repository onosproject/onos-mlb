# SPDX-License-Identifier: Apache-2.0
# Copyright 2019 Open Networking Foundation
# Copyright 2024 Intel Corporation

export CGO_ENABLED=1
export GO111MODULE=on

.PHONY: build

ONOS_MLB_VERSION ?= latest
ONOS_BUILD_VERSION := v0.6.6
ONOS_PROTOC_VERSION := v0.6.6
BUF_VERSION := 0.27.1

all: build docker-build

build: # @HELP build the Go binaries and run all validations (default)
build:
	GOPRIVATE="github.com/onosproject/*" go build -o build/_output/onos-mlb ./cmd/onos-mlb

test: # @HELP run the unit tests and source code validation
test: build lint license
	go test -race github.com/onosproject/onos-mlb/pkg/...
	go test -race github.com/onosproject/onos-mlb/cmd/...

docker-build-onos-mlb: # @HELP build onos-mlb Docker image
	@go mod vendor
	docker build . -f build/onos-mlb/Dockerfile \
		-t onosproject/onos-mlb:${ONOS_MLB_VERSION}
	@rm -rf vendor

docker-build: # @HELP build all Docker images
docker-build: build docker-build-onos-mlb

docker-push-onos-mlb: # @HELP push onos-mlb Docker image
	docker push onosproject/onos-mlb:${ONOS_MLB_VERSION}

docker-push: # @HELP push docker images
docker-push: docker-push-onos-mlb

lint: # @HELP examines Go source code and reports coding problems
	golangci-lint --version | grep $(GOLANG_CI_VERSION) || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b `go env GOPATH`/bin $(GOLANG_CI_VERSION)
	golangci-lint run --timeout 15m

license: # @HELP run license checks
	rm -rf venv
	python3 -m venv venv
	. ./venv/bin/activate;\
	python3 -m pip install --upgrade pip;\
	python3 -m pip install reuse;\
	reuse lint

check-version: # @HELP check version is duplicated
	./build/bin/version_check.sh all

clean: # @HELP remove all the build artifacts
	rm -rf ./build/_output ./vendor ./cmd/onos-mlb/onos-mlb ./cmd/onos/onos venv
	go clean github.com/onosproject/onos-mlb/...

help:
	@grep -E '^.*: *# *@HELP' $(MAKEFILE_LIST) \
    | sort \
    | awk ' \
        BEGIN {FS = ": *# *@HELP"}; \
        {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}; \
    '
