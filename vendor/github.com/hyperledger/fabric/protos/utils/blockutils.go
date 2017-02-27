/*
Copyright IBM Corp. 2016 All Rights Reserved.

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

package utils

import (
	"fmt"

	"github.com/golang/protobuf/proto"
	cb "github.com/hyperledger/fabric/protos/common"
)

// GetChainIDFromBlock returns chain ID in the block
func GetChainIDFromBlock(block *cb.Block) (string, error) {
	if block.Data == nil || block.Data.Data == nil || len(block.Data.Data) == 0 {
		return "", fmt.Errorf("Failed to find chain ID because the block is empty.")
	}
	var err error
	envelope := &cb.Envelope{}
	if err = proto.Unmarshal(block.Data.Data[0], envelope); err != nil {
		return "", fmt.Errorf("Error reconstructing envelope(%s)", err)
	}
	payload := &cb.Payload{}
	if err = proto.Unmarshal(envelope.Payload, payload); err != nil {
		return "", fmt.Errorf("Error reconstructing payload(%s)", err)
	}

	return payload.Header.ChainHeader.ChainID, nil
}

// GetMetadataFromBlock retrieves metadata at the specified index.
func GetMetadataFromBlock(block *cb.Block, index cb.BlockMetadataIndex) (*cb.Metadata, error) {
	md := &cb.Metadata{}
	err := proto.Unmarshal(block.Metadata.Metadata[index], md)
	if err != nil {
		return nil, err
	}
	return md, nil
}

// GetMetadataFromBlockOrPanic retrieves metadata at the specified index, or panics on error.
func GetMetadataFromBlockOrPanic(block *cb.Block, index cb.BlockMetadataIndex) *cb.Metadata {
	md, err := GetMetadataFromBlock(block, index)
	if err != nil {
		panic(err)
	}
	return md
}

// GetLastConfigurationIndexFromBlock retrieves the index of the last configuration block as encoded in the block metadata
func GetLastConfigurationIndexFromBlock(block *cb.Block) (uint64, error) {
	md, err := GetMetadataFromBlock(block, cb.BlockMetadataIndex_LAST_CONFIGURATION)
	if err != nil {
		return 0, err
	}
	lc := &cb.LastConfiguration{}
	err = proto.Unmarshal(md.Value, lc)
	if err != nil {
		return 0, err
	}
	return lc.Index, nil
}

// GetLastConfigurationIndexFromBlockOrPanic retrieves the index of the last configuration block as encoded in the block metadata, or panics on error.
func GetLastConfigurationIndexFromBlockOrPanic(block *cb.Block) uint64 {
	index, err := GetLastConfigurationIndexFromBlock(block)
	if err != nil {
		panic(err)
	}
	return index
}

// GetBlockFromBlockBytes marshals the bytes into Block
func GetBlockFromBlockBytes(blockBytes []byte) (*cb.Block, error) {
	block := &cb.Block{}
	err := proto.Unmarshal(blockBytes, block)
	return block, err
}

// CopyBlockMetadata copies metadata from one block into another
func CopyBlockMetadata(src *cb.Block, dst *cb.Block) {
	dst.Metadata = src.Metadata
	// Once copied initialize with rest of the
	// required metadata positions.
	InitBlockMetadata(dst)
}

// InitBlockMetadata copies metadata from one block into another
func InitBlockMetadata(block *cb.Block) {
	if block.Metadata == nil {
		block.Metadata = &cb.BlockMetadata{Metadata: [][]byte{[]byte{}, []byte{}, []byte{}}}
	} else if len(block.Metadata.Metadata) < int(cb.BlockMetadataIndex_TRANSACTIONS_FILTER+1) {
		for i := int(len(block.Metadata.Metadata)); i <= int(cb.BlockMetadataIndex_TRANSACTIONS_FILTER); i++ {
			block.Metadata.Metadata = append(block.Metadata.Metadata, []byte{})
		}
	}
}
