// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package e2control

import (
	"context"
	e2api "github.com/onosproject/onos-api/go/onos/e2t/e2/v1beta1"
	"github.com/onosproject/onos-e2-sm/servicemodels/e2sm_rc_pre/pdubuilder"
	e2sm_rc_pre_v2 "github.com/onosproject/onos-e2-sm/servicemodels/e2sm_rc_pre/v2/e2sm-rc-pre-v2"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	"github.com/onosproject/onos-mlb/pkg/store/storage"
	decodeutils "github.com/onosproject/onos-mlb/pkg/utils/decode"
	idutils "github.com/onosproject/onos-mlb/pkg/utils/parse"
	e2client "github.com/onosproject/onos-ric-sdk-go/pkg/e2/v1beta1"
	"google.golang.org/protobuf/proto"
	"strconv"
	"strings"
)

const (
	// DefaultE2TPort is the default E2T port
	DefaultE2TPort = 5150

	// RcPreRanParamIDForOCN is the ranparam_id used in RC-PRE control message
	RcPreRanParamIDForOCN = 20

	// RcPreRanParamNameForOCN is the ranparam_name used in RC-PRE control message
	RcPreRanParamNameForOCN = "ocn_rc"
)

var log = logging.GetLogger("southbound", "e2control")

// NewHandler generates the new handler of this e2control session handler
func NewHandler(smName string, smVersion string, appID string, e2tEndpoint string) Handler {
	var e2tPort int
	e2tHost := strings.Split(e2tEndpoint, ":")[0]
	e2tPort, err := strconv.Atoi(strings.Split(e2tEndpoint, ":")[1])
	if err != nil {
		log.Warnf("Failed to cast e2t port - port is not a number - use default value")
		e2tPort = DefaultE2TPort
	}

	return &handler{
		e2client: e2client.NewClient(
			e2client.WithServiceModel(e2client.ServiceModelName(smName), e2client.ServiceModelVersion(smVersion)),
			e2client.WithAppID(e2client.AppID(appID)),
			e2client.WithE2TAddress(e2tHost, e2tPort)),
	}
}

// Handler includes all functions of E2 control handler
type Handler interface {
	// SendControlMessage sends a control message to E2Node
	SendControlMessage(ctx context.Context, ids storage.IDs, nodeID string, offset int32) error
}

type handler struct {
	e2client e2client.Client
}

func (h *handler) SendControlMessage(ctx context.Context, ids storage.IDs, nodeID string, offset int32) error {
	log.Debugf("Sending control message: nCellID %v, nodeID %v, offset %v", ids, nodeID, offset)
	header, err := h.createRcControlHeader(ids)
	if err != nil {
		return err
	}
	payload, err := h.createRcControlMessage(RcPreRanParamIDForOCN, RcPreRanParamNameForOCN, offset)
	if err != nil {
		return err
	}

	node := h.e2client.Node(e2client.NodeID(nodeID))
	outcome, err := node.Control(ctx, &e2api.ControlMessage{
		Header:  header,
		Payload: payload,
	})
	if err != nil {
		return err
	}
	log.Infof("Outcome: %v", outcome)
	return nil
}

func (h *handler) createRcControlHeader(ids storage.IDs) ([]byte, error) {
	plmnid, err := decodeutils.DecodePlmnIDHexStrToBytes(ids.PlmnID)
	if err != nil {
		return nil, err
	}
	cid, err := decodeutils.DecodeCIDHexStrToUint64(ids.CellID)
	if err != nil {
		return nil, err
	}

	cgi := &e2sm_rc_pre_v2.CellGlobalId{
		CellGlobalId: &e2sm_rc_pre_v2.CellGlobalId_NrCgi{
			NrCgi: &e2sm_rc_pre_v2.Nrcgi{
				PLmnIdentity: &e2sm_rc_pre_v2.PlmnIdentity{
					Value: plmnid,
				},
				NRcellIdentity: &e2sm_rc_pre_v2.NrcellIdentity{
					Value: &e2sm_rc_pre_v2.BitString{
						Value: idutils.Uint64ToBitString(cid, 36),
						Len:   36,
					},
				},
			},
		},
	}
	e2SmRcPrePdu, err := pdubuilder.CreateE2SmRcPreControlHeader(nil, cgi)

	if err != nil {
		return []byte{}, err
	}

	err = e2SmRcPrePdu.Validate()

	if err != nil {
		return []byte{}, err
	}
	protoBytes, err := proto.Marshal(e2SmRcPrePdu)
	if err != nil {
		return []byte{}, err
	}

	return protoBytes, nil

}

func (h *handler) createRcControlMessage(ranParamID int32, ranParamName string, ranParamValue int32) ([]byte, error) {
	ranParamValueInt, err := pdubuilder.CreateRanParameterValueInt(uint32(ranParamValue))
	if err != nil {
		return []byte{}, err
	}
	newE2SmRcPrePdu, err := pdubuilder.CreateE2SmRcPreControlMessage(ranParamID, ranParamName, ranParamValueInt)
	if err != nil {
		return []byte{}, err
	}

	err = newE2SmRcPrePdu.Validate()
	if err != nil {
		return []byte{}, err
	}

	protoBytes, err := proto.Marshal(newE2SmRcPrePdu)
	if err != nil {
		return []byte{}, err
	}

	return protoBytes, nil
}
