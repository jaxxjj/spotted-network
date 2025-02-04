package main

import (
	"crypto/ecdsa"
	"fmt"

	"github.com/ethereum/go-ethereum/crypto"
)

func main() {
	privateKey, err := crypto.HexToECDSA("1234567890123456789012345678901234567890123456789012345678901234")
	if err != nil {
		fmt.Printf("Error generating private key: %v\n", err)
		return
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		fmt.Println("Error casting public key to ECDSA")
		return
	}

	address := crypto.PubkeyToAddress(*publicKeyECDSA)
	
	fmt.Printf("Private Key: 0x1234567890123456789012345678901234567890123456789012345678901234\n")
	fmt.Printf("Address: %s\n", address.Hex())
} 