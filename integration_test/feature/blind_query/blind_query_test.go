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

	"github.com/hyperledger/fabric-sdk-go/config"
	"github.com/hyperledger/fabric/common/util"

	integration "github.com/hyperledger/fabric-sdk-go/integration_test"
	pb "github.com/hyperledger/fabric/protos/peer"
)

func TestBlindQuery(t *testing.T) {
	initializeClientConfig(t)
	testSetup := integration.BaseSetupImpl{}

	// Generate transaction id
	eventID := util.GenerateUUID()

	// Light Chaincode Event system chaincode id
	var lcesccID = "lcescc"

	// Create an interest in LCE event with event id that equals generated transaction id
	interestedEvents := []*pb.Interest{{EventType: pb.EventType_CHAINCODE,
		RegInfo: &pb.Interest_ChaincodeRegInfo{
			ChaincodeRegInfo: &pb.ChaincodeReg{
				ChaincodeId: lcesccID,
				EventName:   eventID}}}}

	// Register interest with event hub
	eventHub := testSetup.GetEventHub(t, interestedEvents)
	defer eventHub.Disconnected(nil)

	done := make(chan bool)

	// Register callback for specific LCE
	lce := eventHub.RegisterChaincodeEvent(lcesccID, eventID, func(ce *pb.ChaincodeEvent) {
		fmt.Printf("Received LCE event ( %s ): \n%v\n", time.Now().Format(time.RFC850), ce)
		done <- true
	})
	defer eventHub.UnregisterChaincodeEvent(lce)

	// Create and send invocation transaction
	_, invokeChain := testSetup.GetChains(t)
	// Invoke channelresolver cc on the global channel testchainid and request asset
	invokeTxn, _, err := invokeChain.CreateInvocationTransaction("channelresolver",
		"testchainid", []string{"invoke", "12345", eventID}, nil)
	if err != nil {
		fmt.Printf("Error creating invocation transaction: %s", err)
		t.FailNow()
	}
	err = invokeChain.SendInvocationTransaction(invokeTxn)
	if err != nil {
		fmt.Printf("Error sending invocation transaction: %s", err)
		t.FailNow()
	}

	select {
	case <-done:
	case <-time.After(time.Second * 10):
		t.Fatalf("Did NOT receive LCE for eventId(%s)\n", eventID)
	}
}

func initializeClientConfig(t *testing.T) {
	err := config.InitConfig("./test_resources/config/config_test.yaml")
	if err != nil {
		fmt.Println(err.Error())
		t.FailNow()
	}
}
