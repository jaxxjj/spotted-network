package signer

import (
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"math/big"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	p2pcrypto "github.com/libp2p/go-libp2p/core/crypto"
)

type Config struct {
	// SigningKeyPath and SigningKey are mutually exclusive
	SigningKeyPath string
	SigningKey     string // hex string of private key without 0x prefix
	Password       string
}

// LocalSigner implements Signer interface using a local keystore file
type localSigner struct {
	signingKey *ecdsa.PrivateKey // Used for signing task responses
}

// NewLocalSigner creates a new local signer
func NewLocalSigner(cfg *Config) (Signer, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is nil")
	}

	// Validate config - exactly one signing key option must be provided XOR
	if (cfg.SigningKeyPath == "") == (cfg.SigningKey == "") || (cfg.SigningKeyPath != "" && cfg.SigningKey != "") {
		return nil, fmt.Errorf("exactly one of SigningKeyPath or SigningKey must be provided")
	}

	var privateKey *ecdsa.PrivateKey

	if cfg.SigningKeyPath != "" {
		// Load from keystore file
		signingKeyJson, err := os.ReadFile(cfg.SigningKeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read signing key file: %w", err)
		}
		key, err := keystore.DecryptKey(signingKeyJson, cfg.Password)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt signing key: %w", err)
		}
		privateKey = key.PrivateKey
	} else {
		// Load from hex string
		privateKeyStr := strings.TrimPrefix(cfg.SigningKey, "0x")
		privateKeyBytes, err := hex.DecodeString(privateKeyStr)
		if err != nil {
			return nil, fmt.Errorf("failed to decode private key hex string: %w", err)
		}
		privateKey, err = ethcrypto.ToECDSA(privateKeyBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to convert private key bytes to ECDSA: %w", err)
		}
	}

	return &localSigner{
		signingKey: privateKey,
	}, nil
}

// GetSigningAddress returns the address derived from signing key
func (s *localSigner) GetSigningAddress() ethcommon.Address {
	return ethcrypto.PubkeyToAddress(s.signingKey.PublicKey)
}

// SignTaskResponse signs a task response with all required fields
func (s *localSigner) SignTaskResponse(params TaskSignParams) ([]byte, error) {
	// Pack parameters into bytes
	msg := []byte{}
	msg = append(msg, params.User.Bytes()...)
	msg = append(msg, ethcommon.LeftPadBytes(big.NewInt(int64(params.ChainID)).Bytes(), 32)...)
	msg = append(msg, ethcommon.LeftPadBytes(big.NewInt(int64(params.BlockNumber)).Bytes(), 32)...)
	msg = append(msg, ethcommon.LeftPadBytes(params.Key.Bytes(), 32)...)
	msg = append(msg, ethcommon.LeftPadBytes(params.Value.Bytes(), 32)...)

	// Hash the message
	hash := ethcrypto.Keccak256(msg)

	// Sign the hash
	return ethcrypto.Sign(hash, s.signingKey)
}

// VerifyTaskResponse verifies a task response signature
func (s *localSigner) VerifyTaskResponse(params TaskSignParams, signature []byte, signerAddr string) error {
	// Pack parameters into bytes in the same order as SignTaskResponse
	msg := []byte{}
	msg = append(msg, params.User.Bytes()...)
	msg = append(msg, ethcommon.LeftPadBytes(big.NewInt(int64(params.ChainID)).Bytes(), 32)...)
	msg = append(msg, ethcommon.LeftPadBytes(big.NewInt(int64(params.BlockNumber)).Bytes(), 32)...)
	msg = append(msg, ethcommon.LeftPadBytes(params.Key.Bytes(), 32)...)
	msg = append(msg, ethcommon.LeftPadBytes(params.Value.Bytes(), 32)...)

	// Hash the message
	hash := ethcrypto.Keccak256(msg)

	// Recover public key
	pubKey, err := ethcrypto.SigToPub(hash, signature)
	if err != nil {
		return fmt.Errorf("failed to recover public key: %w", err)
	}

	// Convert public key to address
	recoveredAddr := ethcrypto.PubkeyToAddress(*pubKey)
	if recoveredAddr.Hex() != signerAddr {
		return fmt.Errorf("invalid signature: recovered address %s does not match signer %s", recoveredAddr.Hex(), signerAddr)
	}

	return nil
}

// Sign implements the Signer interface
func (s *localSigner) Sign(message []byte) ([]byte, error) {
	// Hash the message
	hash := ethcrypto.Keccak256(message)
	// Sign the hash
	return ethcrypto.Sign(hash, s.signingKey)
}

// Base64ToPrivKey converts a base64 encoded private key to libp2p private key
func Base64ToPrivKey(encodedKey string) (p2pcrypto.PrivKey, error) {
	privKeyBytes, err := p2pcrypto.ConfigDecodeKey(encodedKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 key: %v", err)
	}

	privKey, err := p2pcrypto.UnmarshalPrivateKey(privKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal private key: %v", err)
	}

	return privKey, nil
}
