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
	// AspectKeyNeighbors is the UENIB aspect key of neighbors
	AspectKeyNeighbors = "neighbors"

	// AspectKeyNumUEsRANSim is the UENIB aspect key of the number of UEs for RAN-Simulator
	AspectKeyNumUEsRANSim = "RRC.Conn.Avg"

	// AspectKeyNumUEsOAI is the UENIB aspect key of the number of UEs for OAI
	AspectKeyNumUEsOAI = "RRC.ConnMean"
)

var log = logging.GetLogger("uenib")

// NewHandler generates the new UENIB handler
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

// Handler includes all UENIB handler's functions
type Handler interface {
	// Get gets all UENIB
	Get(ctx context.Context) ([]Element, error)
}

type handler struct {
	uenibClient uenib.UEServiceClient
}

func (h *handler) Get(ctx context.Context) ([]Element, error) {
	req := &uenib.ListUERequest{
	}
	resp, err := h.uenibClient.ListUEs(ctx, req)
	if err != nil {
		return nil, err
	}

	results := make([]Element, 0)
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
		log.Debugf("uenib: %v", uenib)
		log.Debugf("aspects: %v", aspects)
		log.Debugf("uenibID: %v", uenibID)
		for k, v := range aspects {
			log.Debugf("k: %v", k)
			log.Debugf("v: %v", string(v.Value))
			uenibKey, err := h.createKey(uenibID, k)
			if err != nil {
				return nil, err
			}
			uenibValue, err := h.createValue(string(v.Value), k)
			if err != nil {
				return nil, err
			}
			results = append(results, Element{
				Key:   uenibKey,
				Value: uenibValue,
			})
		}
	}

	log.Debugf("Received UENIB: %v", results)
	return results, nil
}

func (h *handler) createKey(uenibID uenib.ID, aspect string) (Key, error) {
	switch aspect {
	case AspectKeyNeighbors:
		// it has nodeid, plmnid, cid, and cgi type
		nodeID, plmnID, cid, _, err := idutils.ParseUENIBNeighborAspectKey(uenibID)
		if err != nil {
			return Key{}, err
		}
		return Key{
			NodeID: nodeID,
			PlmnID: plmnID,
			CID:    cid,
			Aspect: aspect,
		}, nil
	case AspectKeyNumUEsRANSim, AspectKeyNumUEsOAI:
		// it has nodeid and coi
		nodeID, coi, err := idutils.ParseUENIBNumUEsAspectKey(uenibID)
		if err != nil {
			return Key{}, err
		}
		return Key{
			NodeID: nodeID,
			COI:    coi,
			Aspect: aspect,
		}, nil
	default:
		return Key{}, errors.NewNotSupported("unavailable aspects for this app")
	}
}

func (h *handler) createValue(value string, aspect string) (string, error) {
	switch aspect {
	case AspectKeyNeighbors:
		return idutils.ParseUENIBNeighborAspectValue(value)
	case AspectKeyNumUEsRANSim, AspectKeyNumUEsOAI:
		return value, nil
	default:
		return "", errors.NewNotSupported("unavailable aspects for this app")
	}
}
