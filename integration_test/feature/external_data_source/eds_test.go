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

package main

import (
	"fmt"
	"testing"

	fabric_sdk "github.com/hyperledger/fabric-sdk-go"

	config "github.com/hyperledger/fabric-sdk-go/config"

	integration "github.com/hyperledger/fabric-sdk-go/integration_test"
)

var chainId = "testchainid"

func TestEDS(t *testing.T) {

	initConfig(t)

	testSetup := integration.BaseSetupImpl{}
	// Get invoke chain
	_, invokechain := testSetup.GetChains(t)

	// Invoke external data source
	invokeExternalDataSource(t, invokechain)

}

func invokeExternalDataSource(t *testing.T, chain fabric_sdk.Chain) {

	var args []string
	args = append(args, "invoke")
	args = append(args, "1234")
	args = append(args, "https://172.17.0.1:8443/hello")

	signedProposal, _, _, err := chain.CreateTransactionProposal("edsscc", chainId, args, true, nil)
	if err != nil {
		t.Fatalf("CreateTransactionProposal return error: %v", err)
	}

	transactionProposalResponse, err := chain.SendTransactionProposal(signedProposal, 0)
	if err != nil {
		t.Fatalf("SendTransactionProposal return error: %v", err)
	}

	for _, v := range transactionProposalResponse {
		if v.Err != nil {
			t.Fatalf("Endorser %s return error: %v", v.Endorser, v.Err)
		}
		response := string(v.ProposalResponse.GetResponse().Payload)
		fmt.Printf("Response: %s", response)
	}

}

func initConfig(t *testing.T) {
	err := config.InitConfig("./test_resources/config/config_test.yaml")
	if err != nil {
		t.Fatalf("Failed to read configuration: %v", err)
	}
}
