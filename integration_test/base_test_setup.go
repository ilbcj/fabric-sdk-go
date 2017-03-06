/*
Copyright SecureKey Technologies Inc. All Rights Reserved.


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at


      http://www.apache.org/licenses/LICENSE-2.0


Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package integration

import (
	"encoding/pem"
	"fmt"
	"testing"

	"github.com/hyperledger/fabric-sdk-go/config"
	"github.com/hyperledger/fabric-sdk-go/events"
	"github.com/hyperledger/fabric-sdk-go/msp"
	"github.com/hyperledger/fabric/bccsp"
	"github.com/hyperledger/fabric/bccsp/sw"

	fabric_sdk "github.com/hyperledger/fabric-sdk-go"
	kvs "github.com/hyperledger/fabric-sdk-go/keyvaluestore"
	bccspFactory "github.com/hyperledger/fabric/bccsp/factory"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// BaseTestSetup is an interface used by the integration tests
// it performs setup activities like user enrollment, chain creation,
// crypto suite selection, and event hub initialization
type BaseTestSetup interface {
	GetChains(t *testing.T) (*fabric_sdk.Chain, *fabric_sdk.Chain)
	GetEventHub(t *testing.T, interestedEvents []*pb.Interest) *events.EventHub
}

// BaseSetupImpl implementation of BaseTestSetup
type BaseSetupImpl struct {
}

//
func (setup *BaseSetupImpl) GetChains(t *testing.T) (fabric_sdk.Chain, fabric_sdk.Chain) {
	client := fabric_sdk.NewClient()
	ks := &sw.FileBasedKeyStore{}
	if err := ks.Init(nil, config.GetKeyStorePath(), false); err != nil {
		t.Fatalf("Failed initializing key store [%s]", err)
	}

	cryptoSuite, err := bccspFactory.GetBCCSP(&bccspFactory.SwOpts{Ephemeral_: true, SecLevel: config.GetSecurityLevel(),
		HashFamily: config.GetSecurityAlgorithm(), KeyStore: ks})
	if err != nil {
		t.Fatalf("Failed getting ephemeral software-based BCCSP [%s]", err)
	}
	client.SetCryptoSuite(cryptoSuite)
	stateStore, err := kvs.CreateNewFileKeyValueStore("/tmp/enroll_user")
	if err != nil {
		t.Fatalf("CreateNewFileKeyValueStore return error[%s]", err)
	}
	client.SetStateStore(stateStore)
	user, err := client.GetUserContext("testUser")
	if err != nil {
		t.Fatalf("client.GetUserContext return error: %v", err)
	}
	if user == nil {
		msps, err1 := msp.NewMSPServices(config.GetMspClientPath())
		if err1 != nil {
			t.Fatalf("NewFabricCOPServices return error: %v", err)
		}
		key, cert, err1 := msps.Enroll("testUser", "user1")
		keyPem, _ := pem.Decode(key)
		if err1 != nil {
			t.Fatalf("Enroll return error: %v", err1)
		}
		user := fabric_sdk.NewUser("testUser")
		k, err1 := client.GetCryptoSuite().KeyImport(keyPem.Bytes, &bccsp.ECDSAPrivateKeyImportOpts{Temporary: false})
		if err1 != nil {
			t.Fatalf("KeyImport return error: %v", err)
		}
		user.SetPrivateKey(k)
		user.SetEnrollmentCertificate(cert)
		err = client.SetUserContext(user, false)
		if err != nil {
			t.Fatalf("client.SetUserContext return error: %v", err)
		}
	}

	querychain, err := client.NewChain("querychain")
	if err != nil {
		t.Fatalf("NewChain return error: %v", err)
	}

	for _, p := range config.GetPeersConfig() {
		endorser := fabric_sdk.CreateNewPeer(fmt.Sprintf("%s:%s", p.Host, p.Port))
		querychain.AddPeer(endorser)
		break
	}

	invokechain, err := client.NewChain("invokechain")
	if err != nil {
		t.Fatalf("NewChain return error: %v", err)
	}
	orderer := fabric_sdk.CreateNewOrderer(fmt.Sprintf("%s:%s", config.GetOrdererHost(), config.GetOrdererPort()))
	invokechain.AddOrderer(orderer)

	for _, p := range config.GetPeersConfig() {
		endorser := fabric_sdk.CreateNewPeer(fmt.Sprintf("%s:%s", p.Host, p.Port))
		invokechain.AddPeer(endorser)
	}

	return querychain, invokechain

}

// GetEventHub initilizes the event hub
func (setup *BaseSetupImpl) GetEventHub(t *testing.T,
	interestedEvents []*pb.Interest) events.EventHub {
	eventHub := events.NewEventHub()
	foundEventHub := false
	for _, p := range config.GetPeersConfig() {
		if p.EventHost != "" && p.EventPort != "" {
			eventHub.SetPeerAddr(fmt.Sprintf("%s:%s", p.EventHost, p.EventPort))
			foundEventHub = true
			break
		}
	}

	if !foundEventHub {
		t.Fatalf("No EventHub configuration found")
	}

	if interestedEvents != nil {
		eventHub.SetInterestedEvents(interestedEvents)
	}

	if err := eventHub.Connect(); err != nil {
		t.Fatalf("Failed eventHub.Connect() [%s]", err)
	}

	return eventHub
}
