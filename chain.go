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
	"sync"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric/bccsp"
	"github.com/hyperledger/fabric/core/crypto/primitives"
	msp "github.com/hyperledger/fabric/msp"
	"github.com/hyperledger/fabric/protos/common"
	pb "github.com/hyperledger/fabric/protos/peer"

	protos_utils "github.com/hyperledger/fabric/protos/utils"
	"github.com/op/go-logging"

	config "github.com/hyperledger/fabric-sdk-go/config"
)

var logger = logging.MustGetLogger("fabric_sdk_go")

// Chain ...
/**
 * The “Chain” object captures settings for a channel, which is created by
 * the orderers to isolate transactions delivery to peers participating on channel.
 * A chain must be initialized after it has been configured with the list of peers
 * and orderers. The initialization sends a CONFIGURATION transaction to the orderers
 * to create the specified channel and asks the peers to join that channel.
 *
 */
type Chain interface {
	GetName() string
	IsSecurityEnabled() bool
	GetTCertBatchSize() int
	SetTCertBatchSize(batchSize int)
	AddPeer(peer Peer)
	RemovePeer(peer Peer)
	GetPeers() []Peer
	AddOrderer(orderer Orderer)
	RemoveOrderer(orderer Orderer)
	GetOrderers() []Orderer
	InitializeChain() bool
	UpdateChain() bool
	IsReadonly() bool
	QueryInfo()
	QueryBlock(blockNumber int)
	QueryTransaction(transactionID int)
	CreateTransactionProposal(chaincodeName string, chainID string, args []string, sign bool, transientData map[string][]byte) (*pb.SignedProposal, *pb.Proposal, string, error)
	SendTransactionProposal(signedProposal *pb.SignedProposal, retry int) (map[string]*TransactionProposalResponse, error)
	CreateInvocationTransaction(chaincodeName string, chainID string, args []string, transientData map[string][]byte) (*common.Envelope, string, error)
	SendInvocationTransaction(envelope *common.Envelope) error
	CreateTransaction(proposal *pb.Proposal, resps []*pb.ProposalResponse) (*pb.Transaction, error)
	SendTransaction(proposal *pb.Proposal, tx *pb.Transaction) (map[string]*TransactionResponse, error)
}

type chain struct {
	name            string // Name of the chain is only meaningful to the client
	securityEnabled bool   // Security enabled flag
	peers           map[string]Peer
	tcertBatchSize  int // The number of tcerts to get in each batch
	orderers        map[string]Orderer
	clientContext   Client
}

// TransactionProposalResponse ...
/**
 * The TransactionProposalResponse result object returned from endorsers.
 */
type TransactionProposalResponse struct {
	Endorser         string
	ProposalResponse *pb.ProposalResponse
	Err              error
}

// TransactionResponse ...
/**
 * The TransactionProposalResponse result object returned from orderers.
 */
type TransactionResponse struct {
	Orderer string
	Err     error
}

// NewChain ...
/**
 * @param {string} name to identify different chain instances. The naming of chain instances
 * is enforced by the ordering service and must be unique within the blockchain network
 * @param {Client} clientContext An instance of {@link Client} that provides operational context
 * such as submitting User etc.
 */
func NewChain(name string, client Client) (Chain, error) {
	if name == "" {
		return nil, fmt.Errorf("Failed to create Chain. Missing requirement 'name' parameter.")
	}
	if client == nil {
		return nil, fmt.Errorf("Failed to create Chain. Missing requirement 'clientContext' parameter.")
	}
	p := make(map[string]Peer)
	o := make(map[string]Orderer)
	c := &chain{name: name, securityEnabled: config.IsSecurityEnabled(), peers: p,
		tcertBatchSize: config.TcertBatchSize(), orderers: o, clientContext: client}
	logger.Infof("Constructed Chain instance: %v", c)

	return c, nil
}

// GetName ...
/**
 * Get the chain name.
 * @returns {string} The name of the chain.
 */
