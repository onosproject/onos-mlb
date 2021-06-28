// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package storage

import "github.com/onosproject/rrm-son-lib/pkg/model/measurement/type"

// IDs is a key of this store element
type IDs struct {
	NodeID    string
	PlmnID    string
	CellID    string
	CellObjID string
}

// Entry is an entry of this store element
type Entry struct {
	Key   IDs
	Value interface{}
}

type storageEvent int

const (
	// None means the event not happened or unknown event happened
	None storageEvent = iota

	// Created means that a store element is saved
	Created

	// Updated means that a store element is updated
	Updated

	// Deleted means that a store element is deleted
	Deleted
)

// String returns string value of storageEvent enum value
func (e storageEvent) String() string {
	return [...]string{"None", "Created", "Updated", "Deleted"}[e]
}

// Measurement is the struct to store measurement results
type Measurement struct {
	Value int
}

// OcnMap is the struct to store Ocn values
type OcnMap struct {
	Value map[IDs]meastype.QOffsetRange
}

// Statistics is the struct to store statistics
type Statistics struct {
	Value map[IDs]int
}
