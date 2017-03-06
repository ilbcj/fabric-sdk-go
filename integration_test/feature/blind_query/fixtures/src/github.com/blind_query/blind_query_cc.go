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

// BlindQuery Chaincode implementation
type BlindQuery struct {
}

// Init function
func (bq *BlindQuery) Init(stub shim.ChaincodeStubInterface) pb.Response {
	// Nothing to initialize
	// temp init:
	stub.PutState("12345", []byte("John Doe"))
	return shim.Success(nil)
}

// Invoke blind query chaincode
// args: function name, asset identifier
func (bq *BlindQuery) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	if function != "invoke" {
		return shim.Error("Invalid invoke function name. Expecting \"invoke\"")
	}
	if len(args) != 2 || args[0] == "" || args[1] == "" {
		return shim.Error(
			"Invalid invoke arguments. Expecting 2 arguments: asset identifier and event ID")
	}
	asset, err := stub.GetState(args[0])
	if err != nil {
		return shim.Error(err.Error())
	}
	// Create child transaction on messaging channel
	targetArgs := util.ToChaincodeArgs("invoke", "messaging", "lcescc", "1",
		"vp2:7051", "invoke", args[1], string(asset))
	childTransactionResp := stub.InvokeChaincode("ctscc", targetArgs, "")
	if childTransactionResp.Status != http.StatusOK {
		return shim.Error(fmt.Sprintf("Error from CTSCC: %s",
			childTransactionResp.Message))
	}

	return shim.Success(nil)
}

func main() {
	err := shim.Start(new(BlindQuery))
	if err != nil {
		fmt.Printf("Error starting ChannelResolver: %s", err)
	}
}