func (c *chain) GetName() string {
	return c.name
}

// IsSecurityEnabled ...
/**
 * Determine if security is enabled.
 */
func (c *chain) IsSecurityEnabled() bool {
	return c.securityEnabled
}

// GetTCertBatchSize ...
/**
 * Get the tcert batch size.
 */
func (c *chain) GetTCertBatchSize() int {
	return c.tcertBatchSize
}

// SetTCertBatchSize ...
/**
 * Set the tcert batch size.
 */
func (c *chain) SetTCertBatchSize(batchSize int) {
	c.tcertBatchSize = batchSize
}

// AddPeer ...
/**
 * Add peer endpoint to chain.
 * @param {Peer} peer An instance of the Peer that has been initialized with URL,
 * TLC certificate, and enrollment certificate.
 */
func (c *chain) AddPeer(peer Peer) {
	c.peers[peer.GetURL()] = peer
}

// RemovePeer ...
/**
 * Remove peer endpoint from chain.
 * @param {Peer} peer An instance of the Peer.
 */
func (c *chain) RemovePeer(peer Peer) {
	delete(c.peers, peer.GetURL())
}

// GetPeers ...
/**
 * Get peers of a chain from local information.
 * @returns {[]Peer} The peer list on the chain.
 */
func (c *chain) GetPeers() []Peer {
	var peersArray []Peer
	for _, v := range c.peers {
		peersArray = append(peersArray, v)
	}
	return peersArray
}

// AddOrderer ...
/**
 * Add orderer endpoint to a chain object, this is a local-only operation.
 * A chain instance may choose to use a single orderer node, which will broadcast
 * requests to the rest of the orderer network. Or if the application does not trust
 * the orderer nodes, it can choose to use more than one by adding them to the chain instance.
 * All APIs concerning the orderer will broadcast to all orderers simultaneously.
 * @param {Orderer} orderer An instance of the Orderer class.
 */
func (c *chain) AddOrderer(orderer Orderer) {
	c.orderers[orderer.GetURL()] = orderer
}

// RemoveOrderer ...
/**
 * Remove orderer endpoint from a chain object, this is a local-only operation.
 * @param {Orderer} orderer An instance of the Orderer class.
 */
func (c *chain) RemoveOrderer(orderer Orderer) {
	delete(c.orderers, orderer.GetURL())

}

// GetOrderers ...
/**
 * Get orderers of a chain.
 */
func (c *chain) GetOrderers() []Orderer {
	var orderersArray []Orderer
	for _, v := range c.orderers {
		orderersArray = append(orderersArray, v)
	}
	return orderersArray
}

// InitializeChain ...
/**
 * Calls the orderer(s) to start building the new chain, which is a combination
 * of opening new message stream and connecting the list of participating peers.
 * This is a long-running process. Only one of the application instances needs
 * to call this method. Once the chain is successfully created, other application
 * instances only need to call getChain() to obtain the information about this chain.
 * @returns {bool} Whether the chain initialization process was successful.
 */
func (c *chain) InitializeChain() bool {
	return false
}

// UpdateChain ...
/**
 * Calls the orderer(s) to update an existing chain. This allows the addition and
 * deletion of Peer nodes to an existing chain, as well as the update of Peer
 * certificate information upon certificate renewals.
 * @returns {bool} Whether the chain update process was successful.
 */
func (c *chain) UpdateChain() bool {
	return false
}

// IsReadonly ...
/**
 * Get chain status to see if the underlying channel has been terminated,
 * making it a read-only chain, where information (transactions and states)
 * can be queried but no new transactions can be submitted.
 * @returns {bool} Is read-only, true or not.
 */
func (c *chain) IsReadonly() bool {
	return false //to do
}

// QueryInfo ...
/**
 * Queries for various useful information on the state of the Chain
 * (height, known peers).
 * @returns {object} With height, currently the only useful info.
 */
func (c *chain) QueryInfo() {
	//to do
}

