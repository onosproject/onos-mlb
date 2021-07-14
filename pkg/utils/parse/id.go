// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package idutils

import (
	"fmt"
	"github.com/atomix/go-client/pkg/client/errors"
	"github.com/onosproject/onos-api/go/onos/uenib"
	"strings"
)

// ParseUENIBNeighborAspectKey parses neighbor aspect key in UENIB
func ParseUENIBNeighborAspectKey(key uenib.ID) (string, string, string, string, error) {
	// ToDo: PCI app should store this with hex format
	objects := strings.Split(string(key), ":")
	if len(objects) != 4 {
		return "", "", "", "", errors.NewNotSupported("neighbor aspect's key should have four key elements")
	}

	nodeID := objects[0]
	plmnID := fmt.Sprintf("%x", objects[1])
	cid := objects[2]
	ecgiType := objects[3]

	return nodeID, plmnID, cid, ecgiType, nil
}

// ParseUENIBNeighborAspectValue parses neighbor aspect value in UENIB
func ParseUENIBNeighborAspectValue(value string) (string, error) {
	// ToDo: PCI app should store this with hex format
	results := ""
	idsList := strings.Split(value, ",")
	for _, ids := range idsList {
		idList := strings.Split(ids, ":")
		plmnID := idList[0]
		cid := idList[1]
		if results == "" {
			results = fmt.Sprintf("%s:%s:%s", plmnID, cid, idList[2])
			continue
		}
		results = fmt.Sprintf("%s,%s:%s:%s", results, plmnID, cid, idList[2])
	}
	return results, nil
}

// ParseUENIBNumUEsAspectKey parses the number of UEs aspect key in UENIB
func ParseUENIBNumUEsAspectKey(key uenib.ID) (string, string, error) {
	objects := strings.Split(string(key), ":")
	if len(objects) != 2 {
		return "", "", errors.NewNotSupported("aspect's key for the number of UEs should have two key elements")
	}

	nodeID := objects[0]
	coi := objects[1]
	return nodeID, coi, nil
}
