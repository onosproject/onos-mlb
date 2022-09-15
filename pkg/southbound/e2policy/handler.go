// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package e2policy

import (
	"context"
	"fmt"
	prototypes "github.com/gogo/protobuf/types"
	e2api "github.com/onosproject/onos-api/go/onos/e2t/e2/v1beta1"
	topoapi "github.com/onosproject/onos-api/go/onos/topo"
	"github.com/onosproject/onos-lib-go/pkg/errors"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	"github.com/onosproject/onos-mlb/pkg/nib/rnib"
	"github.com/onosproject/onos-mlb/pkg/store/storage"
	subscriptionutil "github.com/onosproject/onos-mlb/pkg/utils/subscription"
	e2client "github.com/onosproject/onos-ric-sdk-go/pkg/e2/v1beta1"
	meastype "github.com/onosproject/rrm-son-lib/pkg/model/measurement/type"
	"strconv"
	"strings"
	"sync"
	"time"
)

var log = logging.GetLogger()

const (
	// DefaultE2TPort is the default E2T port
	DefaultE2TPort = 5150

	smNameDef = "oran-e2sm-rc"
	oidDef    = "1.3.6.1.4.1.53148.1.1.2.3"
)

func NewHandler(smName string, smVersion string, appID string, e2tEndpoint string, rnibHandler rnib.Handler) Handler {
	var e2tPort int
	e2tHost := strings.Split(e2tEndpoint, ":")[0]
	e2tPort, err := strconv.Atoi(strings.Split(e2tEndpoint, ":")[1])
	if err != nil {
		log.Warnf("Failed to cast e2t port - port is not a number - use default value: %v", DefaultE2TPort)
		e2tPort = DefaultE2TPort
	}

	return &handler{
		e2client: e2client.NewClient(
			e2client.WithServiceModel(e2client.ServiceModelName(smName), e2client.ServiceModelVersion(smVersion)),
			e2client.WithAppID(e2client.AppID(appID)),
			e2client.WithE2TAddress(e2tHost, e2tPort)),
		rnibHandler: rnibHandler,
		subMap:      make(map[string]string),
	}
}

type Handler interface {
	SetPolicyForOcn(ctx context.Context, nodeID string, ocns map[storage.IDs]meastype.QOffsetRange) error
}

type handler struct {
	e2client    e2client.Client
	rnibHandler rnib.Handler
	subMap      map[string]string // key: e2 node id, value: sub name
	mu          sync.Mutex
}

func (h *handler) SetPolicyForOcn(ctx context.Context, nodeID string, ocns map[storage.IDs]meastype.QOffsetRange) error {
	policyForOcns := make([]subscriptionutil.PolicyForOcn, 0)
	for k, v := range ocns {
		policyID, err := subscriptionutil.CreatePolicyID(k.CellID)
		if err != nil {
			return err
		}

		plmnIDStr := k.PlmnID
		plmnIDHex, err := strconv.ParseUint(plmnIDStr, 16, 64)
		if err != nil {
			return err
		}
		cellIDStr := k.CellID
		cellIDHex, err := strconv.ParseUint(cellIDStr, 16, 64)
		if err != nil {
			return err
		}
		ncgi := fmt.Sprintf("%x", (plmnIDHex<<36)+cellIDHex)
		policyForOcns = append(policyForOcns, subscriptionutil.PolicyForOcn{
			PolicyID: policyID,
			Nrcgi:    ncgi,
			Offset:   int(v),
		})
	}
	err := h.createSubscription(ctx, nodeID, policyForOcns)
	if err != nil {
		return err
	}
	return nil
}

func (h *handler) createSubscription(ctx context.Context, nodeID string, policies []subscriptionutil.PolicyForOcn) error {
	log.Infof("Creating subscription for E2 node with ID: %v, policies: %+v", nodeID, policies)

	actions := make([]e2api.Action, 0)

	eventTriggerData, err := subscriptionutil.CreateEventTriggerDefinition()
	if err != nil {
		log.Error(err)
		return err
	}
	aspects, err := h.rnibHandler.GetE2NodeAspects(ctx, topoapi.ID(nodeID))
	if err != nil {
		log.Warn(err)
		return err
	}

	_, err = h.getRanFunction(aspects.ServiceModels)
	if err != nil {
		log.Warn(err)
		return err
	}

	action, err := subscriptionutil.CreateSubscriptionActions(policies)
	if err != nil {
		return err
	}
	actions = append(actions, *action)

	ch := make(chan e2api.Indication)
	node := h.e2client.Node(e2client.NodeID(nodeID))
	subName := fmt.Sprintf("onos-mlb-subscription-%v", time.Now().Nanosecond())
	subSpec := e2api.SubscriptionSpec{
		Actions: actions,
		EventTrigger: e2api.EventTrigger{
			Payload: eventTriggerData,
		},
	}

	channelID, err := node.Subscribe(ctx, subName, subSpec, ch)
	if err != nil {
		log.Warn(err)
		return err
	}
	log.Infof("Subscribe: %s / %+v", subName, subSpec)
	log.Debugf("Channel ID: %s", channelID)

	h.mu.Lock()
	if oldSubName, ok := h.subMap[nodeID]; ok {
		err = node.Unsubscribe(ctx, oldSubName)
		if err != nil {
			log.Warn(err)
		}
		log.Infof("Unsubscribe: %s", subName)
	}

	h.subMap[nodeID] = subName
	h.mu.Unlock()

	return nil
}

func (h *handler) getRanFunction(serviceModelsInfo map[string]*topoapi.ServiceModelInfo) (*topoapi.RCRanFunction, error) {
	for _, sm := range serviceModelsInfo {
		smName := strings.ToLower(sm.Name)
		if smName == smNameDef && sm.OID == oidDef {
			rcRanFunction := &topoapi.RCRanFunction{}
			for _, ranFunction := range sm.RanFunctions {
				if ranFunction.TypeUrl == ranFunction.GetTypeUrl() {
					err := prototypes.UnmarshalAny(ranFunction, rcRanFunction)
					if err != nil {
						return nil, err
					}
					return rcRanFunction, nil
				}
			}
		}
	}
	return nil, errors.New(errors.NotFound, "cannot retrieve ran functions")

}
