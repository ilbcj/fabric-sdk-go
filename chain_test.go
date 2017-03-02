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

package fabricsdk

import (
	"fmt"
	"net"
	"testing"

	"google.golang.org/grpc"

	"github.com/golang/protobuf/proto"
	mocks "github.com/hyperledger/fabric-sdk-go/mocks"
	cb "github.com/hyperledger/fabric/protos/common"
	protoOrderer "github.com/hyperledger/fabric/protos/orderer"
	pb "github.com/hyperledger/fabric/protos/peer"
	"github.com/spf13/viper"
)

var testPayload = &cb.Envelope{
	Payload: []byte("test payload"),
}
var testAddress = "0.0.0.0:19875"

type MockBroadcastServer struct {
	test *testing.T
	srv  protoOrderer.AtomicBroadcast_DeliverServer
}

func (m *MockBroadcastServer) Broadcast(server protoOrderer.AtomicBroadcast_BroadcastServer) error {
	req, err := server.Recv()
	if err != nil {
		m.test.Fatalf("Error at ordering service: %s", err)
	}
	if string(req.Payload) != "test payload" {
		server.Send(&protoOrderer.BroadcastResponse{
			Status: cb.Status_BAD_REQUEST,
		})
	} else {
		server.Send(&protoOrderer.BroadcastResponse{
			Status: cb.Status_SUCCESS,
		})
	}
	return nil
}

func (m *MockBroadcastServer) Deliver(server protoOrderer.AtomicBroadcast_DeliverServer) error {
	// Not implemented
	return nil
}

func TestChainMethods(t *testing.T) {
	client := NewClient()
	chain, err := NewChain("testChain", client)
	if err != nil {
		t.Fatalf("NewChain return error[%s]", err)
	}
	if chain.GetName() != "testChain" {
		t.Fatalf("NewChain create wrong chain")
	}

	_, err = NewChain("", client)
	if err == nil {
		t.Fatalf("NewChain didn't return error")
	}
	if err.Error() != "Failed to create Chain. Missing requirement 'name' parameter." {
		t.Fatalf("NewChain didn't return right error")
	}

	_, err = NewChain("testChain", nil)
	if err == nil {
		t.Fatalf("NewChain didn't return error")
	}
	if err.Error() != "Failed to create Chain. Missing requirement 'clientContext' parameter." {
		t.Fatalf("NewChain didn't return right error")
	}

}

func TestCreateInvocationTransaction(t *testing.T) {
	chain, err := setupTestChain()
	if err != nil {
		t.Fatalf("Failed to create chain: %s", err)
	}
	envelope, err := chain.CreateInvocationTransaction("testChaincode",
		"testChain", []string{"test"}, "123", nil)
	if err != nil {
		t.Fatalf("Got unexpected error from CreateInvocationTransaction: %s", err)
	}
	if string(envelope.Signature) != "testSignature" {
		t.Fatalf("InvocationTransaction had wrong signature")
	}
	// Unmarshal payload
	payload := &cb.Payload{}
	if err := proto.Unmarshal(envelope.Payload, payload); err != nil {
		t.Fatalf("Invalid payload")
	}
	// Unmarshal proposal
	proposal := &pb.SignedProposal{}
	if err := proto.Unmarshal(payload.Data, proposal); err != nil {
		t.Fatalf("Invalid proposal")
	}
	// Verify header type
	// TODO: Change header type once protobuf changes are merged in
	if payload.Header.GetChainHeader().Type != 6 {
		t.Fatalf("Invalid header type")
	}
}

func TestSendInvocationTransaction(t *testing.T) {
	viper.Set("client.tls.enabled", false)
	startMockServer(t)
	chain, err := setupTestChain()
	if err != nil {
		t.Fatalf("Failed to create chain: %s", err)
	}

	orderer := CreateNewOrderer(testAddress)
	chain.AddOrderer(orderer)
	// Test happy flow
	err = chain.SendInvocationTransaction(testPayload)
	if err != nil {
		t.Fatalf("SendInvocationTransaction return error: %s", err)
	}
	// test with invalid payload
	testPayload = &cb.Envelope{
		Payload: []byte("test invalid payload"),
	}
	err = chain.SendInvocationTransaction(testPayload)
	if err.Error() != "Broadcast failed: Received error from all configured orderers" {
		t.Fatalf("Expected invocation transaction broadcast to fail with invalid payload")
	}
	// test with no orderer
	chain.RemoveOrderer(orderer)
	err = chain.SendInvocationTransaction(testPayload)
	if err == nil {
		t.Fatalf("Expected error with no orderer configured")
	}
	// test with invalid orderer configuration
	chain.AddOrderer(CreateNewOrderer("0.0.0.0:29675"))
	if err == nil {
		t.Fatalf("Expected error with invalid orderer configured")
	}
}

func TestConcurrentPeers(t *testing.T) {
	const numPeers = 10000
	chain, err := setupMassiveTestChain(numPeers)
	if err != nil {
		t.Fatalf("Failed to create massive chain: %s", err)
	}

	result, err := chain.SendTransactionProposal(&pb.SignedProposal{}, 1)
	if err != nil {
		t.Fatalf("SendTransactionProposal return error: %s", err)
	}

	if len(result) != numPeers {
		t.Error("SendTransactionProposal returned an unexpected amount of responses")
	}
}

func startMockServer(t *testing.T) {
	grpcServer := grpc.NewServer()
	lis, err := net.Listen("tcp", testAddress)
	broadcastServer := &MockBroadcastServer{
		test: t,
	}
	protoOrderer.RegisterAtomicBroadcastServer(grpcServer, broadcastServer)
	if err != nil {
		t.Fatalf("Error starting test server %s", err)
	}
	fmt.Printf("Starting test server\n")
	go grpcServer.Serve(lis)
}

func setupTestChain() (Chain, error) {
	client := NewClient()
	user := NewUser("test")
	cryptoSuite := &mocks.MockCryptoSuite{}
	client.SetUserContext(user, true)
	client.SetCryptoSuite(cryptoSuite)
	return NewChain("testChain", client)
}

func setupMassiveTestChain(numberOfPeers int) (Chain, error) {
	chain, error := setupTestChain()
	if error != nil {
		return chain, error
	}

	for i := 0; i < numberOfPeers; i++ {
		peer := mockPeer{fmt.Sprintf("MockPeer%d", i), fmt.Sprintf("http://mock%d.peers.r.us", i), []string{}, nil}
		chain.AddPeer(&peer)
	}

	return chain, error
}
