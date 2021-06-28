// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package config

import (
	"fmt"
	"github.com/onosproject/onos-lib-go/pkg/errors"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	app "github.com/onosproject/onos-ric-sdk-go/pkg/config/app/default"
	configurable "github.com/onosproject/onos-ric-sdk-go/pkg/config/registry"
	"github.com/openconfig/gnmi/proto/gnmi"
	"strconv"
)

var log = logging.GetLogger("config")

type Config interface {
	GetInterval(path string) (int, error)
}

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

// GetInterval gets report period
func (c *AppConfig) GetInterval(path string) (int, error) {
	interval, _ := c.appConfig.Get(path)
	val, err := c.ToInt(interval.Value)
	val, err = strconv.Atoi(fmt.Sprintf("%v", interval.Value))
	if err != nil {
		log.Error(err)
		return 0, err
	}

	return val, nil
}

func (c *AppConfig) ToInt(value interface{}) (int, error) {
	switch v := value.(type) {
	case *gnmi.TypedValue:
		return int(toGnmiTypedValue(value).GetIntVal()), nil
	case float64:
		return int(value.(float64)), nil
	case uint64:
		return int(value.(float64)), nil

	default:
		return 0, errors.New(errors.NotSupported, "Not supported type %v", v)
	}
}

// ToGnmiTypedValue
func toGnmiTypedValue(value interface{}) *gnmi.TypedValue {
	return value.(*gnmi.TypedValue)
}