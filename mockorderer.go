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

import "github.com/hyperledger/fabric/protos/common"

// mockOrderer is a mock fabricsdk.Orderer
type mockOrderer struct {
	MockURL   string
	MockError error
}

// GetURL returns the mock URL of the mock Orderer
func (o *mockOrderer) GetURL() string {
	return o.MockURL
}

// SendBroadcast mocks sending a broadcast by sending nothing nowhere
func (o *mockOrderer) SendBroadcast(envelope *common.Envelope) error {
	return o.MockError
}
