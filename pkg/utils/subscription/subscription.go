// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package subscriptionutil

import (
	e2api "github.com/onosproject/onos-api/go/onos/e2t/e2/v1beta1"
	"github.com/onosproject/onos-e2-sm/servicemodels/e2sm_rc/pdubuilder"
	e2smcommonies "github.com/onosproject/onos-e2-sm/servicemodels/e2sm_rc/v1/e2sm-common-ies"
	e2smrcies "github.com/onosproject/onos-e2-sm/servicemodels/e2sm_rc/v1/e2sm-rc-ies"
	"github.com/prometheus/common/log"
	"google.golang.org/protobuf/proto"
	"strconv"
)

func CreateEventTriggerDefinition() ([]byte, error) {
	eventTriggerUeEventIDItem, err := pdubuilder.CreateEventTriggerUeeventInfoItem(2)
	if err != nil {
		return nil, err
	}

	eventTriggerUeEventInfo := &e2smrcies.EventTriggerUeeventInfo{
		UeEventList: []*e2smrcies.EventTriggerUeeventInfoItem{eventTriggerUeEventIDItem},
	}

	eventTriggerItem, err := pdubuilder.CreateE2SmRcEventTriggerFormat1Item(1, &e2smrcies.MessageTypeChoice{
		MessageTypeChoice: &e2smrcies.MessageTypeChoice_MessageTypeChoiceRrc{
			MessageTypeChoiceRrc: &e2smrcies.MessageTypeChoiceRrc{
				RRcMessage: &e2smcommonies.RrcMessageId{
					RrcType: &e2smcommonies.RrcType{
						RrcType: &e2smcommonies.RrcType_Nr{
							Nr: e2smcommonies.RrcclassNr_RRCCLASS_NR_U_L_DCCH,
						},
					},
					MessageId: 0,
				},
			},
		},
	}, nil, nil, eventTriggerUeEventInfo, nil)
	if err != nil {
		return nil, err
	}

	itemList := []*e2smrcies.E2SmRcEventTriggerFormat1Item{eventTriggerItem}

	rcEventTriggerDefinitionFormat1, err := pdubuilder.CreateE2SmRcEventTriggerFormat1(itemList)
	if err != nil {
		return nil, err
	}

	err = rcEventTriggerDefinitionFormat1.Validate()
	if err != nil {
		return nil, err
	}

	protoBytes, err := proto.Marshal(rcEventTriggerDefinitionFormat1)
	if err != nil {
		return nil, err
	}

	return protoBytes, nil
}

type PolicyForOcn struct {
	PolicyID int
	Nrcgi    string
	Offset   int
}

func CreateSubscriptionActions(policies []PolicyForOcn) (*e2api.Action, error) {
	log.Infof("Create subscription for policies: %+v", policies)

	// create RIC Policy Condition List to be used in Action Definition Format2
	rpcl := make([]*e2smrcies.E2SmRcActionDefinitionFormat2Item, 0)

	for _, policy := range policies {
		// create policy action for each policy
		ricPolicyActionID := &e2smrcies.RicControlActionId{
			Value: int32(policy.PolicyID),
		}
		ricPolicyDecision := e2smrcies.RicPolicyDecision_RIC_POLICY_DECISION_ACCEPT
		ranParameterList := make([]*e2smrcies.RicPolicyActionRanparameterItem, 0)
		targetPrimaryCellID, err := createRanParameterItemTargetPrimaryCellID(policy.Nrcgi)
		if err != nil {
			return nil, err
		}
		targetOcn, err := createRanParameterItemCellSpecificOffset(policy.Offset)
		if err != nil {
			return nil, err
		}
		ranParameterList = append(ranParameterList, targetPrimaryCellID)
		ranParameterList = append(ranParameterList, targetOcn)

		ricPolicyAction := &e2smrcies.RicPolicyAction{
			RicPolicyActionId: ricPolicyActionID,
			RanParametersList: ranParameterList,
			RicPolicyDecision: &ricPolicyDecision,
		}

		// create RIC Policy Condition Definition in RIC Policy Condition
		ranParameterTestingList := make([]*e2smrcies.RanparameterTestingItem, 0)
		targetPrimaryCellIDTesting, err := createRanParameterTestingItemTargetPrimaryCellID(policy.Nrcgi)
		if err != nil {
			return nil, err
		}
		targetOcnTesting, err := createRanParameterTestingItemCellSpecificOffset(policy.Offset)
		if err != nil {
			return nil, err
		}
		ranParameterTestingList = append(ranParameterTestingList, targetPrimaryCellIDTesting)
		ranParameterTestingList = append(ranParameterTestingList, targetOcnTesting)
		ricPolicyConditionDefinition := &e2smrcies.RanparameterTesting{
			Value: ranParameterTestingList,
		}

		rpc := &e2smrcies.E2SmRcActionDefinitionFormat2Item{
			RicPolicyAction:              ricPolicyAction,
			RicPolicyConditionDefinition: ricPolicyConditionDefinition,
		}

		rpcl = append(rpcl, rpc)
	}

	ad, err := pdubuilder.CreateE2SmRcActionDefinitionFormat2(1, rpcl)
	if err != nil {
		return nil, err
	}

	err = ad.Validate()
	if err != nil {
		return nil, err
	}

	adProto, err := proto.Marshal(ad)
	if err != nil {
		return nil, err
	}

	action := &e2api.Action{
		ID:      1,
		Type:    e2api.ActionType_ACTION_TYPE_POLICY,
		Payload: adProto,
	}

	return action, nil
}

