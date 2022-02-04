// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"flag"
	"github.com/onosproject/onos-lib-go/pkg/certs"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	"github.com/onosproject/onos-mlb/pkg/manager"
)

var log = logging.GetLogger("main")

func main() {
	caPath := flag.String("caPath", "", "path to CA certificate")
	keyPath := flag.String("keyPath", "", "path to client private key")
	certPath := flag.String("certPath", "", "path to client certificate")
	configPath := flag.String("configPath", "/etc/onos/config/config.json", "path to config.json file")
	e2tEndpoint := flag.String("e2tEndpoint", "onos-e2t:5150", "E2T service endpoint")
	uenibEndpoint := flag.String("uenibEndpoint", "onos-uenib:5150", "UENIB service endpoint")
	ricActionID := flag.Int("ricActionID", 10, "RIC Action ID in E2 message")
	grpcPort := flag.Int("grpcPort", 5150, "grpc Port number")
	overloadThreshold := flag.Int("overloadThreshold", 100, "Overload threshold")
	targetLoadThreshold := flag.Int("targetLoadThreshold", 0, "Target load threshold")

	flag.Parse()

	_, err := certs.HandleCertPaths(*caPath, *keyPath, *certPath, true)
	if err != nil {
		log.Fatal(err)
	}

	log.Info("Starting onos-mlb")

	appConfParams := manager.AppParameters{
		CAPath:              *caPath,
		KeyPath:             *keyPath,
		CertPath:            *certPath,
		ConfigPath:          *configPath,
		E2tEndpoint:         *e2tEndpoint,
		UENIBEndpoint:       *uenibEndpoint,
		GRPCPort:            *grpcPort,
		RicActionID:         int32(*ricActionID),
		OverloadThreshold:   *overloadThreshold,
		TargetLoadThreshold: *targetLoadThreshold,
	}

	done := make(chan bool)

	appMgr := manager.NewManager(appConfParams)

	err = appMgr.Start()
	if err != nil {
		log.Fatal(err)
	}

	<-done
}
