## Writer Service

This document aims to dive deeper into the writer service and how we utilise inscriptions to post data onto the bitcoin chain.
It will dive in deeper into the commit reveal scheme being deployed and what goes into making taproot scripts that allow inscriptions.

The write process mainly achieves the following three tasks
* Create a taproot address
* Commit the transaction by signing it
* Reveal the transaction to the world

Lets dive in deeper

### Creating a taproot address

With the proof that needs to be inscripted, we create a taproot address. It contains a Taproot script with a single leaf containing the spend path with the script. The script being `<embedded data> OP_DROP <pubkey> OP_CHECKSIG`.

Lets understand the above statement a little more in detail. What does it mean for a bitcoin address to "contain" a Taproot script? What is this "leaf" and how does this leaf contain the "spend path"?

An address in bitcoin can be of many types, the main ones being P2PKH(Pay to(2) Public Key Hash, called P2WPKH in Segwit) and P2SH(Pay to Script Hash, called P2WSH in Segwit). In the above case, we are referring to the script address i.e. P2SH. There are many ways to build a script hash and can be further studied here, [Introduction to Bitcion Scripts](https://github.com/BlockchainCommons/Learning-Bitcoin-from-the-Command-Line/blob/master/09_0_Introducing_Bitcoin_Scripts.md).

>Anyone can send bitcoins to any Bitcoin address; those funds can only be spent if they fulfill certain conditions defined in the script. It governs how the next person can spend the sent bitcoins. When senders include a script in a transaction is called PubKey Script, also known as locking script. The receiver of the sent bitcoins will generate a signature script, also known as an unlocking script, which is a collection of data parameters that satisfy a PubKey Script. Signature scripts are called scriptSig in code. - [CoinCodeCap](https://coincodecap.com/bitcoin-taproot)

Here our end goal is to inscribe the Layer 2 chains latest state onto the bitcoin chain by inscribing via a script, the latest state proof(also called Merkle Proof/State Root) of the chain. To build the script, we utilise two private keys, one to sign the reveal transaction and one used for tweaking.

As mentioned in the quote above, the receiver and sender are both controlled by the LayerEdge Node. The locking script is a taproot script that uses a private key to do something called as "tweaking". To know more about how taproot scripts work, you can check out [this beautiful explanation](https://bitcoin.stackexchange.com/questions/111098/what-is-the-script-assembly-and-execution-in-p2tr-spend-spend-from-taproot) on Stack Exchange.

Using these three values, the two private keys and the data(in this case the state root), we generate a unique address that contains all this information combined.

### Commiting the transaction
Once we get the unique address that contains a locking script with the embedded state root data, we now create a transaction of 0.0001BTC to that address. CommitTx commits an output to the given taproot address, such that the output is only spendable by posting the embedded data on chain, as part of the script satisfying the tapscript spend path that commits to the data. This basically means once the transaction has been revealed to the world by posting it on the blockchain, we embedded data would also be posted on chain along with it, thus inscribing the state root.

### Revealing the transaction
RevealTx spends the output from the commit transaction and as part of the script satisfying the tapscript spend path, posts the embedded data on chain. As mentioned above, we spend the bitcoin by satisfying the conditions required to unlock the script we created to embed the data on chain.