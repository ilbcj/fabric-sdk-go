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
	"io"
	"strings"
	"time"

	config "github.com/hyperledger/fabric-sdk-go/config"
	"github.com/hyperledger/fabric/protos/common"
	ab "github.com/hyperledger/fabric/protos/orderer"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// Orderer ...
/**
 * The Orderer class represents a peer in the target blockchain network to which
 * HFC sends a block of transactions of endorsed proposals requiring ordering.
 *
 */
type Orderer interface {
	GetURL() string
	SendBroadcast(envelope *common.Envelope) error
}

type orderer struct {
	url            string
	grpcDialOption []grpc.DialOption
}

// CreateNewOrderer ...
/**
 * Returns a Orderer instance
 */
func CreateNewOrderer(url string) Orderer {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTimeout(time.Second*3))
	if config.IsTLSEnabled() {
		creds := credentials.NewClientTLSFromCert(config.GetTLSCACertPool(), config.GetTLSServerHostOverride())
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}
	return &orderer{url: url, grpcDialOption: opts}
}

// GetURL ...
/**
 * Get the Orderer url. Required property for the instance objects.
 * @returns {string} The address of the Orderer
 */
func (o *orderer) GetURL() string {
	return o.url
}

// SendBroadcast ...
/**
 * Send the created transaction to Orderer.
 */
func (o *orderer) SendBroadcast(envelope *common.Envelope) error {
	conn, err := grpc.Dial(o.url, o.grpcDialOption...)
	if err != nil {
		return err
	}
	defer conn.Close()

	broadcastStream, err := ab.NewAtomicBroadcastClient(conn).Broadcast(context.Background())
	if err != nil {
		return fmt.Errorf("Error Create NewAtomicBroadcastClient %v", err)
	}
	done := make(chan bool)
	var broadcastErr error
	go func() {
		for {
			broadcastResponse, err := broadcastStream.Recv()
			logger.Debugf("Orderer.broadcastStream - response:%v, error:%v\n", broadcastResponse, err)
			if err != nil {
				if strings.Contains(err.Error(), io.EOF.Error()) {
					done <- true
					return
				}
				broadcastErr = fmt.Errorf("Error broadcast respone : %v\n", err)
				continue
			}
			if broadcastResponse.Status != common.Status_SUCCESS {
				broadcastErr = fmt.Errorf("broadcast respone is not success : %v\n", broadcastResponse.Status)
			}
		}
	}()
	if err := broadcastStream.Send(envelope); err != nil {
		return fmt.Errorf("Failed to send a envelope to orderer: %v", err)
	}
	broadcastStream.CloseSend()
	<-done
	return broadcastErr
}
