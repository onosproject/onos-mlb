// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	"context"
	"time"

	"github.com/atomix/go-client/pkg/client/errors"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	"github.com/onosproject/onos-mlb/pkg/monitor"
	"github.com/onosproject/onos-mlb/pkg/southbound/e2control"
	ocnstorage "github.com/onosproject/onos-mlb/pkg/store/ocn"
	paramstorage "github.com/onosproject/onos-mlb/pkg/store/parameters"
	"github.com/onosproject/onos-mlb/pkg/store/storage"
	meastype "github.com/onosproject/rrm-son-lib/pkg/model/measurement/type"
)

var log = logging.GetLogger()

const (
	// RcPreRanParamDefaultOCN is default Ocn value
	RcPreRanParamDefaultOCN = meastype.QOffset0dB
)

// NewHandler generates new MLB controller handler
func NewHandler(e2controlHandler e2control.Handler,
	monitorHandler monitor.Handler,
	numUEsMeasStore storage.Store,
	neighborMeasStore storage.Store,
	ocnStore ocnstorage.Store,
	paramStore paramstorage.Store) Handler {
	return &handler{
		e2controlHandler:  e2controlHandler,
		monitorHandler:    monitorHandler,
		numUEsMeasStore:   numUEsMeasStore,
		neighborMeasStore: neighborMeasStore,
		ocnStore:          ocnStore,
		paramStore:        paramStore,
	}
}

// Handler is an interface including MLB controller
type Handler interface {
	// Run runs MLB controller
	Run(ctx context.Context) error
}

type handler struct {
	e2controlHandler  e2control.Handler
	monitorHandler    monitor.Handler
	numUEsMeasStore   storage.Store
	neighborMeasStore storage.Store
	ocnStore          ocnstorage.Store
	paramStore        paramstorage.Store
}

func (h *handler) Run(ctx context.Context) error {
	for {
		interval, err := h.paramStore.Get(context.Background(), "interval")
		if err != nil {
			log.Error(err)
			continue
		}
		select {
		case <-time.After(time.Duration(interval) * time.Second):
			// ToDo should run as goroutine
			h.startControlLogic(ctx)
		case <-ctx.Done():
			return nil
		}
	}
}

func (h *handler) startControlLogic(ctx context.Context) {
	// run monitor handler
	err := h.monitorHandler.Monitor(ctx)
	if err != nil {
		if err.Error() == monitor.WarnMsgRNIBEmpty {
			log.Warnf(err.Error())
			return
		}
		log.Error(err)
		return
	}

	// update ocn store - to update neighbor or to add new cells coming
	err = h.updateOcnStore(ctx)
	if err != nil {
		log.Error(err)
		return
	}

	// Get total num UE
	totalNumUEs, err := h.getTotalNumUEs(ctx)
	if err != nil {
		log.Error(err)
		return
	}

	// Get Cell IDs
	cells, err := h.getCellList(ctx)
	if err != nil {
		log.Error(err)
		return
	}

	// run control logic for each cell
	for _, cell := range cells {
		err = h.controlLogicEachCell(ctx, cell, cells, totalNumUEs)
		if err != nil {
			log.Error(err)
			return
		}
	}
}

