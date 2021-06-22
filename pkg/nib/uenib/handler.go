// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package uenib

import (
	"context"
	"github.com/atomix/go-client/pkg/client/errors"
	"github.com/onosproject/onos-api/go/onos/uenib"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	"github.com/onosproject/onos-mlb/pkg/utils/conn"
	idutils "github.com/onosproject/onos-mlb/pkg/utils/parse"
	"google.golang.org/grpc"
	"io"
)

const (
	AspectKeyNeighbors = "neighbors"
	AspectKeyNumUEsRANSim = "RRC.Conn.Avg"
	AspectKeyNumUEsOAI = "RRC.ConnMean"
)

var log = logging.GetLogger("uenib")

func NewHandler(uenibAddr string, certPath string, keyPath string) (Handler, error) {
	dialOpt, err := grpcutils.NewDialOptForRetry(certPath, keyPath)
	if err != nil {
		return nil, err
	}
	conn, err := grpc.Dial(uenibAddr, dialOpt...)
	if err != nil {
		return nil, err
	}
	return &handler{
		uenibClient: uenib.NewUEServiceClient(conn),
	}, nil
}

type Handler interface {
	Get(ctx context.Context) ([]UENIBElement, error)
}

type handler struct {
	uenibClient  uenib.UEServiceClient
}

func (h *handler) Get(ctx context.Context) ([]UENIBElement, error) {
	req := &uenib.ListUERequest{
		AspectTypes: []string{AspectKeyNeighbors, AspectKeyNumUEsRANSim, AspectKeyNumUEsOAI},
	}
	resp, err := h.uenibClient.ListUEs(ctx, req)
	if err != nil {
		return nil, err
	}

	results := make([]UENIBElement, 0)
	for {
		object, err := resp.Recv()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		uenib := object.GetUE()
		aspects := uenib.GetAspects()
		uenibID := uenib.GetID()
		for k, v := range aspects {
			uenibKey, err := h.createUENIBKey(uenibID, k)
			if err != nil {
				return nil, err
			}
			results = append(results, UENIBElement{
				Key: uenibKey,
				Value: string(v.Value),
			})
		}
	}

	return results, nil
}

func (h *handler) createUENIBKey(uenibID uenib.ID, aspect string) (UENIBKey, error) {
	switch aspect {
	case AspectKeyNeighbors:
		// it has nodeid, plmnid, cid, and cgi type
		nodeID, plmnID, cid, _, err := idutils.ParseUENIBNeighborAspectKey(uenibID)
		if err != nil {
			return UENIBKey{}, err
		}
		return UENIBKey{
			NodeID: nodeID,
			PlmnID: plmnID,
			CID: cid,
			Aspect: aspect,
		}, nil
	case AspectKeyNumUEsRANSim, AspectKeyNumUEsOAI:
		// it has nodeid and coi
		nodeID, coi, err := idutils.ParseUENIBNumUEsAspectKey(uenibID)
		if err != nil {
			return UENIBKey{}, err
		}
		return UENIBKey{
			NodeID: nodeID,
			COI: coi,
			Aspect: aspect,
		}, nil
	default:
		return UENIBKey{}, errors.NewNotSupported("unavailable aspects for this app")
	}
}

