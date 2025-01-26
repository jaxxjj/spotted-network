package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/crypto"
)

func main() {
	// Create directory for keys if not exists
	keysDir := "keys"
	if err := os.MkdirAll(keysDir, 0755); err != nil {
		panic(err)
	}

	// Clean up any existing operator*.key.json files
	files, err := filepath.Glob(filepath.Join(keysDir, "operator*.key.json"))
	if err != nil {
		panic(err)
	}
	for _, f := range files {
		if err := os.Remove(f); err != nil {
			panic(err)
		}
	}

	password := "testpassword" // Use this password for all test accounts

	// Generate 3 different keystores
	for i := 1; i <= 3; i++ {
		// Generate new private key
		privateKey, err := crypto.GenerateKey()
		if err != nil {
			panic(err)
		}

		// Create keystore
		ks := keystore.NewKeyStore(keysDir, keystore.StandardScryptN, keystore.StandardScryptP)
		
		// Import the private key into the keystore
		_, err = ks.ImportECDSA(privateKey, password)
		if err != nil {
			panic(err)
		}

		// Get all files and find the one that was just created (should be the only UTC-- file)
		files, err := filepath.Glob(filepath.Join(keysDir, "*"))
		if err != nil {
			panic(err)
		}
		
		var utcFile string
		for _, f := range files {
			if strings.HasPrefix(filepath.Base(f), "UTC--") {
				utcFile = f
				break
			}
		}
		
		if utcFile == "" {
			panic("Could not find UTC file")
		}
		
		// Rename to operator{i}.key.json
		newPath := filepath.Join(keysDir, fmt.Sprintf("operator%d.key.json", i))
		if err := os.Rename(utcFile, newPath); err != nil {
			panic(err)
		}

		// Print the address for use in EmitEvent script
		address := crypto.PubkeyToAddress(privateKey.PublicKey)
		signingKey := crypto.PubkeyToAddress(privateKey.PublicKey) // Using address as signing key for simplicity
		fmt.Printf("Operator %d:\n", i)
		fmt.Printf("Address: %s\n", address.Hex())
		fmt.Printf("Signing Key: %s\n", signingKey.Hex())
		fmt.Println("-------------------")
	}
} 