func createRanParameterItemTargetPrimaryCellID(nrcgi string) (*e2smrcies.RicPolicyActionRanparameterItem, error) {
	nrcgiRanParamValuePrint, err := pdubuilder.CreateRanparameterValuePrintableString(nrcgi)
	if err != nil {
		return nil, err
	}
	nrCgiRanParamValue, err := pdubuilder.CreateRanparameterValueTypeChoiceElementFalse(nrcgiRanParamValuePrint)
	if err != nil {
		return nil, err
	}
	nrCgiRanParamValueItem, err := pdubuilder.CreateRanparameterStructureItem(4, nrCgiRanParamValue)
	if err != nil {
		return nil, err
	}
	nrCellRanParamValue, err := pdubuilder.CreateRanParameterStructure([]*e2smrcies.RanparameterStructureItem{nrCgiRanParamValueItem})
	if err != nil {
		return nil, err
	}
	nrCellRanParamValueType, err := pdubuilder.CreateRanparameterValueTypeChoiceStructure(nrCellRanParamValue)
	if err != nil {
		return nil, err
	}
	nrCellRanParamValueItem, err := pdubuilder.CreateRanparameterStructureItem(3, nrCellRanParamValueType)
	if err != nil {
		return nil, err
	}
	targetCellRanParamValue, err := pdubuilder.CreateRanParameterStructure([]*e2smrcies.RanparameterStructureItem{nrCellRanParamValueItem})
	if err != nil {
		return nil, err
	}
	targetCellRanParamValueType, err := pdubuilder.CreateRanparameterValueTypeChoiceStructure(targetCellRanParamValue)
	if err != nil {
		return nil, err
	}
	targetCellRanParamValueItem, err := pdubuilder.CreateRanparameterStructureItem(2, targetCellRanParamValueType)
	if err != nil {
		return nil, err
	}
	targetPrimaryCellIDRanParamValue, err := pdubuilder.CreateRanParameterStructure([]*e2smrcies.RanparameterStructureItem{targetCellRanParamValueItem})
	if err != nil {
		return nil, err
	}
	targetPrimaryCellIDRanParamValueType, err := pdubuilder.CreateRanparameterValueTypeChoiceStructure(targetPrimaryCellIDRanParamValue)
	if err != nil {
		return nil, err
	}
	targetPrimaryCellIDRanParamValueItem, err := pdubuilder.CreateRicPolicyActionRanParameterItem(1, targetPrimaryCellIDRanParamValueType)
	if err != nil {
		return nil, err
	}

	return targetPrimaryCellIDRanParamValueItem, nil
}

