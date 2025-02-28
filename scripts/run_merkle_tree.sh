#!/bin/bash

if [ "$#" -ne 1 ]; then
  echo "Usage: $0 <comma separated data>" >> merkle_treegen.out
  exit 1
fi

REPO_PATH="$RISC0_PATH"
SERVER_ENDPOINT="http://localhost:8080"
INPUT=$1

if [ ! -d "$REPO_PATH" ]; then
  echo "Error: Directory $REPO_PATH does not exist." >> merkle_treegen.out
  exit 1
fi

cd "$REPO_PATH" || exit 1

check_server_running() {
  while ! curl -s $SERVER_ENDPOINT/ > /dev/null; do
    echo "Waiting for the Rust server to start" >> merkle_treegen.out
    sleep 10
  done
  echo "Rust server is running." >> merkle_treegen.out
}

echo "Starting Rust server" >> merkle_treegen.out
cargo run &> server.log &
SERVER_PID=$!

check_server_running

# echo "Enter data for the Merkle Tree (comma-separated, e.g., data1,data2,data3):"
# read -r MERKLE_DATA
IFS=',' read -r -a DATA_ARRAY <<< "$INPUT"

DATA_JSON=$(printf '"%s",' "${DATA_ARRAY[@]}" | sed 's/,$//')

echo "Inserting data into the Merkle Tree" >> merkle_treegen.out
INSERT_RESPONSE=$(curl -s -X POST $SERVER_ENDPOINT/process \
  -H "Content-Type: application/json" \
  -d "{
        \"operation\": \"insert\",
        \"data\": [$DATA_JSON],
        \"proof_request\": null,
        \"proof\": null
      }")
MERKLE_ROOT=$(echo "$INSERT_RESPONSE" | jq -r '.root')
echo "Merkle Root: $MERKLE_ROOT" >> merkle_treegen.out

for LEAF_DATA in "${DATA_ARRAY[@]}"; do
  echo "Generating Merkle proof for leaf data: $LEAF_DATA" >> merkle_treegen.out
  PROOF_RESPONSE=$(curl -s -X POST $SERVER_ENDPOINT/process \
    -H "Content-Type: application/json" \
    -d "{
          \"operation\": \"prove\",
          \"data\": [$DATA_JSON],
          \"proof_request\": \"$LEAF_DATA\",
          \"proof\": null
        }")
  MERKLE_PROOF=$(echo "$PROOF_RESPONSE" | jq -c '.proof')
  echo "Merkle Proof for $LEAF_DATA: $MERKLE_PROOF" >> merkle_treegen.out

  echo "Verifying Merkle proof for leaf data: $LEAF_DATA" >> merkle_treegen.out
  VERIFY_RESPONSE=$(curl -s -X POST $SERVER_ENDPOINT/process \
    -H "Content-Type: application/json" \
    -d "{
          \"operation\": \"verify\",
          \"data\": [$DATA_JSON],
          \"proof_request\": null,
          \"proof\": $MERKLE_PROOF
        }")
  VERIFIED=$(echo "$VERIFY_RESPONSE" | jq -r '.verified')
  echo "Proof Verified for $LEAF_DATA: $VERIFIED" >> merkle_treegen.out
done

echo "Stopping Rust server" >> merkle_treegen.out
kill "$SERVER_PID"

echo "Done" >> merkle_treegen.out
echo "Final Merkle Root: $MERKLE_ROOT" >> merkle_treegen.out
echo "Final Merkle Root: $MERKLE_ROOT"
