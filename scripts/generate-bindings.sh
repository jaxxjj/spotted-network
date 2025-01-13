#!/bin/bash

# Directory containing the ABIs
ABI_DIR="abi"
# Output directory for the bindings
OUT_DIR="pkg/common/contracts/bindings"
# Common types package
TYPES_PKG="github.com/galxe/spotted-network/pkg/common/contracts/types"

# Generate bindings for each contract
for abi in "$ABI_DIR"/*.json; do
    filename=$(basename "$abi")
    contract="${filename%.*}"
    
    # Extract just the ABI portion from the JSON file
    jq .abi "$abi" > "$ABI_DIR/${contract}_abi.json"
    
    # Generate bindings with type overrides
    abigen --abi "$ABI_DIR/${contract}_abi.json" \
           --pkg bindings \
           --type "${contract^}" \
           --out "$OUT_DIR/${contract}.go" \
           --type-substitutes "ISignatureUtilsSignatureWithSaltAndExpiry=$TYPES_PKG.SignatureWithSaltAndExpiry" \
           --type-substitutes "Quorum=$TYPES_PKG.QuorumConfig" \
           --type-substitutes "StrategyParams=$TYPES_PKG.Strategy"
           
    # Clean up temporary ABI file
    rm "$ABI_DIR/${contract}_abi.json"
done 