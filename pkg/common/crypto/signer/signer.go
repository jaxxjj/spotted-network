package signer

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	p2pcrypto "github.com/libp2p/go-libp2p/core/crypto"
)

// TaskSignParams contains all fields needed for signing a task response
type TaskSignParams struct {
	User        ethcommon.Address
	ChainID     uint32
	BlockNumber uint64
	Key         *big.Int
	Value       *big.Int
}
type Config struct {
	SigningKeyPath string
	Password       string
	P2PKeyPath     string
}

// LocalSigner implements Signer interface using a local keystore file
type LocalSigner struct {
	signingKey *ecdsa.PrivateKey // Used for signing task responses
}

// NewLocalSigner creates a new local signer
func NewLocalSigner(cfg *Config) (*LocalSigner, error) {
	// Load signing key
	signingKeyJson, err := os.ReadFile(cfg.SigningKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read signing key file: %w", err)
	}
	signingKey, err := keystore.DecryptKey(signingKeyJson, cfg.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt signing key: %w", err)
	}

	return &LocalSigner{
		signingKey: signingKey.PrivateKey,
	}, nil
}

// GetSigningAddress returns the address derived from signing key
func (s *LocalSigner) GetSigningAddress() ethcommon.Address {
	return ethcrypto.PubkeyToAddress(s.signingKey.PublicKey)
}

// SignTaskResponse signs a task response with all required fields
func (s *LocalSigner) SignTaskResponse(params TaskSignParams) ([]byte, error) {
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
func (s *LocalSigner) VerifyTaskResponse(params TaskSignParams, signature []byte, signerAddr string) error {
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

// LoadPrivateKeyFromFile loads an Ed25519 private key from a file
func LoadPrivateKeyFromFile(keyPath string) (p2pcrypto.PrivKey, error) {
	// Read the key file
	keyBytes, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read key file: %w", err)
	}

	// Unmarshal the private key
	privKey, err := p2pcrypto.UnmarshalPrivateKey(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal private key: %w", err)
	}

	// Verify that it's an Ed25519 key
	if privKey.Type() != p2pcrypto.Ed25519 {
		return nil, fmt.Errorf("invalid key type: expected Ed25519, got %s", privKey.Type())
	}

	return privKey, nil
}
