// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package northbound

import (
	"context"
	"fmt"

	mlbapi "github.com/onosproject/onos-api/go/onos/mlb"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	"github.com/onosproject/onos-lib-go/pkg/logging/service"
	ocnstorage "github.com/onosproject/onos-mlb/pkg/store/ocn"
	paramstorage "github.com/onosproject/onos-mlb/pkg/store/parameters"
	"github.com/onosproject/onos-mlb/pkg/store/storage"
	"google.golang.org/grpc"
)

var log = logging.GetLogger()

// NewService generates a new Service for NBI
func NewService(numUEsMeasStore storage.Store,
	neighborMeasStore storage.Store,
	ocnStore ocnstorage.Store,
	paramStore paramstorage.Store) service.Service {
	return &Service{
		numUEsMeasStore:   numUEsMeasStore,
		neighborMeasStore: neighborMeasStore,
		ocnStore:          ocnStore,
		paramStore:        paramStore,
	}
}

// Service is a struct including stores and service objects
type Service struct {
	service.Service
	numUEsMeasStore   storage.Store
	neighborMeasStore storage.Store
	ocnStore          ocnstorage.Store
	paramStore        paramstorage.Store
}

// Register registers gRPC server
func (s Service) Register(r *grpc.Server) {
	server := &Server{
		numUEsMeasStore:   s.numUEsMeasStore,
		neighborMeasStore: s.neighborMeasStore,
		ocnStore:          s.ocnStore,
		paramStore:        s.paramStore,
	}
	mlbapi.RegisterMlbServer(r, server)
}

// Server is a struct including stores being used for exposing metrics
type Server struct {
	numUEsMeasStore   storage.Store
	neighborMeasStore storage.Store
	ocnStore          ocnstorage.Store
	paramStore        paramstorage.Store
}

// GetMlbParams gets mlb parameters
func (s *Server) GetMlbParams(ctx context.Context, _ *mlbapi.GetMlbParamRequest) (*mlbapi.GetMlbParamResponse, error) {

	interval, err := s.paramStore.Get(ctx, "interval")
	if err != nil {
		return nil, err
	}
	overloadThreshold, err := s.paramStore.Get(ctx, "overload_threshold")
	if err != nil {
		return nil, err
	}
	targetThreshold, err := s.paramStore.Get(ctx, "target_threshold")
	if err != nil {
		return nil, err
	}
	deltaOcn, err := s.paramStore.Get(ctx, "delta_ocn")
	if err != nil {
		return nil, err
	}

	resp := &mlbapi.GetMlbParamResponse{
		Interval:          int32(interval),
		OverloadThreshold: int32(overloadThreshold),
		TargetThreshold:   int32(targetThreshold),
		DeltaOcn:          int32(deltaOcn),
	}

	return resp, nil
}

// SetMlbParams sets mlb parameters
func (s *Server) SetMlbParams(ctx context.Context, request *mlbapi.SetMlbParamRequest) (*mlbapi.SetMlbParamResponse, error) {
	err := s.paramStore.Put(ctx, "interval", int(request.GetInterval()))
	if err != nil {
		return &mlbapi.SetMlbParamResponse{
			Success: false,
		}, nil
	}
	err = s.paramStore.Put(ctx, "delta_ocn", int(request.GetDeltaOcn()))
	if err != nil {
		return &mlbapi.SetMlbParamResponse{
			Success: false,
		}, nil
	}
	err = s.paramStore.Put(ctx, "overload_threshold", int(request.GetOverloadThreshold()))
	if err != nil {
		return &mlbapi.SetMlbParamResponse{
			Success: false,
		}, nil
	}
	err = s.paramStore.Put(ctx, "target_threshold", int(request.GetTargetThreshold()))
	if err != nil {
		return &mlbapi.SetMlbParamResponse{
			Success: false,
		}, nil
	}

	return &mlbapi.SetMlbParamResponse{
		Success: true,
	}, nil
}

// GetOcn gets Ocn map
func (s *Server) GetOcn(ctx context.Context, _ *mlbapi.GetOcnRequest) (*mlbapi.GetOcnResponse, error) {
	ch := make(chan ocnstorage.Entry)
	go func(ch chan ocnstorage.Entry) {
		err := s.ocnStore.ListAllInnerElement(ctx, ch)
		if err != nil {
			log.Warn(err)
			close(ch)
		}
	}(ch)

	mapOcnResp := make(map[string]*mlbapi.OcnRecord)

	// Init map in ocnresp message
	for e := range ch {
		key := fmt.Sprintf("%s:%s:%s:%s", e.Key.NodeID, e.Key.PlmnID, e.Key.CellID, e.Key.CellObjID)
		if _, ok := mapOcnResp[key]; !ok {
			mapOcnResp[key] = &mlbapi.OcnRecord{
				OcnRecord: make(map[string]int32),
			}
		}
		innerKey := fmt.Sprintf("%s:%s:%s:%s", e.Value.Key.NodeID, e.Value.Key.PlmnID, e.Value.Key.CellID, e.Value.Key.CellObjID)
		value := e.Value.Value
		mapOcnResp[key].OcnRecord[innerKey] = int32(value)
	}

	return &mlbapi.GetOcnResponse{
		OcnMap: mapOcnResp,
	}, nil
}
