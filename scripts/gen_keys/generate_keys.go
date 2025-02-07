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
	ecdsaKeysDir := "./keys/signing"
	if err := os.MkdirAll(ecdsaKeysDir, 0755); err != nil {
		fmt.Printf("Failed to create keys directory: %v\n", err)
		os.Exit(1)
	}

	// Generate keys for three operators
	for i := 1; i <= 3; i++ {
		switch *keyType {
		case "ed25519":
			if err := generateP2PKey(i); err != nil {
				fmt.Printf("Failed to generate P2P key for operator%d: %v\n", i, err)
				os.Exit(1)
			}
		case "ecdsa":
			if err := generateECDSAKey(i, ecdsaKeysDir); err != nil {
				fmt.Printf("Failed to generate ECDSA key for operator%d: %v\n", i, err)
				os.Exit(1)
			}
		default:
			fmt.Printf("Unsupported key type: %s\n", *keyType)
			os.Exit(1)
		}
	}
}

func generateP2PKey(operatorNum int) error {
	// 生成Ed25519密钥对
	privKey, pubKey, err := p2pcrypto.GenerateEd25519Key(nil)
	if err != nil {
		return fmt.Errorf("failed to generate P2P key: %v", err)
	}

	// 序列化并编码私钥
	privKeyBytes, err := p2pcrypto.MarshalPrivateKey(privKey)
	if err != nil {
		return fmt.Errorf("failed to marshal private key: %v", err)
	}
	privKeyStr := p2pcrypto.ConfigEncodeKey(privKeyBytes)

	// 序列化并编码公钥
	pubKeyBytes, err := p2pcrypto.MarshalPublicKey(pubKey)
	if err != nil {
		return fmt.Errorf("failed to marshal public key: %v", err)
	}
	pubKeyStr := p2pcrypto.ConfigEncodeKey(pubKeyBytes)

	// 生成地址
	address, err := generateAddress(pubKey)
	if err != nil {
		return fmt.Errorf("failed to generate address: %v", err)
	}

	// 打印密钥信息
	fmt.Printf("Generated P2P key for operator%d:\n", operatorNum)
	fmt.Printf("  Private key (base64): %s\n", privKeyStr)
	fmt.Printf("  Public key (base64): %s\n", pubKeyStr)
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
