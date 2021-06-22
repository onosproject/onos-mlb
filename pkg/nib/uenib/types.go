// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package uenib

type UENIBElement struct {
	Key UENIBKey
	Value interface{}
}

type UENIBKey struct {
	NodeID string
	PlmnID string
	CID string
	COI string
	Aspect string
}
