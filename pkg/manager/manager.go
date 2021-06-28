// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package manager

import (
	"context"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	"github.com/onosproject/onos-mlb/pkg/config"
	"github.com/onosproject/onos-mlb/pkg/controller"
	"github.com/onosproject/onos-mlb/pkg/monitor"
	"github.com/onosproject/onos-mlb/pkg/nib/rnib"
	"github.com/onosproject/onos-mlb/pkg/nib/uenib"
	"github.com/onosproject/onos-mlb/pkg/southbound/e2control"
	"github.com/onosproject/onos-mlb/pkg/store/storage"
)

var log = logging.GetLogger("manager")

type AppParameters struct {
	CAPath              string
	KeyPath             string
	CertPath            string
	ConfigPath          string
	E2tEndpoint         string
	UENIBEndpoint	    string
	GRPCPort            int
	RicActionID         int32
	OverloadThreshold   int
	TargetLoadThreshold int
}

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
	statisticsStore := storage.NewStore()
	ocnStore := storage.NewStore()

	rnibHandler, err := rnib.NewHandler()
	if err != nil {
		log.Error(err)
	}
	uenibHandler, err := uenib.NewHandler(parameters.UENIBEndpoint, parameters.CertPath, parameters.KeyPath)
	if err != nil {
		log.Error(err)
	}
	monitorHandler := monitor.NewHandler(rnibHandler, uenibHandler, numUEsMeasStore, neighborMeasStore, ocnStore)

	e2ControlHandler := e2control.NewHandler(RcPreServiceModelName, RcPreServiceModelVersion,
		AppID, parameters.E2tEndpoint)

	ctrlHandler := controller.NewHandler(interval, e2ControlHandler, monitorHandler, numUEsMeasStore, neighborMeasStore, statisticsStore, ocnStore, parameters.OverloadThreshold, parameters.TargetLoadThreshold)

	return &Manager{
		handlers: handlers{
			rnibHandler: rnibHandler,
			uenibHandler: uenibHandler,
			monitorHandler: monitorHandler,
			e2ControlHandler: e2ControlHandler,
			controllerHandler: ctrlHandler,
		},
		stores: stores{
			numUEsMeasStore: numUEsMeasStore,
			neighborMeasStore: neighborMeasStore,
			statisticsStore: statisticsStore,
			ocnStore: ocnStore,
		},
		channels: channels{},
		configs: configs{
			appConfigParams: parameters,
			appConfig: appCfg,
		},
	}
}

type Manager struct {
	handlers handlers
	stores stores
	channels channels
	configs configs
}

type handlers struct {
	rnibHandler rnib.Handler
	uenibHandler uenib.Handler
	monitorHandler monitor.Handler
	e2ControlHandler e2control.Handler
	controllerHandler controller.Handler
}

type stores struct {
	numUEsMeasStore storage.Store
	neighborMeasStore storage.Store
	statisticsStore storage.Store
	ocnStore storage.Store
}

type channels struct {

}

type configs struct {
	appConfigParams AppParameters
	appConfig config.Config
}

func (m *Manager) Start() error {
	m.handlers.controllerHandler.Run(context.Background())
	return nil
}