// QueryBlock ...
/**
 * Queries the ledger for Block by block number.
 * @param {int} blockNumber The number which is the ID of the Block.
 * @returns {object} Object containing the block.
 */
func (c *chain) QueryBlock(blockNumber int) {
	//to do
}

// QueryTransaction ...
/**
 * Queries the ledger for Transaction by number.
 * @param {int} transactionID
 * @returns {object} Transaction information containing the transaction.
 */
func (c *chain) QueryTransaction(transactionID int) {
	//to do
}

// CreateTransactionProposal ...
/**
 * Create  a proposal for transaction. This involves assembling the proposal
 * with the data (chaincodeName, function to call, arguments, transient data, etc.) and signing it using the private key corresponding to the
 * ECert to sign.
 */
func (c *chain) CreateTransactionProposal(chaincodeName string, chainID string,
	args []string, sign bool, transientData map[string][]byte) (*pb.SignedProposal,
	*pb.Proposal, string, error) {

	argsArray := make([][]byte, len(args))
	for i, arg := range args {
		argsArray[i] = []byte(arg)
	}
	ccis := &pb.ChaincodeInvocationSpec{ChaincodeSpec: &pb.ChaincodeSpec{
		Type: pb.ChaincodeSpec_GOLANG, ChaincodeId: &pb.ChaincodeID{Name: chaincodeName},
		Input: &pb.ChaincodeInput{Args: argsArray}}}

	user, err := c.clientContext.GetUserContext("")
	if err != nil {
		return nil, nil, "", fmt.Errorf("GetUserContext return error: %s", err)
	}

	creatorID, err := getSerializedIdentity(user.GetEnrollmentCertificate())
	if err != nil {
		return nil, nil, "", err
	}
	// create a proposal from a ChaincodeInvocationSpec
	proposal, txID, err := protos_utils.CreateChaincodeProposalWithTransient(common.HeaderType_ENDORSER_TRANSACTION, chainID, ccis, creatorID, transientData)
	if err != nil {
		return nil, nil, "", fmt.Errorf("Could not create chaincode proposal, err %s", err)
	}

	proposalBytes, err := protos_utils.GetBytesProposal(proposal)
	if err != nil {
		return nil, nil, "", err
	}

	signature, err := c.signObjectWithKey(proposalBytes, user.GetPrivateKey(),
		&bccsp.SHAOpts{}, nil)
	if err != nil {
		return nil, nil, "", err
	}
	signedProposal := &pb.SignedProposal{ProposalBytes: proposalBytes, Signature: signature}
	return signedProposal, proposal, txID, nil
}

// SendTransactionProposal ...
// Send  the created proposal to peer for endorsement.
func (c *chain) SendTransactionProposal(signedProposal *pb.SignedProposal, retry int) (map[string]*TransactionProposalResponse, error) {
	if c.peers == nil || len(c.peers) == 0 {
		return nil, fmt.Errorf("peers is nil")
	}
	if signedProposal == nil {
		return nil, fmt.Errorf("signedProposal is nil")
	}

	var responseMtx sync.Mutex
	transactionProposalResponseMap := make(map[string]*TransactionProposalResponse)
	var wg sync.WaitGroup

	for _, p := range c.peers {
		wg.Add(1)
		go func(peer Peer) {
			defer wg.Done()
			var err error
			var proposalResponse *pb.ProposalResponse
			var transactionProposalResponse *TransactionProposalResponse
			logger.Debugf("Send ProposalRequest to peer :%s\n", peer.GetURL())
			if proposalResponse, err = peer.SendProposal(signedProposal); err != nil {
				logger.Debugf("Receive Error Response :%v\n", proposalResponse)
				transactionProposalResponse = &TransactionProposalResponse{peer.GetURL(), nil, fmt.Errorf("Error calling endorser '%s':  %s", peer.GetURL(), err)}
			} else {
				prp1, _ := protos_utils.GetProposalResponsePayload(proposalResponse.Payload)
				act1, _ := protos_utils.GetChaincodeAction(prp1.Extension)
				logger.Debugf("%s ProposalResponsePayload Extension ChaincodeAction Results\n%s\n", peer.GetURL(), string(act1.Results))

				logger.Debugf("Receive Proposal ChaincodeActionResponse :%v\n", proposalResponse)
				transactionProposalResponse = &TransactionProposalResponse{peer.GetURL(), proposalResponse, nil}
			}

			responseMtx.Lock()
			transactionProposalResponseMap[transactionProposalResponse.Endorser] = transactionProposalResponse
			responseMtx.Unlock()
		}(p)
	}
	wg.Wait()
	return transactionProposalResponseMap, nil
}

