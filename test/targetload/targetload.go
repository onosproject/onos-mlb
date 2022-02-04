// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package targetload

import (
	"context"
	"github.com/onosproject/onos-lib-go/pkg/certs"
	"github.com/onosproject/onos-mlb/pkg/manager"
	"github.com/onosproject/onos-mlb/test/utils"
	"github.com/stretchr/testify/assert"
	"testing"
)

// TestTargetLoadedCellsMlb is the test function for Targetload suite
func (s *TestSuite) TestTargetLoadedCellsMlb(t *testing.T) {
	cfg := manager.AppParameters{
		CAPath:              "/tmp/tls.cacrt",
		KeyPath:             "/tmp/tls.key",
		CertPath:            "/tmp/tls.crt",
		ConfigPath:          "/tmp/config.json",
		E2tEndpoint:         "onos-e2t:5150",
		UENIBEndpoint:       "onos-uenib:5150",
		GRPCPort:            5150,
		RicActionID:         10,
		OverloadThreshold:   utils.HighestThreshold,
		TargetLoadThreshold: utils.LowestThreshold,
	}

	_, err := certs.HandleCertPaths(cfg.CAPath, cfg.KeyPath, cfg.CertPath, true)
	assert.NoError(t, err)

	mgr := manager.NewManager(cfg)
	go func() {
		err = mgr.Start()
		assert.NoError(t, err)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TestTimeout)
	defer cancel()

	err = utils.WaitForOcnNoChangeAfterExecMLB(ctx, t, mgr)
	assert.NoError(t, err)

	t.Log("Targetload suite test passed")
}