func createRanParameterTestingItemTargetPrimaryCellID(nrcgi string) (*e2smrcies.RanparameterTestingItem, error) {
	logicalOr := e2smrcies.LogicalOr_LOGICAL_OR_FALSE
	nrCgiRanParamTestingCondition, err := pdubuilder.CreateRanparameterTestingConditionComparison(pdubuilder.CreateRanPChoiceComparisonContains())
	if err != nil {
		return nil, err
	}
	nrCgiRanParamType := &e2smrcies.RanParameterType{
		RanParameterType: &e2smrcies.RanParameterType_RanPChoiceElementFalse{
			RanPChoiceElementFalse: &e2smrcies.RanparameterTestingItemChoiceElementFalse{
				RanParameterTestCondition: nrCgiRanParamTestingCondition,
				LogicalOr:                 &logicalOr,
				RanParameterValue: &e2smrcies.RanparameterValue{
					RanparameterValue: &e2smrcies.RanparameterValue_ValuePrintableString{
						ValuePrintableString: nrcgi,
					},
				},
			},
		},
	}
	nrCgiRanParamTestingItem, err := pdubuilder.CreateRanparameterTestingItem(4, nrCgiRanParamType)
	if err != nil {
		return nil, err
	}
	nrCgiRanParamTestingStructure := &e2smrcies.RanparameterTestingStructure{
		Value: []*e2smrcies.RanparameterTestingItem{nrCgiRanParamTestingItem},
	}
	nrCellRanParamType, err := pdubuilder.CreateRanParameterTypeChoiceStructure(nrCgiRanParamTestingStructure)
	if err != nil {
		return nil, err
	}
	nrCellRanParamTestingItem, err := pdubuilder.CreateRanparameterTestingItem(3, nrCellRanParamType)
	if err != nil {
		return nil, err
	}
	targetCellRanParamTestingStructure := &e2smrcies.RanparameterTestingStructure{
		Value: []*e2smrcies.RanparameterTestingItem{nrCellRanParamTestingItem},
	}
	targetCellRanParamType, err := pdubuilder.CreateRanParameterTypeChoiceStructure(targetCellRanParamTestingStructure)
	if err != nil {
		return nil, err
	}
	targetCellRanParamTestingItem, err := pdubuilder.CreateRanparameterTestingItem(2, targetCellRanParamType)
	if err != nil {
		return nil, err
	}
	targetPrimaryCellIDRanParamTestingStructure := &e2smrcies.RanparameterTestingStructure{
		Value: []*e2smrcies.RanparameterTestingItem{targetCellRanParamTestingItem},
	}
	targetPrimaryCellIDRanParamType, err := pdubuilder.CreateRanParameterTypeChoiceStructure(targetPrimaryCellIDRanParamTestingStructure)
	if err != nil {
		return nil, err
	}
	targetPrimaryCellIDRanParamTestingItem, err := pdubuilder.CreateRanparameterTestingItem(1, targetPrimaryCellIDRanParamType)
	if err != nil {
		return nil, err
	}

	return targetPrimaryCellIDRanParamTestingItem, nil
}

func createRanParameterItemCellSpecificOffset(ocn int) (*e2smrcies.RicPolicyActionRanparameterItem, error) {
	ocnRanParamValueInt, err := pdubuilder.CreateRanparameterValueInt(int64(ocn))
	if err != nil {
		return nil, err
	}
	ocnRanParamValue, err := pdubuilder.CreateRanparameterValueTypeChoiceElementFalse(ocnRanParamValueInt)
	if err != nil {
		return nil, err
	}
	ocnRanParamValueItem, err := pdubuilder.CreateRicPolicyActionRanParameterItem(10201, ocnRanParamValue)
	if err != nil {
		return nil, err
	}
	return ocnRanParamValueItem, nil
}

func createRanParameterTestingItemCellSpecificOffset(ocn int) (*e2smrcies.RanparameterTestingItem, error) {
	logicalOr := e2smrcies.LogicalOr_LOGICAL_OR_FALSE
	ocnRanParamTestingCondition, err := pdubuilder.CreateRanparameterTestingConditionComparison(pdubuilder.CreateRanPChoiceComparisonDifference())
	if err != nil {
		return nil, err
	}
	ocnRanParamType := &e2smrcies.RanParameterType{
		RanParameterType: &e2smrcies.RanParameterType_RanPChoiceElementFalse{
			RanPChoiceElementFalse: &e2smrcies.RanparameterTestingItemChoiceElementFalse{
				RanParameterTestCondition: ocnRanParamTestingCondition,
				LogicalOr:                 &logicalOr,
				RanParameterValue: &e2smrcies.RanparameterValue{
					RanparameterValue: &e2smrcies.RanparameterValue_ValueInt{
						ValueInt: int64(ocn),
					},
				},
			},
		},
	}
	ocnRanParamTestingItem, err := pdubuilder.CreateRanparameterTestingItem(10201, ocnRanParamType)
	if err != nil {
		return nil, err
	}
	return ocnRanParamTestingItem, nil
}

func CreatePolicyID(cgi string) (int, error) {
	cgiUInt64, err := strconv.ParseUint(cgi, 16, 64)
	if err != nil {
		return 0, err
	}

	policyID := cgiUInt64 & 0xFFFF
	return int(policyID), nil
}
