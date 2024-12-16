bitcoin-da:
===========
This package provides a reader / writer interface to bitcoin.

# Timestamping Module

In this project, we aim to create a module that regularly updates the state of one chain onto another chain. In this case, the chains being LayerEdge and Bitcoin, we aim to post the latest state root of the LayerEdge chain onto the Bitcoin chain as an inscription which allows for later verification of the same.

## Installation

- [Install ZeroMQ](https://zeromq.org/download/)
- [Install CZMQ](https://zeromq.org/languages/c/)


## Config

```
Usage: ./bitcoin-da [options]
Options:
  -c
        config path (default: config.yml)
  -w
        Run Writer Service (default: reader service)
```

- [default config file](./config.yml) defines the main configuration to run the project
- set custom config path with flag `-c`. eg. ` -c ./custom-config.yml`
- by default project runs with the reader service 
- switch to writer service by config change `enable-writer: true` or use flag `-w`

Services:
========

### [Writer](./docs/writer-service.md)

`go build && ./bitcoin-da -w` or `go run . -w` or `make build`

In the writer service, we listen to state proofs posted by the LayerEdge chain and write the latest proof onto the bitcoin chain every 1 hour. For this we perform the following tasks:
* Open a relayer connection to a bitcoin node service
* Open a websocket connection to the LayerEdge chain to listen to latest blocks
* Every one hour, take the latest confirmed state proof of the LayerEdge chain
* With the above state root, we create an inscription containing the state proof
* Finally we post the inscription onto the bitcoin chain using a commit reveal scheme 

The inscription posted has an 'ID' appended in order to be able to easily read the inscription later.

A commit transaction containing a taproot with one leaf script

    OP_FALSE
    OP_IF
      "roll" marker
      <embedded data>
    OP_ENDIF
    <pubkey>
    OP_CHECKSIG

is used to create a new bech32m address and is sent an output.


A reveal transaction then posts the embedded data on chain and spends the
commit output.


### [Reader](./docs/reader-service.md)

`go build && ./bitcoin-da`  or  `go run .` or `make run`

In the Reader service, we provide a way to listen to inscriptions being posted onto the bitcoin chain. We do this by listening to all transactions and go through each one looking for an inscription that matches the "PROTOCOL_ID" that posted it. These are the following steps we take to get this done
* Open a websocket connection to a bitcoin node
* Listen to new blocks being produced
* Go through all the transaction once we recieve block data of the latest block
* Check if any transaction begins with PROTOCOL_ID
* Print out the inscription

The address of the reveal transaction is implicity used as a namespace.


Clients may call listunspent on the reveal transaction address to get a list of
transactions and read the embedded data from the first witness input.

Spec:
=====

For more details, [read the spec](./spec.md)
