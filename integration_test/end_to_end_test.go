package integration_test

import (
	"encoding/pem"
	"fmt"
	"strconv"
	"testing"
	"time"

	fabric_sdk "github.com/hyperledger/fabric-sdk-go"
	config "github.com/hyperledger/fabric-sdk-go/config"
	kvs "github.com/hyperledger/fabric-sdk-go/keyvaluestore"
	msp "github.com/hyperledger/fabric-sdk-go/msp"
	"github.com/hyperledger/fabric/bccsp"
	bccspFactory "github.com/hyperledger/fabric/bccsp/factory"

	"github.com/hyperledger/fabric/bccsp/sw"
	pb "github.com/hyperledger/fabric/protos/peer"
)

var chainCodeId = "end2end"
var chainId = "test_chainid"

func TestChainCodeInvoke(t *testing.T) {
	InitConfigForEndToEnd()
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
		msps, err := msp.NewMSPServices(config.GetMspUrl(), config.GetMspClientPath())
		if err != nil {
			t.Fatalf("NewFabricCOPServices return error: %v", err)
		}
		key, cert, err := msps.Enroll("testUser", "user1")
		keyPem, _ := pem.Decode(key)
		if err != nil {
			t.Fatalf("Enroll return error: %v", err)
		}
		user := fabric_sdk.NewUser("testUser")
		k, err := client.GetCryptoSuite().KeyImport(keyPem.Bytes, &bccsp.ECDSAPrivateKeyImportOpts{Temporary: false})
		if err != nil {
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

	// Get Query value before invoke
	value, err := getQueryValue(t, querychain)
	if err != nil {
		t.Fatalf("getQueryValue return error: %v", err)
	}
	fmt.Printf("*** QueryValue before invoke %s\n", value)

	err = invoke(t, invokechain)
	if err != nil {
		t.Fatalf("invoke return error: %v", err)
	}

	fmt.Println("need to wait now for the committer to catch up")
	time.Sleep(time.Second * 20)
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

func getQueryValue(t *testing.T, chain *fabric_sdk.Chain) (string, error) {

	var args []string
	args = append(args, "invoke")
	args = append(args, "query")
	args = append(args, "b")
	signedProposal, _, err := chain.CreateTransactionProposal(chainCodeId, chainId, args, true)
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

func invoke(t *testing.T, chain *fabric_sdk.Chain) error {

	var args []string
	args = append(args, "invoke")
	args = append(args, "move")
	args = append(args, "a")
	args = append(args, "b")
	args = append(args, "1")
	signedProposal, proposal, err := chain.CreateTransactionProposal(chainCodeId, chainId, args, true)
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
	return nil

}

func InitConfigForEndToEnd() {
	err := config.InitConfig("./test_resources/config/config_test.yaml")
	if err != nil {
		fmt.Println(err.Error())
	}
}
