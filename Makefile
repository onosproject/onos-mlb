# SPDX-FileCopyrightText: 2019-present Open Networking Foundation <info@opennetworking.org>
#
# SPDX-License-Identifier: Apache-2.0

export CGO_ENABLED=1
export GO111MODULE=on

.PHONY: build

ONOS_MLB_VERSION := latest
ONOS_BUILD_VERSION := v0.6.6
ONOS_PROTOC_VERSION := v0.6.6
BUF_VERSION := 0.27.1

build: # @HELP build the Go binaries and run all validations (default)
build:
	GOPRIVATE="github.com/onosproject/*" go build -o build/_output/onos-mlb ./cmd/onos-mlb

build-tools:=$(shell if [ ! -d "./build/build-tools" ]; then cd build && git clone https://github.com/onosproject/build-tools.git; fi)
include ./build/build-tools/make/onf-common.mk

test: # @HELP run the unit tests and source code validation
test: build deps linters license_check_apache
	go test -race github.com/onosproject/onos-mlb/pkg/...
	go test -race github.com/onosproject/onos-mlb/cmd/...

jenkins-test:  # @HELP run the unit tests and source code validation producing a junit style report for Jenkins
jenkins-test: deps license_check_apache linters
	TEST_PACKAGES=github.com/onosproject/onos-mlb/... ./build/build-tools/build/jenkins/make-unit

buflint: #@HELP run the "buf check lint" command on the proto files in 'api'
	docker run -it -v `pwd`:/go/src/github.com/onosproject/onos-mlb \
		-w /go/src/github.com/onosproject/onos-mlb/api \
		bufbuild/buf:${BUF_VERSION} check lint

protos: # @HELP compile the protobuf files (using protoc-go Docker)
protos:
	docker run -it -v `pwd`:/go/src/github.com/onosproject/onos-mlb \
		-w /go/src/github.com/onosproject/onos-mlb \
		--entrypoint build/bin/compile-protos.sh \
		onosproject/protoc-go:${ONOS_PROTOC_VERSION}

helmit-overload: integration-test-namespace # @HELP run overload tests locally
	helmit test -n test ./cmd/onos-mlb-test --timeout 30m --no-teardown \
			--secret sd-ran-username=${repo_user} --secret sd-ran-password=${repo_password} \
			--suite overload

helmit-underload: integration-test-namespace # @HELP run underload tests locally
	helmit test -n test ./cmd/onos-mlb-test --timeout 30m --no-teardown \
			--secret sd-ran-username=${repo_user} --secret sd-ran-password=${repo_password} \
			--suite underload

helmit-targetload: integration-test-namespace # @HELP run underload tests locally
	helmit test -n test ./cmd/onos-mlb-test --timeout 30m --no-teardown \
			--secret sd-ran-username=${repo_user} --secret sd-ran-password=${repo_password} \
			--suite targetload

integration-tests: helmit-overload helmit-underload helmit-targetload

onos-mlb-docker: # @HELP build onos-mlb Docker image
onos-mlb-docker:
	@go mod vendor
	docker build . -f build/onos-mlb/Dockerfile \
		-t onosproject/onos-mlb:${ONOS_MLB_VERSION}
	@rm -rf vendor

images: # @HELP build all Docker images
images: build onos-mlb-docker

kind: # @HELP build Docker images and add them to the currently configured kind cluster
kind: images
	@if [ "`kind get clusters`" = '' ]; then echo "no kind cluster found" && exit 1; fi
	kind load docker-image onosproject/onos-mlb:${ONOS_MLB_VERSION}

all: build images

publish: # @HELP publish version on github and dockerhub
	./build/build-tools/publish-version ${VERSION} onosproject/onos-mlb

jenkins-publish: jenkins-tools # @HELP Jenkins calls this to publish artifacts
	./build/bin/push-images
	./build/build-tools/release-merge-commit

clean:: # @HELP remove all the build artifacts
	rm -rf ./build/_output ./vendor ./cmd/onos-mlb/onos-mlb ./cmd/onos/onos
	go clean -testcache github.com/onosproject/onos-mlb/...

