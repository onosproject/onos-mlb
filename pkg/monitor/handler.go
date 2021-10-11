// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-Only-1.0

package monitor

import (
	"context"
	"fmt"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	"github.com/onosproject/onos-mlb/pkg/nib/rnib"
	ocnstorage "github.com/onosproject/onos-mlb/pkg/store/ocn"
	"github.com/onosproject/onos-mlb/pkg/store/storage"
)

var log = logging.GetLogger("monitor")

const (
	WarnMsgRNIBEmpty = "R-NIB is empty"
)

// NewHandler generates monitoring handler
func NewHandler(rnibHandler rnib.Handler, numUEsMeasStore storage.Store, neighborMeasStore storage.Store, ocnStore ocnstorage.Store) Handler {
	return &handler{
		rnibHandler:       rnibHandler,
		numUEsMeasStore:   numUEsMeasStore,
		neighborMeasStore: neighborMeasStore,
		ocnStore:          ocnStore,
	}
}

// Handler is an interface including this handler's functions
type Handler interface {
	// Monitor starts to monitor UENIB and RNIB
	Monitor(ctx context.Context) error
}

type handler struct {
	rnibHandler       rnib.Handler
	numUEsMeasStore   storage.Store
	neighborMeasStore storage.Store
	ocnStore          ocnstorage.Store
}

func (h *handler) Monitor(ctx context.Context) error {
	// get RNIB
	rnibList, err := h.rnibHandler.Get(ctx)
	if err != nil {
		return err
	} else if len(rnibList) == 0 {
		return fmt.Errorf(WarnMsgRNIBEmpty)
	}

	// fill PLMN IDs in each element key since topo key does not have PLMN ID
	h.fillPlmnID(rnibList)

	// store monitoring result
	h.storeRNIB(ctx, rnibList)

	log.Debugf("RNIB List %v", rnibList)

	return nil
}

func (h *handler) fillPlmnID(rnibList []rnib.Element) {
	mapPlmnID := make(map[string]string)
	for _, e := range rnibList {
		if e.Key.Aspect == rnib.Neighbors {
			for _, id := range e.Value.([]rnib.CellGlobalID) {
				mapPlmnID[id.CellIdentity] = id.PlmnID
			}
		}
	}
	for i := 0; i < len(rnibList); i++ {
		rnibList[i].Key.IDs.CellGlobalID.PlmnID = mapPlmnID[rnibList[i].Key.IDs.CellGlobalID.CellIdentity]
	}
}

func (h *handler) storeRNIB(ctx context.Context, rnibList []rnib.Element) {
	for _, e := range rnibList {
		key := storage.IDs{
			NodeID:    e.Key.IDs.E2NodeID,
			PlmnID:    e.Key.IDs.CellGlobalID.PlmnID,
			CellID:    e.Key.IDs.CellGlobalID.CellIdentity,
			CellObjID: e.Key.IDs.CellObjectID,
		}
		switch e.Key.Aspect {
		case rnib.Neighbors:
			err := h.storeRNIBNeighbors(ctx, key, e.Value.([]rnib.CellGlobalID))
			if err != nil {
				log.Error(err)
			}
		case rnib.NumUEs:
			err := h.storeRNIBNumUEs(ctx, key, e.Value.(uint32))
			if err != nil {
				log.Error(err)
			}
		default:
			log.Warnf("Unavailable aspects for this app - to be discarded: %v", e.Key.Aspect.String())
		}
	}
}

func (h *handler) storeRNIBNeighbors(ctx context.Context, key storage.IDs, neighborIDs []rnib.CellGlobalID) error {
	nidList := make([]storage.IDs, 0)
	for _, id := range neighborIDs {
		nid := storage.IDs{
			PlmnID: id.PlmnID,
			CellID: id.CellIdentity,
		}
		nidList = append(nidList, nid)
	}
	_, err := h.neighborMeasStore.Put(ctx, key, nidList)
	return err
}

func (h *handler) storeRNIBNumUEs(ctx context.Context, key storage.IDs, value uint32) error {
	measurement := storage.Measurement{
		Value: int(value),
	}
	_, err := h.numUEsMeasStore.Put(ctx, key, measurement)
	return err
}