// CreateInvocationTransaction creates an invocation tranasaction that is broadcast
// to the ordering service. Its payload contains a signed proposal which is
// forwarded to the endorser server on the node to invoke chaincode
// arguments: It takes the arguments required to create a transaction proposal
// returns: transac envelope, error
func (c *chain) CreateInvocationTransaction(chaincodeName string, chainID string,
	args []string, transientData map[string][]byte) (*common.Envelope, string, error) {
	// Get user info and creator id
	user, err := c.clientContext.GetUserContext("")
	if err != nil {
		return nil, "", fmt.Errorf("GetUserContext returned error: %s", err)
	}

	creatorID, err := getSerializedIdentity(user.GetEnrollmentCertificate())
	if err != nil {
		return nil, "", err
	}

	// Create and marshal signed transaction proposal
	signedProposal, _, txID, err := c.CreateTransactionProposal(chaincodeName,
		chainID, args, true, transientData)
	if err != nil {
		return nil, "", err
	}
	signedProposalBytes, err := proto.Marshal(signedProposal)
	if err != nil {
		return nil, "", err
	}

	// generate a random nonce
	nonce, err := primitives.GetRandomNonce()
	if err != nil {
		return nil, "", err
	}

	signatureHeader := &common.SignatureHeader{Nonce: nonce, Creator: creatorID}
	signatureHeaderBytes, err := proto.Marshal(signatureHeader)
	if err != nil {
		return nil, "", err
	}

	// TODO: Change this header type once protobufs are merged into fabric
	channelHeader := &common.ChannelHeader{Type: 6,
		TxId:      txID,
		ChannelId: chainID}
	channelHeaderBytes, err := proto.Marshal(channelHeader)
	if err != nil {
		return nil, "", err
	}

	header := &common.Header{ChannelHeader: channelHeaderBytes,
		SignatureHeader: signatureHeaderBytes}

	payload := &common.Payload{
		Header: header,
		Data:   signedProposalBytes,
	}
	payloadBytes, err := proto.Marshal(payload)
	if err != nil {
		return nil, "", err
	}

	// Sign payload
	signature, err := c.signObjectWithKey(payloadBytes,
		user.GetPrivateKey(), &bccsp.SHAOpts{}, nil)
	if err != nil {
		return nil, "", err
	}

	envelope := &common.Envelope{
		Signature: signature,
		Payload:   payloadBytes,
	}

	return envelope, txID, nil
}

// SendInvocationTransaction broadcasts an invocation transaction through the
// ordering service. Transaction Invocation Listener System Chaincode(TILSCC)
// must be deployed on the peer to understand this transaction
// arguments: tranasaction
// returns: error
func (c *chain) SendInvocationTransaction(envelope *common.Envelope) error {
	var failureCount int
	transactionResponseMap, err := c.broadcastEnvelope(envelope)
	if err != nil {
		return err
	}
	for URL, resp := range transactionResponseMap {
		if resp.Err != nil {
			logger.Warningf("Could not broadcast to orderer: %s", URL)
			failureCount++
		}
	}
	// If all orderers returned error, the operation failed
	if failureCount == len(transactionResponseMap) {
		return fmt.Errorf("Broadcast failed: Received error from all configured orderers")
	}
	return nil
}

// CreateTransaction ...
/**
 * Create a transaction with proposal response, following the endorsement policy.
 */