func (h *handler) updateOcnStore(ctx context.Context) error {
	ch := make(chan *storage.Entry)
	go func(ch chan *storage.Entry) {
		err := h.neighborMeasStore.ListElements(ctx, ch)
		if err != nil {
			log.Error(err)
			close(ch)
		}
	}(ch)

	for e := range ch {
		ids := e.Key
		neighborList := e.Value.([]storage.IDs)

		if _, err := h.ocnStore.Get(ctx, ids); err != nil {
			// the new cells connected
			_, err = h.ocnStore.Put(ctx, ids, &ocnstorage.OcnMap{
				Value: make(map[storage.IDs]meastype.QOffsetRange),
			})
			if err != nil {
				close(ch)
				return err
			}
			for _, nIDs := range neighborList {
				err = h.ocnStore.PutInnerMap(ctx, ids, nIDs, RcPreRanParamDefaultOCN)
				if err != nil {
					close(ch)
					return err
				}
			}
		} else {
			// delete removed neighbor
			inCh := make(chan ocnstorage.InnerEntry)
			go func(inCh chan ocnstorage.InnerEntry) {
				err := h.ocnStore.ListInnerElement(ctx, ids, inCh)
				if err != nil {
					log.Error(err)
					close(ch)
					close(inCh)
				}
			}(inCh)

			for k := range inCh {
				if !h.containsIDs(k.Key, neighborList) {
					err = h.ocnStore.DeleteInnerElement(ctx, ids, k.Key)
					close(ch)
					return err
				}
			}

			// add new neighbor
			for _, n := range neighborList {
				if _, err = h.ocnStore.GetInnerMap(ctx, ids, n); err != nil {
					err = h.ocnStore.PutInnerMap(ctx, ids, n, RcPreRanParamDefaultOCN)
					if err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

func (h *handler) containsIDs(ids storage.IDs, idsList []storage.IDs) bool {
	for _, e := range idsList {
		if e == ids {
			return true
		}
	}
	return false
}

func (h *handler) getTotalNumUEs(ctx context.Context) (int, error) {
	result := 0
	ch := make(chan *storage.Entry)
	go func(ch chan *storage.Entry) {
		err := h.numUEsMeasStore.ListElements(ctx, ch)
		if err != nil {
			log.Error(err)
			close(ch)
		}
	}(ch)
	for e := range ch {
		result += e.Value.(storage.Measurement).Value
	}
	return result, nil
}

func (h *handler) getCellList(ctx context.Context) ([]storage.IDs, error) {
	result := make([]storage.IDs, 0)
	ch := make(chan storage.IDs)
	go func(chan storage.IDs) {
		err := h.numUEsMeasStore.ListKeys(ctx, ch)
		if err != nil {
			log.Error(err)
			close(ch)
		}
	}(ch)
	for k := range ch {
		result = append(result, k)
	}
	return result, nil
}

func (h *handler) controlLogicEachCell(ctx context.Context, ids storage.IDs, cells []storage.IDs, totalNumUEs int) error {

	targetThreshold, err := h.paramStore.Get(context.Background(), "target_threshold")
	if err != nil {
		return err
	}
	overloadThreshold, err := h.paramStore.Get(context.Background(), "overload_threshold")
	if err != nil {
		return err
	}

	ocnDeltaFactor, err := h.paramStore.Get(context.Background(), "delta_ocn")
	if err != nil {
		return err
	}

	neighbors, err := h.neighborMeasStore.Get(ctx, ids)
	if err != nil {
		return err
	}

	// calculate for each capacity and check sCell's and its neighbors' capacity
	// if sCell load < target load threshold
	// reduce Ocn
	neighborList := neighbors.Value.([]storage.IDs)
	numUEsSCell, err := h.numUE(ctx, ids.PlmnID, ids.CellID, cells)
	if err != nil {
		return err
	}
	capSCell := h.getCapacity(1, totalNumUEs, numUEsSCell)
	log.Debugf("Serving cell (%v) capacity: %v, load: %v / neighbor: %v / overload threshold %v, target threshold %v", ids, capSCell, 100-capSCell, cells, overloadThreshold, targetThreshold)
	if 100-capSCell < targetThreshold && 100-capSCell < overloadThreshold {
		// send control message to reduce OCn for all neighbors
		for _, nCellID := range neighborList {
			ocn, err := h.ocnStore.GetInnerMap(ctx, ids, nCellID)
			if err != nil {
				return err
			}
			if ocn-meastype.QOffsetRange(ocnDeltaFactor) < meastype.QOffsetMinus24dB {
				ocn = meastype.QOffsetMinus24dB
			} else {
				ocn = ocn - meastype.QOffsetRange(ocnDeltaFactor)
			}

			err = h.e2controlHandler.SendControlMessage(ctx, nCellID, ids.NodeID, int32(ocn))
			if err != nil {
				return err
			}
			err = h.ocnStore.PutInnerMap(ctx, ids, nCellID, ocn)
			if err != nil {
				return err
			}
		}
		return nil
	}

	// if sCell load > overload threshold && nCell < target load threshold
	// increase Ocn
	if 100-capSCell > overloadThreshold {
		for _, nCellID := range neighborList {
			numUEsNCell, err := h.numUE(ctx, nCellID.PlmnID, nCellID.CellID, cells)
			if err != nil {
				log.Warnf("there is no num(UEs) measurement value; this neighbor (plmnid-%v:cid-%v) may not be controlled by this xAPP; set num(UEs) to 0", nCellID.PlmnID, nCellID.CellID)
			}
			capNCell := h.getCapacity(1, totalNumUEs, numUEsNCell)
			log.Debugf("Serving cell (%v)'s neighbor cell (%v) capacity: %v, load: %v / overload threshold %v, target threshold %v", ids, nCellID, capNCell, 100-capNCell, overloadThreshold, targetThreshold)
			if 100-capNCell < targetThreshold {
				ocn, err := h.ocnStore.GetInnerMap(ctx, ids, nCellID)
				if err != nil {
					return err
				}
				if ocn+meastype.QOffsetRange(ocnDeltaFactor) > meastype.QOffset24dB {
					ocn = meastype.QOffset24dB
				} else {
					ocn = ocn + meastype.QOffsetRange(ocnDeltaFactor)
				}
				err = h.e2controlHandler.SendControlMessage(ctx, nCellID, ids.NodeID, int32(ocn))
				if err != nil {
					return err
				}
				err = h.ocnStore.PutInnerMap(ctx, ids, nCellID, ocn)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (h *handler) getCapacity(denominationFactor float64, totalNumUEs int, numUEs int) int {
	capacity := (1 - float64(numUEs)/(denominationFactor*float64(totalNumUEs))) * 100
	return int(capacity)
}

func (h *handler) numUE(ctx context.Context, plmnID string, cid string, cells []storage.IDs) (int, error) {
	storageID, err := h.findIDWithCGI(plmnID, cid, cells)
	if err != nil {
		return 0, err
	}

	entry, err := h.numUEsMeasStore.Get(ctx, storageID)
	if err != nil {
		return 0, err
	}
	return entry.Value.(storage.Measurement).Value, nil
}

func (h *handler) findIDWithCGI(plmnid string, cid string, cells []storage.IDs) (storage.IDs, error) {
	for _, cell := range cells {
		if cell.PlmnID == plmnid && cell.CellID == cid {
			return cell, nil
		}
	}
	return storage.IDs{}, errors.NewNotFound("ID not found with plmnid and cgi")
}
