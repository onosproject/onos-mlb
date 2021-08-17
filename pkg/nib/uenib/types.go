// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package uenib

import "github.com/onosproject/onos-api/go/onos/uenib"

// Element is an element of UENIB
type Element struct {
	Key   Key
	Value interface{}
}

// Key is the key of UENIB element
type Key struct {
	E2ID   uenib.ID
	NodeID string
	PlmnID string
	CID    string
	COI    string
	Aspect string
}
