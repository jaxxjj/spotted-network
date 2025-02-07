package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ethereum/go-ethereum/crypto"
	p2pcrypto "github.com/libp2p/go-libp2p/core/crypto"
	"golang.org/x/crypto/sha3"
)

func main() {
	keyType := flag.String("type", "ed25519", "Type of key to generate (ed25519 or ecdsa)")
	flag.Parse()

	// Create keys directory if it doesn't exist
	keysDir := "./keys"
	if err := os.MkdirAll(keysDir, 0755); err != nil {
		fmt.Printf("Failed to create keys directory: %v\n", err)
		os.Exit(1)
	}

	// Generate keys for three operators
	for i := 1; i <= 3; i++ {
		switch *keyType {
		case "ed25519":
			if err := generateEd25519Key(i, keysDir); err != nil {
				fmt.Printf("Failed to generate Ed25519 key for operator%d: %v\n", i, err)
				os.Exit(1)
			}
		case "ecdsa":
			if err := generateECDSAKey(i, keysDir); err != nil {
				fmt.Printf("Failed to generate ECDSA key for operator%d: %v\n", i, err)
				os.Exit(1)
			}
		default:
			fmt.Printf("Unsupported key type: %s\n", *keyType)
			os.Exit(1)
		}
	}
}

func generateEd25519Key(operatorNum int, keysDir string) error {
	// Generate Ed25519 key pair
	privKey, pubKey, err := p2pcrypto.GenerateEd25519Key(nil)
	if err != nil {
		return fmt.Errorf("failed to generate Ed25519 key: %v", err)
	}

	// Get key bytes
	keyBytes, err := p2pcrypto.MarshalPrivateKey(privKey)
	if err != nil {
		return fmt.Errorf("failed to marshal private key: %v", err)
	}

	// Generate address from public key
	address, err := generateAddress(pubKey)
	if err != nil {
		return fmt.Errorf("failed to generate address: %v", err)
	}

	// Save private key
	filename := filepath.Join(keysDir, fmt.Sprintf("operator%d.ed25519.key", operatorNum))
	if err := os.WriteFile(filename, keyBytes, 0600); err != nil {
		return fmt.Errorf("failed to write key file: %v", err)
	}

	fmt.Printf("Generated Ed25519 key for operator%d:\n", operatorNum)
	fmt.Printf("  Key file: %s\n", filename)
	fmt.Printf("  Address: %s\n\n", address)

	return nil
}

func generateECDSAKey(operatorNum int, keysDir string) error {
	// Generate ECDSA key
	key, err := crypto.GenerateKey()
	if err != nil {
		return fmt.Errorf("failed to generate ECDSA key: %v", err)
	}

	// Get address
	address := crypto.PubkeyToAddress(key.PublicKey).Hex()

	// Save private key
	filename := filepath.Join(keysDir, fmt.Sprintf("operator%d.ecdsa.key", operatorNum))
	keyBytes := crypto.FromECDSA(key)
	if err := os.WriteFile(filename, keyBytes, 0600); err != nil {
		return fmt.Errorf("failed to write key file: %v", err)
	}

	fmt.Printf("Generated ECDSA key for operator%d:\n", operatorNum)
	fmt.Printf("  Key file: %s\n", filename)
	fmt.Printf("  Address: %s\n\n", address)

	return nil
}

func generateAddress(pubKey p2pcrypto.PubKey) (string, error) {
	// Marshal public key
	pubKeyBytes, err := p2pcrypto.MarshalPublicKey(pubKey)
	if err != nil {
		return "", fmt.Errorf("failed to marshal public key: %v", err)
	}

	// Generate Keccak256 hash
	hasher := sha3.NewLegacyKeccak256()
	hasher.Write(pubKeyBytes)
	hash := hasher.Sum(nil)

	// Get last 20 bytes as address
	address := hash[len(hash)-20:]

	// Convert to hex with 0x prefix
	return "0x" + hex.EncodeToString(address), nil
}
