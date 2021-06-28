// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package storage

import "github.com/onosproject/rrm-son-lib/pkg/model/measurement/type"

type IDs struct {
	NodeID string
	PlmnID string
	CellID string
	CellObjID string
}

type Entry struct {
	Key IDs
	Value interface{}
}

type storageEvent int

const (
	None storageEvent = iota
	Created
	Updated
	Deleted
)

func (e storageEvent) String() string {
	return [...]string{"None", "Created", "Updated", "Deleted"}[e]
}

type Measurement struct {
	Value int
}

type OcnMap struct {
	Value map[IDs]meastype.QOffsetRange
}

type Statistics struct {
	Value map[IDs]int
}