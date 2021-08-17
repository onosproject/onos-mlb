// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package monitor

import (
	"context"
	"github.com/atomix/go-client/pkg/client/errors"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	"github.com/onosproject/onos-mlb/pkg/nib/rnib"
	"github.com/onosproject/onos-mlb/pkg/nib/uenib"
	ocnstorage "github.com/onosproject/onos-mlb/pkg/store/ocn"
	"github.com/onosproject/onos-mlb/pkg/store/storage"
	"strconv"
	"strings"
)

var log = logging.GetLogger("monitor")

// NewHandler generates monitoring handler
func NewHandler(rnibHandler rnib.Handler, uenibHandler uenib.Handler, numUEsMeasStore storage.Store, neighborMeasStore storage.Store, ocnStore ocnstorage.Store) Handler {
	return &handler{
		rnibHandler:       rnibHandler,
		uenibHandler:      uenibHandler,
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
	uenibHandler      uenib.Handler
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
		return errors.NewNotFound("rnib list is empty")
	}

	// get UENIB
	uenibList, err := h.uenibHandler.Get(ctx)
	if err != nil {
		return err
	} else if len(uenibList) == 0 {
		return errors.NewNotFound("uenib element list is empty")
	}

	// verification - multiple plmn id does not support
	plmnid, err := h.plmnidVerification(uenibList)
	if err != nil {
		return err
	}

	// fill keys - coi or cgi
	// it is essential since uenib from kpimon and uenib from pci use different key
	// the uenib from kpimon uses node id and coi as a key, whereas that from pci uses node id, cell id, and plmn id as a key
	monResults, err := h.fillKeys(rnibList, uenibList, plmnid)
	if err != nil {
		return err
	}

	// store monitoring result to each key
	h.storeUENIB(ctx, monResults)

	return nil
}

func (h *handler) plmnidVerification(uenibList []uenib.Element) (string, error) {
	var plmnid string
	for _, elem := range uenibList {
		if elem.Key.Aspect == uenib.AspectKeyNeighbors {
			if plmnid == "" {
				plmnid = elem.Key.PlmnID
				continue
			} else if plmnid != elem.Key.PlmnID {
				return "", errors.NewNotSupported("this app does not support multiple plmn ids")
			}
		}
	}
	if plmnid == "" {
		return "", errors.NewNotFound("plmn id not found in uenib")
	}
	return plmnid, nil
}

func (h *handler) fillKeys(rnibList []rnib.IDs, uenibList []uenib.Element, plmnid string) ([]uenib.Element, error) {
	results := make([]uenib.Element, 0)
	for _, elem := range uenibList {
		switch elem.Key.Aspect {
		case uenib.AspectKeyNeighbors:
			coi, err := h.getCOI(string(elem.Key.E2ID), elem.Key.NodeID, elem.Key.CID, rnibList)
			if err != nil {
				return nil, err
			}
			elem.Key.COI = coi
			results = append(results, elem)
		case uenib.AspectKeyNumUEsRANSim, uenib.AspectKeyNumUEsOAI:
			elem.Key.PlmnID = plmnid
			cid, err := h.getCID(string(elem.Key.E2ID), elem.Key.NodeID, elem.Key.COI, rnibList)
			if err != nil {
				return nil, err
			}
			elem.Key.CID = cid
			results = append(results, elem)
		default:
			log.Warnf("Unavailable aspects for this app - to be discarded: %v", elem.Key.Aspect)
		}
	}
	return results, nil
}

func (h *handler) getCOI(e2id string, nodeID string, cid string, rnibList []rnib.IDs) (string, error) {
	for _, ids := range rnibList {
		if ids.E2ID == e2id && ids.CID == cid && ids.NodeID == nodeID {
			return ids.COI, nil
		}
	}
	return "", errors.NewNotFound("could not search cell object id with CID and nodeID in rnib list")
}

func (h *handler) getCID(e2id string, nodeID string, coi string, rnibList []rnib.IDs) (string, error) {
	for _, ids := range rnibList {
		if ids.E2ID == e2id && ids.COI == coi && ids.NodeID == nodeID {
			return ids.CID, nil
		}
	}
	return "", errors.NewNotFound("could not search CID with cell object ID and nodeID in rnib list")
}

func (h *handler) storeUENIB(ctx context.Context, uenibList []uenib.Element) {
	for _, u := range uenibList {
		key := storage.IDs{
			NodeID:    u.Key.NodeID,
			PlmnID:    u.Key.PlmnID,
			CellID:    u.Key.CID,
			CellObjID: u.Key.COI,
		}
		switch u.Key.Aspect {
		case uenib.AspectKeyNeighbors:
			err := h.storeUENIBNeighbors(ctx, key, u.Value.(string))
			if err != nil {
				log.Error(err)
			}
		case uenib.AspectKeyNumUEsRANSim, uenib.AspectKeyNumUEsOAI:
			err := h.storeUENIBNumUEs(ctx, key, u.Value.(string))
			if err != nil {
				log.Error(err)
			}
		default:
			log.Warnf("Unavailable aspects for this app - to be discarded: %v", u.Key.Aspect)
		}

	}
}

func (h *handler) storeUENIBNumUEs(ctx context.Context, key storage.IDs, value string) error {
	measValue, err := strconv.Atoi(value)
	if err != nil {
		return err
	}
	measurement := storage.Measurement{
		Value: measValue,
	}
	_, err = h.numUEsMeasStore.Put(ctx, key, measurement)
	return err
}

func (h *handler) storeUENIBNeighbors(ctx context.Context, key storage.IDs, value string) error {
	neighborList := strings.Split(value, ",")
	neighborIDsList := make([]storage.IDs, 0)
	for _, neighbor := range neighborList {
		nIDs := strings.Split(neighbor, ":")
		plmnID := nIDs[0]
		cid := nIDs[1]
		neighborID := storage.IDs{
			PlmnID: plmnID,
			CellID: cid,
		}
		neighborIDsList = append(neighborIDsList, neighborID)
	}
	_, err := h.neighborMeasStore.Put(ctx, key, neighborIDsList)
	return err
}
