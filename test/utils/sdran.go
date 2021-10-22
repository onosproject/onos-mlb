// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-Only-1.0

package utils

import (
	"context"
	"fmt"
	"github.com/onosproject/helmit/pkg/helm"
	"github.com/onosproject/helmit/pkg/input"
	"github.com/onosproject/helmit/pkg/kubernetes"
	"github.com/onosproject/onos-mlb/pkg/manager"
	ocnstorage "github.com/onosproject/onos-mlb/pkg/store/ocn"
	"github.com/onosproject/onos-mlb/pkg/store/storage"
	"github.com/onosproject/onos-test/pkg/onostest"
	meastype "github.com/onosproject/rrm-son-lib/pkg/model/measurement/type"
	"testing"
	"time"
)

func getCredentials() (string, string, error) {
	kubClient, err := kubernetes.New()
	if err != nil {
		return "", "", err
	}
	secrets, err := kubClient.CoreV1().Secrets().Get(context.Background(), onostest.SecretsName)
	if err != nil {
		return "", "", err
	}
	username := string(secrets.Object.Data["sd-ran-username"])
	password := string(secrets.Object.Data["sd-ran-password"])

	return username, password, nil
}

// CreateSdranRelease creates a helm release for an sd-ran instance
func CreateSdranRelease(c *input.Context) (*helm.HelmRelease, error) {
	username, password, err := getCredentials()
	registry := c.GetArg("registry").String("")

	if err != nil {
		return nil, err
	}

	sdran := helm.Chart("sd-ran", onostest.SdranChartRepo).
		Release("sd-ran").
		SetUsername(username).
		SetPassword(password).
		Set("import.onos-config.enabled", false).
		Set("import.onos-topo.enabled", true).
		Set("import.ran-simulator.enabled", true).
		Set("import.onos-pci.enabled", true).
		Set("import.onos-kpimon.enabled", true).
		Set("global.image.registry", registry)

	return sdran, nil
}

// WaitForAllOcnIncreased waits until all Ocn values increased
func WaitForAllOcnIncreased(ctx context.Context, t *testing.T, mgr *manager.Manager) error {
	store := mgr.GetOcnStore()

	for {
		select {
		case <-ctx.Done():
			if verifyOcnStoreSize(ctx, t, store) {
				if verifyOcnIncreased(ctx, t, store) {
					return nil
				}
				return fmt.Errorf("%s", "Test failed - Ocn values were not increased")
			}
			return fmt.Errorf("%s", "Test failed - Ocn store size is not matched")
		case <-time.After(TestInterval):
			if verifyOcnStoreSize(ctx, t, store) && verifyOcnIncreased(ctx, t, store) {
				return nil
			}
		}
	}
}

func verifyOcnIncreased(ctx context.Context, t *testing.T, store ocnstorage.Store) bool {
	verify := true
	ch := make(chan ocnstorage.Entry)
	go func(ch chan ocnstorage.Entry) {
		err := store.ListAllInnerElement(ctx, ch)
		if err != nil {
			close(ch)
			t.Log(err)
		}
	}(ch)

	numElem := 0
	for e := range ch {
		numElem++
		if e.Value.Value <= meastype.QOffset0dB {
			t.Logf("Waiting until Ocn values increased; currently %s", e.Value.Value.String())
			verify = false
		}
	}

	if numElem == 0 {
		verify = false
	}
	return verify
}

// WaitForAllOcnDecreased waits until all Ocn values decreased
func WaitForAllOcnDecreased(ctx context.Context, t *testing.T, mgr *manager.Manager) error {
	store := mgr.GetOcnStore()

	for {
		select {
		case <-ctx.Done():
			if verifyOcnStoreSize(ctx, t, store) {
				if verifyOcnDecreased(ctx, t, store) {
					return nil
				}
				return fmt.Errorf("%s", "Test failed - Ocn values were not decreased")
			}
			return fmt.Errorf("%s", "Test failed - Ocn store size is not matched")
		case <-time.After(TestInterval):
			if verifyOcnStoreSize(ctx, t, store) && verifyOcnDecreased(ctx, t, store) {
				return nil
			}
		}
	}
}

