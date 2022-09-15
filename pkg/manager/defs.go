// SPDX-FileCopyrightText: 2022-present Intel Corporation
// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package manager

const (
	// RcPreServiceModelName is RC service model name
	RcPreServiceModelName = "oran-e2sm-rc"

	// RcPreServiceModelVersion is RC service model version
	RcPreServiceModelVersion = "v1"

	// AppID is an ID of this map used in RC message
	AppID = "onos-mlb"

	// MLBAppIntervalPath is the path to get MLB controller interval
	MLBAppIntervalPath = "/controller/interval"

	// MLBAppDefaultInterval is the default value of MLB controller interval
	MLBAppDefaultInterval = 10

	// OCNDeltaFactor is the value how many inc/dec Ocn
	OCNDeltaFactor = 3
)
