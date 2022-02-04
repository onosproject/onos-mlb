// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"fmt"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	app "github.com/onosproject/onos-ric-sdk-go/pkg/config/app/default"
	configurable "github.com/onosproject/onos-ric-sdk-go/pkg/config/registry"
	"strconv"
)

var log = logging.GetLogger("config")

// Config is an interface including app config
type Config interface {
	GetInterval(path string) (int, error)
}

// AppConfig is a struct including app config
type AppConfig struct {
	appConfig *app.Config
}

// NewConfig initialize the xApp config
func NewConfig(path string) (*AppConfig, error) {
	appConfig, err := configurable.RegisterConfigurable(path, &configurable.RegisterRequest{})
	if err != nil {
		return nil, err
	}

	cfg := &AppConfig{
		appConfig: appConfig.Config.(*app.Config),
	}
	return cfg, nil
}

// GetInterval gets interval
func (c *AppConfig) GetInterval(path string) (int, error) {
	interval, _ := c.appConfig.Get(path)
	val, err := strconv.Atoi(fmt.Sprintf("%v", interval.Value))
	if err != nil {
		log.Error(err)
		return 0, err
	}

	return val, nil
}