func verifyOcnDecreased(ctx context.Context, t *testing.T, store ocnstorage.Store) bool {
	verify := true
	ch := make(chan ocnstorage.Entry)
	go func(ch chan ocnstorage.Entry) {
		err := store.ListAllInnerElement(ctx, ch)
		if err != nil {
			close(ch)
			t.Log(err)
		}
	}(ch)

	numElem := 0
	for e := range ch {
		numElem++
		if e.Value.Value >= meastype.QOffset0dB {
			t.Logf("Waiting until Ocn values decreased; currently %s", e.Value.Value.String())
			verify = false
		}
	}

	if numElem == 0 {
		verify = false
	}
	return verify
}

// WaitForOcnNoChangeAfterExecMLB check Ocn values not changed
func WaitForOcnNoChangeAfterExecMLB(ctx context.Context, t *testing.T, mgr *manager.Manager) error {
	for {
		select {
		case <-ctx.Done():
			if verifyRNibNumUEs(ctx, t, mgr.GetNumUEsStore()) &&
				verifyRNibNeighbor(ctx, t, mgr.GetNeighborStore()) {
				if verifyOcnStoreSize(ctx, t, mgr.GetOcnStore()) {
					if verifyOcnNoChanged(ctx, t, mgr.GetOcnStore()) {
						return nil
					}
					return fmt.Errorf("%s", "Test failed - All Ocn values have to be 0dB but changed")
				}
				return fmt.Errorf("%s", "Test failed - Ocn store size is not matched")
			}
			return fmt.Errorf("%s", "Test failed - RNIB is not still ready")
		case <-time.After(TestInterval):
			if verifyRNibNumUEs(ctx, t, mgr.GetNumUEsStore()) &&
				verifyRNibNeighbor(ctx, t, mgr.GetNeighborStore()) {
				if verifyOcnStoreSize(ctx, t, mgr.GetOcnStore()) {
					if verifyOcnNoChanged(ctx, t, mgr.GetOcnStore()) {
						return nil
					}
					return fmt.Errorf("%s", "Test failed - All Ocn values have to be 0dB but changed")
				}
			}
		}
	}
}

func verifyRNibNumUEs(ctx context.Context, t *testing.T, numUEStore storage.Store) bool {
	result := 0
	ch := make(chan *storage.Entry)
	go func(ch chan *storage.Entry) {
		err := numUEStore.ListElements(ctx, ch)
		if err != nil {
			close(ch)
			t.Log(err)
		}
	}(ch)
	for e := range ch {
		result += e.Value.(storage.Measurement).Value
	}

	if result != TotalNumUEs {
		t.Log("Waiting until RNIB has the number of UEs")
		return false
	}
	return true
}

func verifyRNibNeighbor(ctx context.Context, t *testing.T, neighborStore storage.Store) bool {
	result := 0
	ch := make(chan *storage.Entry)
	go func(ch chan *storage.Entry) {
		err := neighborStore.ListElements(ctx, ch)
		if err != nil {
			close(ch)
			t.Log(err)
		}
	}(ch)
	for range ch {
		result++
	}

	if result != TotalNumCells {
		t.Log("Waiting until RNIB has neighbors")
		return false
	}
	return true
}

func verifyOcnStoreSize(ctx context.Context, t *testing.T, store ocnstorage.Store) bool {
	verify := true
	ch := make(chan ocnstorage.Entry)
	go func(ch chan ocnstorage.Entry) {
		err := store.ListAllInnerElement(ctx, ch)
		if err != nil {
			close(ch)
			t.Log(err)
		}
	}(ch)

	numElem := 0
	for c := range ch {
		t.Logf("Received store: %v", c)
		numElem++
	}
	if numElem != OcnStoreSize {
		t.Logf("Waiting until Ocn store size become %d; currently %d", OcnStoreSize, numElem)
		verify = false
	}
	return verify
}

func verifyOcnNoChanged(ctx context.Context, t *testing.T, store ocnstorage.Store) bool {
	verify := true
	ch := make(chan ocnstorage.Entry)
	go func(ch chan ocnstorage.Entry) {
		err := store.ListAllInnerElement(ctx, ch)
		if err != nil {
			close(ch)
			t.Log(err)
		}
	}(ch)

	numElem := 0
	for e := range ch {
		numElem++
		if e.Value.Value != meastype.QOffset0dB {
			t.Logf("Ocn values should be always 0dB; currently %v", e)
			verify = false
		}
	}

	if numElem == 0 {
		verify = false
	}
	return verify
}
