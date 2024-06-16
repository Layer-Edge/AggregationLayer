## Reader Service

This document aims to dive in deeper into how the Reader service functions.

The reader service is fairly simple as compared to the writer service. These are the main steps followed by the service.
* Get block data
* Process each transaction
* Check if transaction matches script template; and
* Check if data starts with corresponding PROTOCOL_ID used when writing

Block data is read from a bitcoin full node using socket connections. A reader service opens a socket connection to a "ZeroMQ Port" defined in the full node to read new blocks.

More information on how to setup ZeroMQ in a bitcoin node can be found here -> [Block and Transaction Broadcasting with ZeroMQ](https://github.com/bitcoin/bitcoin/blob/master/doc/zmq.md#block-and-transaction-broadcasting-with-zeromq)

Each transaction is processed and the "push data" of the first witness of each transaction is extracted. This data section might contain the inscription we're looking for. Hence, we process each data to check whether the script matches our script template and if the data starts with the corresponding PROTOCOL_ID.