#!/bin/bash
set -e

PEER_BINARY=$GOPATH/src/github.com/hyperledger/fabric/build/bin/peer

cp -av ./src/github.com/blind_query $GOPATH/src/github.com
cp -av ./src/github.com/channel_resolver $GOPATH/src/github.com

cd $GOPATH/src/github.com/hyperledger/fabric

$PEER_BINARY channel create -c channel1 -o 0.0.0.0:7050

$PEER_BINARY  channel create -c channel2 -o 0.0.0.0:7050

$PEER_BINARY  channel create -c messaging -o 0.0.0.0:7050

CORE_PEER_ADDRESS=0.0.0.0:7051 $PEER_BINARY channel join -b channel1.block -o 0.0.0.0:7050

CORE_PEER_ADDRESS=0.0.0.0:7056 $PEER_BINARY channel join -b channel2.block -o 0.0.0.0:7050

CORE_PEER_ADDRESS=0.0.0.0:7059 $PEER_BINARY channel join -b messaging.block -o 0.0.0.0:7050

CORE_PEER_ADDRESS=0.0.0.0:7051 $PEER_BINARY chaincode install -n channelresolver -p github.com/channel_resolver -o 0.0.0.0:7050 -v v0

CORE_PEER_ADDRESS=0.0.0.0:7051 $PEER_BINARY chaincode instantiate -C testchainid -n channelresolver -p github.com/channel_resolver -v v0 -c '{"Args":["init"]}' -o 0.0.0.0:7050

CORE_PEER_ADDRESS=0.0.0.0:7051 $PEER_BINARY chaincode install -n blindquery -p github.com/blind_query -o 0.0.0.0:7050 -v v0

CORE_PEER_ADDRESS=0.0.0.0:7051 $PEER_BINARY chaincode instantiate -C channel1 -n blindquery -p github.com/blind_query -v v0 -c '{"Args":["init"]}' -o 0.0.0.0:7050

CORE_PEER_ADDRESS=0.0.0.0:7056 $PEER_BINARY chaincode install -n blindquery -p github.com/blind_query -o 0.0.0.0:7050 -v v0

CORE_PEER_ADDRESS=0.0.0.0:7056 $PEER_BINARY chaincode instantiate -C channel2 -n blindquery -p github.com/blind_query -v v0 -c '{"Args":["init"]}' -o 0.0.0.0:7050


rm -rf $GOPATH/src/github.com/blind_query
rm -rf $GOPATH/src/github.com/channel_resolver
rm -rf $GOPATH/src/github.com/hyperledger/fabric-sdk-go/integration_test/feature/blind_query/fixtures/*.block
