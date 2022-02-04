// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/onosproject/helmit/pkg/registry"
	"github.com/onosproject/helmit/pkg/test"
	"github.com/onosproject/onos-mlb/test/overload"
	"github.com/onosproject/onos-mlb/test/targetload"
	"github.com/onosproject/onos-mlb/test/underload"
)

func main() {
	registry.RegisterTestSuite("overload", &overload.TestSuite{})
	registry.RegisterTestSuite("targetload", &targetload.TestSuite{})
	registry.RegisterTestSuite("underload", &underload.TestSuite{})
	test.Main()
}
