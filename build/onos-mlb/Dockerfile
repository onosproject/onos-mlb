# SPDX-FileCopyrightText: 2019-present Open Networking Foundation <info@opennetworking.org>
#
# SPDX-License-Identifier: Apache-2.0

ARG ONOS_BUILD_VERSION=undefined

FROM onosproject/golang-build:v1.3.0 as build

ENV GO111MODULE=on
ARG ONOS_MAKE_TARGET=build

COPY Makefile go.mod go.sum /go/src/github.com/onosproject/onos-mlb/
COPY cmd/ /go/src/github.com/onosproject/onos-mlb/cmd/
COPY pkg/ /go/src/github.com/onosproject/onos-mlb/pkg/
COPY vendor/ /go/src/github.com/onosproject/onos-mlb/vendor/
COPY api/ /go/src/github.com/onosproject/onos-mlb/api
COPY build/ /go/src/github.com/onosproject/onos-mlb/build/

RUN cd /go/src/github.com/onosproject/onos-mlb && GOFLAGS=-mod=vendor make ${ONOS_MAKE_TARGET}

FROM frolvlad/alpine-glibc:alpine-3.20

USER nobody

COPY --from=build /go/src/github.com/onosproject/onos-mlb/build/_output/onos-mlb /usr/local/bin/onos-mlb

ENTRYPOINT ["onos-mlb"]
