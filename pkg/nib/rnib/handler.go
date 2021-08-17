// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package rnib

import (
	"context"
	topoapi "github.com/onosproject/onos-api/go/onos/topo"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	"github.com/onosproject/onos-ric-sdk-go/pkg/topo"
)

var log = logging.GetLogger("rnib")

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
	Get(ctx context.Context) ([]IDs, error)

	// GetE2NodeIDs gets all E2 Node IDs
	GetE2NodeIDs(ctx context.Context) ([]topoapi.ID, error)

	// GetE2Cells gets all cells managed by all E2 nodes
	GetE2Cells(ctx context.Context, nodeID topoapi.ID) ([]topoapi.E2Cell, error)
}

type handler struct {
	rnibClient topo.Client
}

func (h *handler) Get(ctx context.Context) ([]IDs, error) {
	e2NodeIDs, err := h.GetE2NodeIDs(ctx)
	log.Debugf("e2NodeIDs: %v", e2NodeIDs)
	if err != nil {
		return nil, err
	}

	ids := make([]IDs, 0)
	for _, e2NodeID := range e2NodeIDs {
		e2Cells, err := h.GetE2Cells(ctx, e2NodeID)
		if err != nil {
			return nil, err
		}
		for _, cell := range e2Cells {
			log.Debugf("nodeID: %v", e2NodeID)
			ids = append(ids, IDs{
				NodeID: string(e2NodeID),
				COI:    cell.CellObjectID,
				CID:    cell.CellGlobalID.GetValue(),
			})
		}
	}
	log.Debugf("Received RNIB: %v", ids)
	return ids, nil
}

func (h *handler) GetE2NodeIDs(ctx context.Context) ([]topoapi.ID, error) {
	objects, err := h.rnibClient.List(ctx, topo.WithListFilters(getControlRelationFilter()))
	if err != nil {
		return nil, err
	}

	e2NodeIDs := make([]topoapi.ID, len(objects))
	for _, object := range objects {
		relation := object.Obj.(*topoapi.Object_Relation)
		e2NodeID := relation.Relation.TgtEntityID
		e2NodeIDs = append(e2NodeIDs, e2NodeID)
	}

	return e2NodeIDs, nil
}

func (h *handler) GetE2Cells(ctx context.Context, nodeID topoapi.ID) ([]topoapi.E2Cell, error) {
	filter := &topoapi.Filters{
		RelationFilter: &topoapi.RelationFilter{SrcId: string(nodeID),
			RelationKind: topoapi.CONTAINS,
			TargetKind:   ""}}

	objects, err := h.rnibClient.List(ctx, topo.WithListFilters(filter))
	if err != nil {
		return nil, err
	}
	var e2Cells []topoapi.E2Cell
	for _, obj := range objects {
		targetEntity := obj.GetEntity()
		if targetEntity.GetKindID() == topoapi.E2CELL {
			cellObject := topoapi.E2Cell{}
			obj.GetAspect(&cellObject)
			e2Cells = append(e2Cells, cellObject)
		}
	}

	return e2Cells, nil
}

func getControlRelationFilter() *topoapi.Filters {
	controlRelationFilter := &topoapi.Filters{
		KindFilter: &topoapi.Filter{
			Filter: &topoapi.Filter_Equal_{
				Equal_: &topoapi.EqualFilter{
					Value: topoapi.CONTROLS,
				},
			},
		},
	}
	return controlRelationFilter
}
