package signer

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"os"
	"sort"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// TaskSignParams contains all fields needed for signing a task response
type TaskSignParams struct {
	User        ethcommon.Address
	ChainID     uint32
	BlockNumber uint64
	Key         *big.Int
	Value       *big.Int
}

// Signer interface defines methods for signing messages
type Signer interface {
	// Sign signs the message with the private key
	Sign(message []byte) ([]byte, error)
	// GetOperatorAddress returns the operator's address (from operator key)
	GetOperatorAddress() ethcommon.Address
	// GetSigningAddress returns the address derived from signing key
	GetSigningAddress() ethcommon.Address
	// SignTaskResponse signs a task response
	SignTaskResponse(params TaskSignParams) ([]byte, error)
	// VerifyTaskResponse verifies a task response signature
	VerifyTaskResponse(params TaskSignParams, signature []byte, signerAddr string) error
	// AggregateSignatures combines multiple signatures into one
	AggregateSignatures(sigs map[string][]byte) []byte
}

// LocalSigner implements Signer interface using a local keystore file
type LocalSigner struct {
	operatorKey *ecdsa.PrivateKey  // Used for operator registration and join
	signingKey  *ecdsa.PrivateKey  // Used for signing task responses
	address     ethcommon.Address  // Operator's address (from operatorKey)
}

// NewLocalSigner creates a new local signer
func NewLocalSigner(operatorKeyPath string, signingKeyPath string, password string) (*LocalSigner, error) {
	// Load operator key
	operatorKeyJson, err := os.ReadFile(operatorKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read operator key file: %w", err)
	}
	operatorKey, err := keystore.DecryptKey(operatorKeyJson, password)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt operator key: %w", err)
	}

	// Load signing key
	signingKeyJson, err := os.ReadFile(signingKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read signing key file: %w", err)
	}
	signingKey, err := keystore.DecryptKey(signingKeyJson, password)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt signing key: %w", err)
	}

	// Create signer with both keys
	return &LocalSigner{
		operatorKey: operatorKey.PrivateKey,
		signingKey:  signingKey.PrivateKey,
		address:     crypto.PubkeyToAddress(operatorKey.PrivateKey.PublicKey),
	}, nil
}

// GetOperatorAddress returns the operator's address (from operator key)
func (s *LocalSigner) GetOperatorAddress() ethcommon.Address {
	return s.address
}

// GetSigningAddress returns the address derived from signing key
func (s *LocalSigner) GetSigningAddress() ethcommon.Address {
	return crypto.PubkeyToAddress(s.signingKey.PublicKey)
}

// GetSigningPublicKey implements Signer interface
func (s *LocalSigner) GetSigningPublicKey() *ecdsa.PublicKey {
	return &s.signingKey.PublicKey
}

// Sign implements Signer interface
func (s *LocalSigner) Sign(message []byte) ([]byte, error) {
	// Hash the message
	hash := crypto.Keccak256Hash(message)
	
	// Sign the hash
	signature, err := crypto.Sign(hash.Bytes(), s.operatorKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign message: %v", err)
	}

	return signature, nil
}

// VerifySignature verifies if the signature was signed by the given address
func VerifySignature(address ethcommon.Address, message []byte, signature []byte) bool {
	// Hash the message
	hash := crypto.Keccak256Hash(message)

	// Get public key from signature
	pubkey, err := crypto.SigToPub(hash.Bytes(), signature)
	if err != nil {
		return false
	}

	// Verify address matches
	recoveredAddr := crypto.PubkeyToAddress(*pubkey)
	return address == recoveredAddr
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
	hash := crypto.Keccak256(msg)

	// Sign the hash
	return crypto.Sign(hash, s.signingKey)
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
	hash := crypto.Keccak256(msg)

	// Recover public key
	pubKey, err := crypto.SigToPub(hash, signature)
	if err != nil {
		return fmt.Errorf("failed to recover public key: %w", err)
	}

	// Convert public key to address
	recoveredAddr := crypto.PubkeyToAddress(*pubKey)
	if recoveredAddr.Hex() != signerAddr {
		return fmt.Errorf("invalid signature: recovered address %s does not match signer %s", recoveredAddr.Hex(), signerAddr)
	}

	return nil
}

// AggregateSignatures combines multiple ECDSA signatures by concatenation
// The signatures are sorted by signer address to ensure deterministic ordering
func (s *LocalSigner) AggregateSignatures(sigs map[string][]byte) []byte {
	// Convert map to sorted slice to ensure deterministic ordering
	type sigPair struct {
		address string
		sig     []byte
	}
	pairs := make([]sigPair, 0, len(sigs))
	for addr, sig := range sigs {
		pairs = append(pairs, sigPair{addr, sig})
	}
	
	// Sort by address
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].address < pairs[j].address
	})
	
	// Concatenate all signatures
	var aggregated []byte
	for _, pair := range pairs {
		aggregated = append(aggregated, pair.sig...)
	}
	return aggregated
}
