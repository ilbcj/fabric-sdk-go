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
	"net/http"

	"github.com/hyperledger/fabric/common/util"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// ChannelResolver Chaincode implementation
type ChannelResolver struct {
}

// Init function
func (cr *ChannelResolver) Init(stub shim.ChaincodeStubInterface) pb.Response {
	// Nothing to initialize
	return shim.Success(nil)
}

// Invoke channel resolver chaincode.
// args: function name, asset identifier
func (cr *ChannelResolver) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	if function != "invoke" {
		return shim.Error("Invalid invoke function name. Expecting \"invoke\"")
	}

	if len(args) != 2 || args[0] == "" || args[1] == "" {
		return shim.Error(
			"Invalid invoke args. Expecting two arguments: asset identifier, event ID")
	}

	// Read home channel name from pcrscc
	pcrsccArgs := util.ToChaincodeArgs("invoke", "chaincode.system.config.homechannel")
	configReaderResponse := stub.InvokeChaincode("pcrscc", pcrsccArgs, "")
	if configReaderResponse.Status != http.StatusOK {
		err := fmt.Sprintf("Error from Peer Configuration Reader SCC(pcrscc): %s",
			configReaderResponse.Message)
		return shim.Error(err)
	} else if configReaderResponse.Payload == nil {
		return shim.Error("Peer Configuration Reader SCC returned nil")
	}

	// Invoke blind query chaincode
	homeChannel := string(configReaderResponse.Payload)
	blindQueryArgs := util.ToChaincodeArgs("invoke", args[0], args[1])
	blindQueryResponse := stub.InvokeChaincode("blindquery",
		blindQueryArgs, homeChannel)
	if blindQueryResponse.Status != http.StatusOK {
		err := fmt.Sprintf("Error from Blind Query Chaincode: %s",
			blindQueryResponse.Message)
		return shim.Error(err)
	}
	// Success, blind query response will be returned via events
	return shim.Success(nil)
}

func main() {
	err := shim.Start(new(ChannelResolver))
	if err != nil {
		fmt.Printf("Error starting ChannelResolver: %s", err)
	}
}
