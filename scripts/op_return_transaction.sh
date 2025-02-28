#!/usr/bin/bash

### Script to create an OP RETURN transaction
###   Arg1 : Data <= 80 bytes

cli_path=$BTC_CLI_PATH
op_return_data=$1
num_inputs=1
rpc_endpoint="$2"
auth="$3"

### Validate data length
if [[ ${#op_return_data} -gt 80 ]]
then
	echo "Size check of data failed; ${#op_return_data} > 80" >> op_return_script.out
	exit 1
fi

exit_if_fail() {
	if [[ $? -ne 0 ]]
	then
		echo "error" $1 $? >> op_return_script.out
		exit 1
	fi
}

### Arg 1 : JSON
### Arg 2 : Field
find_field_from_json_dict() {
	jq -r ".$2" <<< "$1" 2>>2.out
}

### Arg 1 : JSON
### Arg 2 : Array index
### Arg 3 : Field
find_field_from_json_array() {
	jq -r ".[$1] | .$2" <<< "$3"
}

# unspent="`$cli_path listunspent > /tmp/unspent.out`"
unspent="`curl -u $auth --data-binary '{"jsonrpc": "1.0", "method": "listunspent", "params": [1, 9999999, [] , true, { "maximumCount": 10 }]}' -H 'content-type: text/plain;' $rpc_endpoint`"
exit_if_fail "unspent"

# $cli_path listunspent >> op_return_script.out
# cat /tmp/unspent.out >> op_return_script.out

# change_addr=`$cli_path getrawchangeaddress`
change_addr=`curl -u $auth --data-binary '{"jsonrpc": "1.0", "method": "getrawchangeaddress", "params": []}' -H 'content-type: text/plain;' $rpc_endpoint`
exit_if_fail "getrawchangeaddress"
change_addr=`find_field_from_json_dict $change_addr "result"`
echo "Raw change addr: $change_addr" >> op_return_script.out

## Get info from local wallet
inputs=""
total_amt=0

calculate_required() {
	echo $(bc <<<"(53 + $1 * 68 + $2)*0.00000001") | awk '{printf "%.8f", $0}'
}

data=`find_field_from_json_dict $unspent "result"`
# echo $data
while :
do
	echo "Trying num_inputs=$num_inputs" >> op_return_script.out
	index=$(( $num_inputs - 1 ))
	txid=`find_field_from_json_array $index "txid" "$data"`
	vout=`find_field_from_json_array $index "vout" "$data"`
	amt=`find_field_from_json_array  $index "amount" "$data"`
	exit_if_fail "getting values"
	echo $i $txid $vout $amt >> op_return_script.out
	total_amt=`echo $(bc <<<"$amt + $total_amt") | awk '{printf "%.8f", $0}'`
	if [[ -z $inputs ]]
	then
		inputs='''{"txid":"'$txid'","vout":'$vout'}'''
	else
		inputs=''''$inputs',{"txid":"'$txid'","vout":'$vout'}'''
	fi

	required=`calculate_required $num_inputs ${#op_return_data}`
	echo "Req: $required, Actual: $total_amt" >> op_return_script.out
	if (( $(echo "$total_amt >= $required" | bc -l) )); then
		break
	fi
	num_inputs=$(( $num_inputs + 1))
done

echo "Done: $total_amt" >> op_return_script.out

inputs="[$inputs]"
echo $total_amt $inputs >> op_return_script.out

## Create transaction
cost=$required
change=`echo $(bc <<<"$total_amt - $cost") | awk '{printf "%.8f", $0}'`
echo "Cost: $cost, Return Change: $change" >> op_return_script.out

echo "Inputs: $inputs" >> op_return_script.out
raw_tx_hex=$($cli_path -named createrawtransaction inputs=''''$inputs'''' outputs='''{ "data": "'$op_return_data'", "'$change_addr'": "'$change'" }''')
raw_tx_hex=`curl -u $auth --data-binary '{"jsonrpc": "1.0", "method": "createrawtransaction", "params": ["'''$inputs'''", "[{"data":"'''$op_return_data'''", "'''change_addr'''": '''$change'''}]"]}' -H 'content-type: text/plain;' $rpc_endpoint`
echo $raw_tx_hex
echo '"params": ["'''$inputs'''", "[{"data":"'''$op_return_data'''", "'''$change_addr'''": '''$change'''}]"]}'

echo "Raw tx hex: $raw_tx_hex" >> op_return_script.out

## Sign Transaction
signed_raw_tx_hex=$($cli_path signrawtransactionwithwallet $raw_tx_hex)
exit_if_fail "signing transaction"
signed_raw_tx_hex=`find_field_from_json_dict "$signed_raw_tx_hex" "hex"`
echo "Signed tx: Hex: " $signed_raw_tx_hex >> op_return_script.out
echo "Decoded tx: `$cli_path decoderawtransaction $signed_raw_tx_hex`" >> op_return_script.out

## Send Transaction
send_raw_tx=$($cli_path sendrawtransaction $signed_raw_tx_hex)
echo "Sent tx hex: $send_raw_tx" >> op_return_script.out
echo $send_raw_tx
