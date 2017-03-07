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
	"fmt"
	"strconv"
	"testing"
	"time"

	fabric_sdk "github.com/hyperledger/fabric-sdk-go"
	events "github.com/hyperledger/fabric-sdk-go/events"

	config "github.com/hyperledger/fabric-sdk-go/config"

	pb "github.com/hyperledger/fabric/protos/peer"
)

var chainCodeId = "end2end"
var chainId = "testchainid"

func TestChainCodeInvoke(t *testing.T) {
	InitConfigForEndToEnd()
	testSetup := BaseSetupImpl{}

	eventHub := testSetup.GetEventHub(t, nil)
	querychain, invokechain := testSetup.GetChains(t)

	// Get Query value before invoke
	value, err := getQueryValue(t, querychain)
	if err != nil {
		t.Fatalf("getQueryValue return error: %v", err)
	}
	fmt.Printf("*** QueryValue before invoke %s\n", value)

	err = invoke(t, invokechain, eventHub)
	if err != nil {
		t.Fatalf("invoke return error: %v", err)
	}

	valueAfterInvoke, err := getQueryValue(t, querychain)
	if err != nil {
		t.Errorf("getQueryValue return error: %v", err)
		return
	}
	fmt.Printf("*** QueryValue after invoke %s\n", valueAfterInvoke)

	valueInt, _ := strconv.Atoi(value)
	valueInt = valueInt + 1
	valueAfterInvokeInt, _ := strconv.Atoi(valueAfterInvoke)
	if valueInt != valueAfterInvokeInt {
		t.Fatalf("SendTransaction didn't change the QueryValue")

	}

}

func getQueryValue(t *testing.T, chain fabric_sdk.Chain) (string, error) {

	var args []string
	args = append(args, "invoke")
	args = append(args, "query")
	args = append(args, "b")

	signedProposal, _, _, err := chain.CreateTransactionProposal(chainCodeId, chainId, args, true, nil)
	if err != nil {
		return "", fmt.Errorf("SendTransactionProposal return error: %v", err)
	}
	transactionProposalResponse, err := chain.SendTransactionProposal(signedProposal, 0)
	if err != nil {
		return "", fmt.Errorf("SendTransactionProposal return error: %v", err)
	}

	for _, v := range transactionProposalResponse {
		if v.Err != nil {
			return "", fmt.Errorf("Endorser %s return error: %v", v.Endorser, v.Err)
		}
		return string(v.ProposalResponse.GetResponse().Payload), nil
	}
	return "", nil
}

func invoke(t *testing.T, chain fabric_sdk.Chain, eventHub events.EventHub) error {

	var args []string
	args = append(args, "invoke")
	args = append(args, "move")
	args = append(args, "a")
	args = append(args, "b")
	args = append(args, "1")

	signedProposal, proposal, txID, err := chain.CreateTransactionProposal(chainCodeId, chainId, args, true, nil)
	if err != nil {
		return fmt.Errorf("SendTransactionProposal return error: %v", err)
	}
	transactionProposalResponse, err := chain.SendTransactionProposal(signedProposal, 0)
	if err != nil {
		return fmt.Errorf("SendTransactionProposal return error: %v", err)
	}

	var proposalResponses []*pb.ProposalResponse
	for _, v := range transactionProposalResponse {
		if v.Err != nil {
			return fmt.Errorf("Endorser %s return error: %v", v.Endorser, v.Err)
		}
		proposalResponses = append(proposalResponses, v.ProposalResponse)
		fmt.Printf("Endorser '%s' return ProposalResponse:%v\n", v.Endorser, v.ProposalResponse.GetResponse())
	}

	tx, err := chain.CreateTransaction(proposal, proposalResponses)
	if err != nil {
		return fmt.Errorf("CreateTransaction return error: %v", err)

	}
	transactionResponse, err := chain.SendTransaction(proposal, tx)
	if err != nil {
		return fmt.Errorf("SendTransaction return error: %v", err)

	}
	for _, v := range transactionResponse {
		if v.Err != nil {
			return fmt.Errorf("Orderer %s return error: %v", v.Orderer, v.Err)
		}
	}
	done := make(chan bool)
	eventHub.RegisterTxEvent(txID, func(txId string, err error) {
		fmt.Printf("receive success event for txid(%s)\n", txId)
		done <- true
	})

	select {
	case <-done:
	case <-time.After(time.Second * 20):
		return fmt.Errorf("Didn't receive block event for txid(%s)\n", txID)
	}
	return nil

}

func InitConfigForEndToEnd() {
	err := config.InitConfig("./test_resources/config/config_test.yaml")
	if err != nil {
		fmt.Println(err.Error())
	}
}