func (c *chain) CreateTransaction(proposal *pb.Proposal, resps []*pb.ProposalResponse) (*pb.Transaction, error) {
	if len(resps) == 0 {
		return nil, fmt.Errorf("At least one proposal response is necessary")
	}

	// the original header
	hdr, err := protos_utils.GetHeader(proposal.Header)
	if err != nil {
		return nil, fmt.Errorf("Could not unmarshal the proposal header")
	}

	// the original payload
	pPayl, err := protos_utils.GetChaincodeProposalPayload(proposal.Payload)
	if err != nil {
		return nil, fmt.Errorf("Could not unmarshal the proposal payload")
	}

	// get header extensions so we have the visibility field
	hdrExt, err := protos_utils.GetChaincodeHeaderExtension(hdr)
	if err != nil {
		return nil, err
	}

	// This code is commented out because the ProposalResponsePayload Extension ChaincodeAction Results
	// return from endorsements is different so the compare will fail

	//	var a1 []byte
	//	for n, r := range resps {
	//		if n == 0 {
	//			a1 = r.Payload
	//			if r.Response.Status != 200 {
	//				return nil, fmt.Errorf("Proposal response was not successful, error code %d, msg %s", r.Response.Status, r.Response.Message)
	//			}
	//			continue
	//		}

	//		if bytes.Compare(a1, r.Payload) != 0 {
	//			return nil, fmt.Errorf("ProposalResponsePayloads do not match")
	//		}
	//	}

	for _, r := range resps {
		if r.Response.Status != 200 {
			return nil, fmt.Errorf("Proposal response was not successful, error code %d, msg %s", r.Response.Status, r.Response.Message)
		}
	}

	// fill endorsements
	endorsements := make([]*pb.Endorsement, len(resps))
	for n, r := range resps {
		endorsements[n] = r.Endorsement
	}
	// create ChaincodeEndorsedAction
	cea := &pb.ChaincodeEndorsedAction{ProposalResponsePayload: resps[0].Payload, Endorsements: endorsements}

	// obtain the bytes of the proposal payload that will go to the transaction
	propPayloadBytes, err := protos_utils.GetBytesProposalPayloadForTx(pPayl, hdrExt.PayloadVisibility)
	if err != nil {
		return nil, err
	}

	// serialize the chaincode action payload
	cap := &pb.ChaincodeActionPayload{ChaincodeProposalPayload: propPayloadBytes, Action: cea}
	capBytes, err := protos_utils.GetBytesChaincodeActionPayload(cap)
	if err != nil {
		return nil, err
	}

	// create a transaction
	taa := &pb.TransactionAction{Header: hdr.SignatureHeader, Payload: capBytes}
	taas := make([]*pb.TransactionAction, 1)
	taas[0] = taa
	tx := &pb.Transaction{Actions: taas}

	return tx, nil

}

// SendTransaction ...
/**
 * Send a transaction to the chain’s orderer service (one or more orderer endpoints) for consensus and committing to the ledger.
 * This call is asynchronous and the successful transaction commit is notified via a BLOCK or CHAINCODE event. This method must provide a mechanism for applications to attach event listeners to handle “transaction submitted”, “transaction complete” and “error” events.
 * Note that under the cover there are two different kinds of communications with the fabric backend that trigger different events to
 * be emitted back to the application’s handlers:
 * 1-)The grpc client with the orderer service uses a “regular” stateless HTTP connection in a request/response fashion with the “broadcast” call.
 * The method implementation should emit “transaction submitted” when a successful acknowledgement is received in the response,
 * or “error” when an error is received
 * 2-)The method implementation should also maintain a persistent connection with the Chain’s event source Peer as part of the
 * internal event hub mechanism in order to support the fabric events “BLOCK”, “CHAINCODE” and “TRANSACTION”.
 * These events should cause the method to emit “complete” or “error” events to the application.
 */
