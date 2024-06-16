# Timestamping Module

In this project, we aim to create a module that regularly updates the state of one chain onto another chain. In this case, the chains being LayerEdge and Bitcoin, we aim to post the latest state root of the LayerEdge chain onto the Bitcoin chain as an inscription which allows for later verification of the same.

## Services

### Writer Service
In the writer service, we listen to state proofs posted by the LayerEdge chain and write the latest proof onto the bitcoin chain every 1 hour. For this we perform the following tasks:
* Open a relayer connection to a bitcoin node service
* Open a websocket connection to the LayerEdge chain to listen to latest blocks
* Every one hour, take the latest confirmed state proof of the LayerEdge chain
* With the above state root, we create an inscription containing the state proof
* Finally we post the inscription onto the bitcoin chain using a commit reveal scheme 

The inscription posted has an 'ID' appended in order to be able to easily read the inscription later.

### Reader Service
In the Reader service, we provide a way to listen to inscriptions being posted onto the bitcoin chain. We do this by listening to all transactions and go through each one looking for an inscription that matches the "PROTOCOL_ID" that posted it. These are the following steps we take to get this done
* Open a websocket connection to a bitcoin node
* Listen to new blocks being produced
* Go through all the transaction once we recieve block data of the latest block
* Check if any transaction begins with PROTOCOL_ID
* Print out the inscription