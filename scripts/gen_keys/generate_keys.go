package main

import (
	"crypto/ed25519"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/crypto"
)

func main() {
	// Add flag for key type
	keyType := flag.String("type", "ecdsa", "Key type to generate: 'ecdsa' or 'ed25519'")
	flag.Parse()

	// Create directory for keys if not exists
	keysDir := "keys"
	if err := os.MkdirAll(keysDir, 0755); err != nil {
		panic(err)
	}

	// Clean up existing files based on key type
	pattern := "operator*.key.json"
	if *keyType == "ed25519" {
		pattern = "operator*.ed25519.key"
	}
	
	files, err := filepath.Glob(filepath.Join(keysDir, pattern))
	if err != nil {
		panic(err)
	}
	for _, f := range files {
		if err := os.Remove(f); err != nil {
			panic(err)
		}
	}

	// Generate 3 different keys
	for i := 1; i <= 3; i++ {
		if *keyType == "ed25519" {
			generateEd25519Key(i, keysDir)
		} else {
			generateECDSAKey(i, keysDir)
		}
	}
}

func generateEd25519Key(index int, keysDir string) {
	// Generate Ed25519 keypair
	pubKey, privKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		panic(err)
	}

	// Save private key to file
	filename := filepath.Join(keysDir, fmt.Sprintf("operator%d.ed25519.key", index))
	if err := os.WriteFile(filename, privKey, 0600); err != nil {
		panic(err)
	}

	// Print the public key for reference
	fmt.Printf("Operator %d (Ed25519):\n", index)
	fmt.Printf("Public Key: %s\n", hex.EncodeToString(pubKey))
	fmt.Printf("Private Key saved to: %s\n", filename)
	fmt.Println("-------------------")
}

func generateECDSAKey(index int, keysDir string) {
	password := "testpassword"
	
	// Generate new private key
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		panic(err)
	}

	// Create keystore and import key
	ks := keystore.NewKeyStore(keysDir, keystore.StandardScryptN, keystore.StandardScryptP)
	_, err = ks.ImportECDSA(privateKey, password)
	if err != nil {
		panic(err)
	}

	// Find and rename the UTC file
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
	
	newPath := filepath.Join(keysDir, fmt.Sprintf("operator%d.key.json", index))
	if err := os.Rename(utcFile, newPath); err != nil {
		panic(err)
	}

	address := crypto.PubkeyToAddress(privateKey.PublicKey)
	fmt.Printf("Operator %d (ECDSA):\n", index)
	fmt.Printf("Address: %s\n", address.Hex())
	fmt.Println("-------------------")
} 