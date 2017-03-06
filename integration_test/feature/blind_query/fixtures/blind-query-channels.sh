#!/bin/bash
set -e

PEER_BINARY=$GOPATH/src/github.com/hyperledger/fabric/build/bin/peer

cp -av ./src/github.com/blind_query $GOPATH/src/github.com
cp -av ./src/github.com/channel_resolver $GOPATH/src/github.com

CORE_PEER_COMMITTER_LEDGER_ORDERER=0.0.0.0:7050 $PEER_BINARY channel create -c channel1

CORE_PEER_COMMITTER_LEDGER_ORDERER=0.0.0.0:7050 $PEER_BINARY  channel create -c channel2

CORE_PEER_COMMITTER_LEDGER_ORDERER=0.0.0.0:7050 $PEER_BINARY  channel create -c messaging

CORE_PEER_COMMITTER_LEDGER_ORDERER=0.0.0.0:7050 CORE_PEER_ADDRESS=0.0.0.0:7051 $PEER_BINARY channel join -b channel1.block

CORE_PEER_COMMITTER_LEDGER_ORDERER=0.0.0.0:7050 CORE_PEER_ADDRESS=0.0.0.0:7056 $PEER_BINARY channel join -b channel2.block

CORE_PEER_COMMITTER_LEDGER_ORDERER=0.0.0.0:7050 CORE_PEER_ADDRESS=0.0.0.0:7059 $PEER_BINARY channel join -b messaging.block

CORE_PEER_ADDRESS=0.0.0.0:7051 CORE_PEER_COMMITTER_LEDGER_ORDERER=0.0.0.0:7050 $PEER_BINARY chaincode deploy -C testchainid -n channelresolver -p github.com/channel_resolver -c '{"Args":["init"]}'

CORE_PEER_ADDRESS=0.0.0.0:7051 CORE_PEER_COMMITTER_LEDGER_ORDERER=0.0.0.0:7050 $PEER_BINARY chaincode deploy -C channel1 -n blindquery -p github.com/blind_query -c '{"Args":["init"]}'

CORE_PEER_ADDRESS=0.0.0.0:7056 CORE_PEER_COMMITTER_LEDGER_ORDERER=0.0.0.0:7050 $PEER_BINARY chaincode deploy -C channel2 -n blindquery -p github.com/blind_query -c '{"Args":["init"]}'

rm -rf $GOPATH/src/github.com/blind_query
rm -rf $GOPATH/src/github.com/channel_resolver
rm -rf *.block
