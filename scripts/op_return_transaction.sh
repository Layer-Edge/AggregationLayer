#!/usr/bin/bash

### Script to create an OP RETURN transaction
###   Arg1 : Data <= 80 bytes

### Validate data length
op_return_data=$1
if [[ ${#op_return_data} -gt 80 ]]
then
	exit
fi

exit_if_fail() {
	if [[ $? -ne 0 ]]
	then
		>&2 echo "error" $?
		exit 1
	fi
}

cli_path=$BTC_CLI_PATH
unspent="`$cli_path listunspent > /tmp/unspent.out`"
exit_if_fail
# echo $unspent
## Remove newline
unspent=`echo ${unspent/\n/}`
# echo $unspent

change_addr=`$cli_path getrawchangeaddress`
exit_if_fail

### Arg 1 : JSON
### Arg 2 : Field
find_field_from_json_dict() {
	jq -r ".$2" <<< "$1" 2>>2.out
}

### Arg 1 : JSON
### Arg 2 : Field
find_field_from_json_array() {
	jq -r ".[0] | .$2" < "/tmp/unspent.out"
}

## Get info from local wallet

txid=`find_field_from_json_array "$unspent" "txid"`
vout=`find_field_from_json_array "$unspent" "vout"`
amt=`find_field_from_json_array "$unspent" "amount"`

exit_if_fail

# echo $txid $vout $amt


## Create transaction
cost=0.0001
change=$(bc <<<"scale=10;$amt - $cost")

raw_tx_hex=$($cli_path -named createrawtransaction inputs='''[ { "txid": "'$txid'", "vout": '$vout' } ]''' outputs='''{ "data": "'$op_return_data'", "'$change_addr'": "'$change'" }''')


## Sign Transaction
signed_raw_tx_hex=$($cli_path signrawtransactionwithwallet $raw_tx_hex)
exit_if_fail
signed_raw_tx_hex=`find_field_from_json_dict "$signed_raw_tx_hex" "hex"`
# echo "Hex: " $signed_raw_tx_hex
# echo "Decoded: " "`$cli_path decoderawtransaction $signed_raw_tx_hex`"


## Send Transaction
send_raw_tx=$($cli_path sendrawtransaction $signed_raw_tx_hex)
echo $send_raw_tx
