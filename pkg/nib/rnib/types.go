// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-Only-1.0

package rnib

import "github.com/onosproject/onos-api/go/onos/topo"

// Element is an element of R-NIB
type Element struct {
	Key   Key
	Value interface{}
}

// Key is a key of R-NIB
type Key struct {
	IDs    IDs
	Aspect AspectType
}

// IDs is an ID of R-NIB
type IDs struct {
	TopoID       topo.ID
	E2NodeID     string
	CellObjectID string
	CellGlobalID CellGlobalID
}

// CellGlobalID is an ID for a cell
type CellGlobalID struct {
	CellIdentity string
	PlmnID       string
}

type AspectType int

const (
	Neighbors = iota
	NumUEs
)

func (a AspectType) String() string {
	return [...]string{"Neighbors", "NumUEs"}[a]
}
