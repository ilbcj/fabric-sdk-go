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
	"time"

	fabric_sdk "github.com/hyperledger/fabric-sdk-go"

	config "github.com/hyperledger/fabric-sdk-go/config"
	"github.com/hyperledger/fabric/common/util"

	integration "github.com/hyperledger/fabric-sdk-go/integration_test"
	pb "github.com/hyperledger/fabric/protos/peer"
)

var chainId = "testchainid"

func TestLCE(t *testing.T) {

	initConfigForLCE(t)
	testSetup := integration.BaseSetupImpl{}
	// Get invoke chain
	_, invokechain := testSetup.GetChains(t)

	// Generate transaction id
	txId := util.GenerateUUID()

	// Light Chaincode Event system chaincode id
	lcesccId := "lcescc"

	// Create an interest in LCE event with event id that equals generated transaction id
	interestedEvents := []*pb.Interest{{EventType: pb.EventType_CHAINCODE,
		RegInfo: &pb.Interest_ChaincodeRegInfo{
			ChaincodeRegInfo: &pb.ChaincodeReg{
				ChaincodeId: lcesccId,
				EventName:   txId}}}}

	// Register interest with event hub
	eventHub := testSetup.GetEventHub(t, interestedEvents)

	defer eventHub.Disconnected(nil)

	done := make(chan bool)

	// Register callback for specific LCE
	lce := eventHub.RegisterChaincodeEvent(lcesccId, txId, func(ce *pb.ChaincodeEvent) {
		fmt.Printf("Received LCE event ( %s ): \n%v\n", time.Now().Format(time.RFC850), ce)
		done <- true
	})

	defer eventHub.UnregisterChaincodeEvent(lce)

	// Generate LCE with eventId=txId
	invokeLCEWithTxID(t, invokechain, lcesccId, txId)

	select {
	case <-done:
	case <-time.After(time.Second * 20):
		t.Fatalf("Did NOT receive LCE for eventId(%s)\n", txId)
	}

}

func invokeLCEWithTxID(t *testing.T, chain fabric_sdk.Chain, lcesccId string, txId string) {

	var args []string
	args = append(args, "invoke")
	args = append(args, txId)
	args = append(args, "Test Payload")

	signedProposal, _, _, err := chain.CreateTransactionProposal(lcesccId, chainId, args, true, nil)
	if err != nil {
		t.Fatalf("SendTransactionProposal return error: %v", err)
	}

	fmt.Printf("Send LCE event ( %s ): \n", time.Now().Format(time.RFC850))

	if _, err := chain.SendTransactionProposal(signedProposal, 0); err != nil {
		t.Fatalf("SendTransactionProposal return error: %v", err)
	}

}

func initConfigForLCE(t *testing.T) {
	err := config.InitConfig("./test_resources/config/config_test.yaml")
	if err != nil {
		t.Fatalf("Failed to read configuration: %v", err)
	}
}
