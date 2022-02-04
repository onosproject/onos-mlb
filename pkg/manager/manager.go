// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package manager

import (
	"context"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	"github.com/onosproject/onos-lib-go/pkg/northbound"
	"github.com/onosproject/onos-mlb/pkg/config"
	"github.com/onosproject/onos-mlb/pkg/controller"
	"github.com/onosproject/onos-mlb/pkg/monitor"
	"github.com/onosproject/onos-mlb/pkg/nib/rnib"
	mlbnbi "github.com/onosproject/onos-mlb/pkg/northbound"
	"github.com/onosproject/onos-mlb/pkg/southbound/e2control"
	ocnstorage "github.com/onosproject/onos-mlb/pkg/store/ocn"
	paramstorage "github.com/onosproject/onos-mlb/pkg/store/parameters"
	"github.com/onosproject/onos-mlb/pkg/store/storage"
)

var log = logging.GetLogger("manager")

// AppParameters includes all application parameters coming from arguments when starting this app
type AppParameters struct {
	CAPath              string
	KeyPath             string
	CertPath            string
	ConfigPath          string
	E2tEndpoint         string
	UENIBEndpoint       string
	GRPCPort            int
	RicActionID         int32
	OverloadThreshold   int
	TargetLoadThreshold int
}

// NewManager generates this application's manager
func NewManager(parameters AppParameters) *Manager {
	appCfg, err := config.NewConfig(parameters.ConfigPath)
	if err != nil {
		log.Warn(err)
	}
	interval, err := appCfg.GetInterval(MLBAppIntervalPath)
	if err != nil {
		log.Warn("set interval to default interval - reason: %v", err)
		interval = MLBAppDefaultInterval
	}

	numUEsMeasStore := storage.NewStore()
	neighborMeasStore := storage.NewStore()
	ocnStore := ocnstorage.NewStore()
	paramStore := paramstorage.NewStore()
	err = paramStore.Put(context.Background(), "interval", interval)
	if err != nil {
		log.Error(err)
	}
	err = paramStore.Put(context.Background(), "delta_ocn", OCNDeltaFactor)
	if err != nil {
		log.Error(err)
	}
	err = paramStore.Put(context.Background(), "overload_threshold", parameters.OverloadThreshold)
	if err != nil {
		log.Error(err)
	}
	err = paramStore.Put(context.Background(), "target_threshold", parameters.TargetLoadThreshold)
	if err != nil {
		log.Error(err)
	}

	rnibHandler, err := rnib.NewHandler()
	if err != nil {
		log.Error(err)
	}
	monitorHandler := monitor.NewHandler(rnibHandler, numUEsMeasStore, neighborMeasStore, ocnStore)

	e2ControlHandler := e2control.NewHandler(RcPreServiceModelName, RcPreServiceModelVersion,
		AppID, parameters.E2tEndpoint)

	ctrlHandler := controller.NewHandler(e2ControlHandler, monitorHandler, numUEsMeasStore, neighborMeasStore, ocnStore, paramStore)

	return &Manager{
		handlers: handlers{
			rnibHandler:       rnibHandler,
			monitorHandler:    monitorHandler,
			e2ControlHandler:  e2ControlHandler,
			controllerHandler: ctrlHandler,
		},
		stores: stores{
			numUEsMeasStore:   numUEsMeasStore,
			neighborMeasStore: neighborMeasStore,
			ocnStore:          ocnStore,
			paramStore:        paramStore,
		},
		channels: channels{},
		configs: configs{
			appConfigParams: parameters,
			appConfig:       appCfg,
		},
	}
}

// Manager is a struct including this app's manager information and objects
type Manager struct {
	handlers handlers
	stores   stores
	channels channels
	configs  configs
}

type handlers struct {
	rnibHandler       rnib.Handler
	monitorHandler    monitor.Handler
	e2ControlHandler  e2control.Handler
	controllerHandler controller.Handler
}

type stores struct {
	numUEsMeasStore   storage.Store
	neighborMeasStore storage.Store
	ocnStore          ocnstorage.Store
	paramStore        paramstorage.Store
}

type channels struct {
}

type configs struct {
	appConfigParams AppParameters
	appConfig       config.Config
}

// Start starts this app's manager
func (m *Manager) Start() error {
	err := m.startNorthboundServer()
	if err != nil {
		return err
	}
	err = m.handlers.controllerHandler.Run(context.Background())
	return err
}

func (m *Manager) startNorthboundServer() error {
	s := northbound.NewServer(northbound.NewServerCfg(
		m.configs.appConfigParams.CAPath,
		m.configs.appConfigParams.KeyPath,
		m.configs.appConfigParams.CertPath,
		int16(m.configs.appConfigParams.GRPCPort),
		true, northbound.SecurityConfig{}))
	s.AddService(mlbnbi.NewService(m.stores.numUEsMeasStore,
		m.stores.neighborMeasStore,
		m.stores.ocnStore,
		m.stores.paramStore))

	doneCh := make(chan error)
	go func() {
		err := s.Serve(func(started string) {
			log.Info("Started NBI on ", started)
			close(doneCh)
		})
		if err != nil {
			doneCh <- err
		}
	}()
	return <-doneCh
}

// GetOcnStore returns Ocn store
func (m *Manager) GetOcnStore() ocnstorage.Store {
	return m.stores.ocnStore
}

// GetNumUEsStore returns NumUEsStore
func (m *Manager) GetNumUEsStore() storage.Store {
	return m.stores.numUEsMeasStore
}

// GetNeighborStore returns neighbor store
func (m *Manager) GetNeighborStore() storage.Store {
	return m.stores.neighborMeasStore
}
