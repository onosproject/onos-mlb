// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package config

import (
	"github.com/onosproject/onos-lib-go/pkg/logging"
	app "github.com/onosproject/onos-ric-sdk-go/pkg/config/app/default"
	configurable "github.com/onosproject/onos-ric-sdk-go/pkg/config/registry"
	configutils "github.com/onosproject/onos-ric-sdk-go/pkg/config/utils"
)

var log = logging.GetLogger("config", "appConfig")

func NewAppConfig(configPath string) (Config, error) {
	conf, err := configurable.RegisterConfigurable(configPath, &configurable.RegisterRequest{})
	if err != nil {
		return nil, err
	}

	cfg := &appConfig{
		appConfig: conf.Config.(*app.Config),
	}

	return cfg, nil
}

type appConfig struct {
	appConfig *app.Config
}

func (a *appConfig) GetConfigUint64(path string) (uint64, error) {
	value, err := a.appConfig.Get(path)
	if err != nil {
		return 0, err
	}

	valueUint64, err := configutils.ToUint64(value)
	if err != nil {
		return 0, err
	}

	return valueUint64, nil
}