func (c *chain) SendTransaction(proposal *pb.Proposal, tx *pb.Transaction) (map[string]*TransactionResponse, error) {
	if c.orderers == nil || len(c.orderers) == 0 {
		return nil, fmt.Errorf("orderers is nil")
	}
	if proposal == nil {
		return nil, fmt.Errorf("proposal is nil")
	}
	if tx == nil {
		return nil, fmt.Errorf("Transaction is nil")
	}
	// the original header
	hdr, err := protos_utils.GetHeader(proposal.Header)
	if err != nil {
		return nil, fmt.Errorf("Could not unmarshal the proposal header")
	}
	// serialize the tx
	txBytes, err := protos_utils.GetBytesTransaction(tx)
	if err != nil {
		return nil, err
	}

	// create the payload
	payl := &common.Payload{Header: hdr, Data: txBytes}
	paylBytes, err := protos_utils.GetBytesPayload(payl)
	if err != nil {
		return nil, err
	}

	//Get user info
	user, err := c.clientContext.GetUserContext("")
	if err != nil {
		return nil, fmt.Errorf("GetUserContext return error: %s\n", err)
	}

	// sign payload
	signature, err := c.signObjectWithKey(paylBytes, user.GetPrivateKey(),
		&bccsp.SHAOpts{}, nil)
	if err != nil {
		return nil, err
	}
	// here's the envelope
	envelope := &common.Envelope{Payload: paylBytes, Signature: signature}

	transactionResponseMap, err := c.broadcastEnvelope(envelope)
	if err != nil {
		return nil, err
	}

	return transactionResponseMap, nil
}

//broadcastEnvelope will send the given envelope to each orderer
func (c *chain) broadcastEnvelope(envelope *common.Envelope) (map[string]*TransactionResponse, error) {
	// Check if orderers are defined
	if c.orderers == nil || len(c.orderers) == 0 {
		return nil, fmt.Errorf("orderers not set")
	}

	var responseMtx sync.Mutex
	transactionResponseMap := make(map[string]*TransactionResponse)
	var wg sync.WaitGroup

	for _, o := range c.orderers {
		wg.Add(1)
		go func(orderer Orderer) {
			defer wg.Done()
			var transactionResponse *TransactionResponse

			logger.Debugf("Broadcasting envelope to orderer :%s\n", orderer.GetURL())
			if err := orderer.SendBroadcast(envelope); err != nil {
				logger.Debugf("Receive Error Response from orderer :%v\n", err)
				transactionResponse = &TransactionResponse{orderer.GetURL(),
					fmt.Errorf("Error calling orderer '%s':  %s", orderer.GetURL(), err)}
			} else {
				logger.Debugf("Receive Success Response from orderer\n")
				transactionResponse = &TransactionResponse{orderer.GetURL(), nil}
			}

			responseMtx.Lock()
			transactionResponseMap[transactionResponse.Orderer] = transactionResponse
			responseMtx.Unlock()
		}(o)
	}
	wg.Wait()

	return transactionResponseMap, nil
}

// signObjectWithKey will sign the given object with the given key,
// hashOpts and signerOpts
func (c *chain) signObjectWithKey(object []byte, key bccsp.Key,
	hashOpts bccsp.HashOpts, signerOpts bccsp.SignerOpts) ([]byte, error) {
	cryptoSuite := c.clientContext.GetCryptoSuite()
	digest, err := cryptoSuite.Hash(object, hashOpts)
	if err != nil {
		return nil, err
	}
	signature, err := cryptoSuite.Sign(key, digest, signerOpts)
	if err != nil {
		return nil, err
	}

	return signature, nil
}

func getSerializedIdentity(userCertificate []byte) ([]byte, error) {
	serializedIdentity := &msp.SerializedIdentity{Mspid: config.GetMspID(),
		IdBytes: userCertificate}
	creatorID, err := proto.Marshal(serializedIdentity)
	if err != nil {
		return nil, fmt.Errorf("Could not Marshal serializedIdentity, err %s", err)
	}
	return creatorID, nil
}
