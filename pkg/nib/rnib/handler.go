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

const CellEntityKind = "e2cell"

var log = logging.GetLogger("rnib")

func NewHandler() (Handler, error) {
	rnibClient, err := topo.NewClient()
	if err != nil {
		return nil, err
	}
	return &handler{
		rnibClient: rnibClient,
	}, nil
}

type Handler interface {
	Get(ctx context.Context) ([]RNIBIDs, error)
	E2NodeIDs(ctx context.Context) ([]topoapi.ID, error)
	GetE2Cells(ctx context.Context, nodeID topoapi.ID) ([]topoapi.E2Cell, error)
}

type handler struct {
	rnibClient topo.Client
}

func (h *handler) Get(ctx context.Context) ([]RNIBIDs, error) {
	nodeIDs, err := h.E2NodeIDs(ctx)
	if err != nil {
		return nil, err
	}

	ids := make([]RNIBIDs, 0)
	for _, nodeID := range nodeIDs {
		e2Cells, err := h.GetE2Cells(ctx, nodeID)
		if err != nil {
			return nil, err
		}
		for _, cell := range e2Cells {
			ids = append(ids, RNIBIDs{
				NodeID: string(nodeID),
				COI: cell.CellObjectID,
				CID: cell.CellGlobalID.GetValue(),
			})
		}
	}
	return ids, nil
}

func (h *handler) E2NodeIDs(ctx context.Context) ([]topoapi.ID, error) {
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