package signer

import (
	"crypto/ecdsa"
	"fmt"
	"io/ioutil"
	"math/big"
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
	// GetAddress returns the ethereum address associated with the signer
	GetAddress() ethcommon.Address
	// GetPublicKey returns the public key associated with the signer
	GetPublicKey() *ecdsa.PublicKey
	// GetSigningKey returns the signing key (public key) as hex string
	GetSigningKey() string
	// SignTaskResponse signs a task response
	SignTaskResponse(params TaskSignParams) ([]byte, error)
	// VerifyTaskResponse verifies a task response signature
	VerifyTaskResponse(params TaskSignParams, signature []byte, signerAddr string) error
	// Address returns the signer's address as string
	Address() string
	// AggregateSignatures combines multiple signatures into one
	AggregateSignatures(sigs map[string][]byte) []byte
}

// LocalSigner implements Signer interface using a local keystore file
type LocalSigner struct {
	privateKey *ecdsa.PrivateKey
	address    ethcommon.Address
}

// NewLocalSigner creates a new LocalSigner from a keystore file
func NewLocalSigner(keystorePath, password string) (*LocalSigner, error) {
	// Read keystore file
	keyjson, err := ioutil.ReadFile(keystorePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read keystore file: %v", err)
	}

	// Decrypt key with password
	key, err := keystore.DecryptKey(keyjson, password)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt key: %v", err)
	}

	return &LocalSigner{
		privateKey: key.PrivateKey,
		address:    crypto.PubkeyToAddress(key.PrivateKey.PublicKey),
	}, nil
}

// Sign implements Signer interface
func (s *LocalSigner) Sign(message []byte) ([]byte, error) {
	// Hash the message
	hash := crypto.Keccak256Hash(message)
	
	// Sign the hash
	signature, err := crypto.Sign(hash.Bytes(), s.privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign message: %v", err)
	}

	return signature, nil
}

// GetAddress implements Signer interface
func (s *LocalSigner) GetAddress() ethcommon.Address {
	return s.address
}

// GetPublicKey implements Signer interface
func (s *LocalSigner) GetPublicKey() *ecdsa.PublicKey {
	return &s.privateKey.PublicKey
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

// GetSigningKey returns the signing key (public key) as hex string
func (s *LocalSigner) GetSigningKey() string {
	return crypto.PubkeyToAddress(s.privateKey.PublicKey).Hex()
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
	return crypto.Sign(hash, s.privateKey)
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

// Address returns the signer's address as string
func (s *LocalSigner) Address() string {
	return s.address.Hex()
}

// VerifySignature verifies if the signature matches the task parameters
func (s *LocalSigner) VerifySignature(signature []byte, params TaskSignParams) error {
	// Pack parameters into bytes
	msg := []byte{}
	msg = append(msg, params.User.Bytes()...)
	msg = append(msg, ethcommon.LeftPadBytes(big.NewInt(int64(params.ChainID)).Bytes(), 32)...)
	msg = append(msg, ethcommon.LeftPadBytes(big.NewInt(int64(params.BlockNumber)).Bytes(), 32)...)
	msg = append(msg, ethcommon.LeftPadBytes(params.Key.Bytes(), 32)...)
	msg = append(msg, ethcommon.LeftPadBytes(params.Value.Bytes(), 32)...)

	// Hash the message
	hash := crypto.Keccak256(msg)

	// Recover public key
	pubkey, err := crypto.SigToPub(hash, signature)
	if err != nil {
		return fmt.Errorf("failed to recover public key: %w", err)
	}

	// Convert public key to address
	recoveredAddr := crypto.PubkeyToAddress(*pubkey)
	if recoveredAddr != params.User {
		return fmt.Errorf("invalid signature: recovered address %s does not match expected %s", recoveredAddr.Hex(), params.User.Hex())
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
