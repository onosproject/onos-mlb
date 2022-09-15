// SPDX-FileCopyrightText: 2022-present Intel Corporation
// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package rnib

import (
	"context"
	"fmt"

	topoapi "github.com/onosproject/onos-api/go/onos/topo"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	idutils "github.com/onosproject/onos-mlb/pkg/utils/parse"
	"github.com/onosproject/onos-ric-sdk-go/pkg/topo"
)

var log = logging.GetLogger()

const (
	// AspectKeyNumUEsRANSim is the R-NIB aspect key of the number of UEs for RAN-Simulator
	AspectKeyNumUEsRANSim = "RRC.Conn.Avg"

	// AspectKeyNumUEsOAI is the R-NIB aspect key of the number of UEs for OAI
	AspectKeyNumUEsOAI = "RRC.ConnMean"
)

// NewHandler generates the new RNIB handler
func NewHandler() (Handler, error) {
	rnibClient, err := topo.NewClient()
	if err != nil {
		return nil, err
	}
	return &handler{
		rnibClient: rnibClient,
	}, nil
}

// Handler includes RNIB handler's all functions
type Handler interface {
	// Get gets all RNIB
	Get(ctx context.Context) ([]Element, error)
	GetE2NodeAspects(ctx context.Context, nodeID topoapi.ID) (*topoapi.E2Node, error)
}

type handler struct {
	rnibClient topo.Client
}

func (h *handler) GetE2NodeAspects(ctx context.Context, nodeID topoapi.ID) (*topoapi.E2Node, error) {
	object, err := h.rnibClient.Get(ctx, nodeID)
	if err != nil {
		return nil, err
	}
	e2Node := &topoapi.E2Node{}
	err = object.GetAspect(e2Node)

	return e2Node, err
}

func (h *handler) Get(ctx context.Context) ([]Element, error) {
	objects, err := h.rnibClient.List(ctx)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	result := make([]Element, 0)

	log.Debugf("R-NIB objects - %s", objects)
	for _, obj := range objects {
		if obj.GetEntity() == nil || obj.GetEntity().GetKindID() != topoapi.E2CELL {
			continue
		}
		log.Debugf("R-NIB each obj: %s", obj)
		cellTopoID := obj.GetID()
		e2NodeID, cellIdentity := idutils.ParseCellTopoID(string(cellTopoID))
		cellObject := topoapi.E2Cell{}
		err = obj.GetAspect(&cellObject)
		if err != nil {
			return nil, err
		}

		cellObjectID := cellObject.CellObjectID
		if cellIdentity != cellObject.CellGlobalID.GetValue() {
			return nil, fmt.Errorf("verification failed: In R-NIB, cell IDs in topo ID field and aspects are different")
		}
		// ToDo: add PLMN ID here for cell object in the future
		plmnID := ""

		if cellObjectID == "" || cellIdentity == "" {
			return nil, fmt.Errorf("R-NIB is not ready yet")
		}

		ids := IDs{
			TopoID:       cellTopoID,
			E2NodeID:     e2NodeID,
			CellObjectID: cellObjectID,
			CellGlobalID: CellGlobalID{
				CellIdentity: cellIdentity,
				PlmnID:       plmnID,
			},
		}

		if len(cellObject.NeighborCellIDs) == 0 || len(cellObject.KpiReports) == 0 {
			continue
		}

		neighbors := make([]CellGlobalID, 0)
		for _, neighborCellID := range cellObject.NeighborCellIDs {
			neighborCellGlobalID := CellGlobalID{
				CellIdentity: neighborCellID.CellGlobalID.GetValue(),
				PlmnID:       neighborCellID.PlmnID,
			}
			neighbors = append(neighbors, neighborCellGlobalID)
			plmnID = neighborCellID.PlmnID
		}
		ids.CellGlobalID.PlmnID = plmnID
		neighborElement := Element{
			Key: Key{
				IDs:    ids,
				Aspect: Neighbors,
			},
			Value: neighbors,
		}
		result = append(result, neighborElement)

		for kpiKey, kpiValue := range cellObject.KpiReports {
			if kpiKey == AspectKeyNumUEsOAI || kpiKey == AspectKeyNumUEsRANSim {
				kpiElement := Element{
					Key: Key{
						IDs:    ids,
						Aspect: NumUEs,
					},
					Value: kpiValue,
				}
				result = append(result, kpiElement)
				break
			}
		}
	}

	return result, nil
}
