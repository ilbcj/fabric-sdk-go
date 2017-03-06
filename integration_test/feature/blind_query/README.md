## Blind Query

This is an end to end test of the blind query functionality. Data is retrieved from an unknown ledger using the following functionality:
- [Request Broadcast](https://jira.hyperledger.org/browse/FAB-2440) to deliver a query request to all sub-ledgers on a global channel.
- [Peer Configuration Reader](https://jira.hyperledger.org/browse/FAB-2585) to resolve the 'home channel' of the organization receiving the request.
- [Child Transactions](https://jira.hyperledger.org/browse/FAB-2438) to create a transaction on the messaging channel that the client is listening on.
- [Light Chaincode Events](https://jira.hyperledger.org/browse/FAB-2567) to deliver the payload to the client.

All of these system chaincodes must be compiled in your fabric image for this test to run. For instructions on how to set up these system chaincodes, refer to the fabric-extension [read me](https://github.com/securekey/fabric-extension)

### Test Setup

In order to run this test, we require a complex channel setup consisting of three channels (channel1, channel2, and messaging) each comprising of a single node. As the Golang SDK client does not currently support channel creation and deployment, we use the peer CLI and a shell script to set this up. There are four steps in this setup:

###### Build Fabric

With Hyperledger Fabric installed on your GOPATH, run this command inside the Fabric directory:
```
$ make docker && make peer
```

**NOTE:** The system chaincodes mentioned above must be imported into Fabric before building the image. Additionally, we only support the commit levels specified in the README at the root of this project

###### Build and run fabric-ca

In a new terminal:
```
$ cd $GOPATH/src/github.com/hyperledger/fabric-ca
$ make fabric-ca
$ bin/fabric-ca server start --address "" -ca testdata/ec.pem  -ca-key testdata/ec-key.pem -config testdata/testconfig.json
```

###### Run nodes, create channels, deploy chaincode

In a new terminal window, inside the fixtures directory in this package, run:
```
$ docker-compose up --force-recreate
```
In another terminal window, inside the fixtures directory, run:
```
sh blind-query-channels.sh
```
This script will take a couple of minutes to run.

NOTE for MacOS: This script freezes on mac but has been tested inside the Vagrant VM provided by Fabric

###### Go Test

Run the integration test:
```
$ go test
```
