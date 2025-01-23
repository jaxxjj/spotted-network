package signer

import (
	"crypto/ecdsa"
	"fmt"
	"io/ioutil"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// TaskSignParams contains all fields needed for signing a task response
type TaskSignParams struct {
	User        common.Address
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
	GetAddress() common.Address
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
}

// LocalSigner implements Signer interface using a local keystore file
type LocalSigner struct {
	privateKey *ecdsa.PrivateKey
	address    common.Address
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
func (s *LocalSigner) GetAddress() common.Address {
	return s.address
}

// GetPublicKey implements Signer interface
func (s *LocalSigner) GetPublicKey() *ecdsa.PublicKey {
	return &s.privateKey.PublicKey
}

// VerifySignature verifies if the signature was signed by the given address
func VerifySignature(address common.Address, message []byte, signature []byte) bool {
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
	msg = append(msg, common.LeftPadBytes(big.NewInt(int64(params.ChainID)).Bytes(), 32)...)
	msg = append(msg, common.LeftPadBytes(big.NewInt(int64(params.BlockNumber)).Bytes(), 32)...)
	msg = append(msg, common.LeftPadBytes(params.Key.Bytes(), 32)...)
	msg = append(msg, common.LeftPadBytes(params.Value.Bytes(), 32)...)

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
	msg = append(msg, common.LeftPadBytes(big.NewInt(int64(params.ChainID)).Bytes(), 32)...)
	msg = append(msg, common.LeftPadBytes(big.NewInt(int64(params.BlockNumber)).Bytes(), 32)...)
	msg = append(msg, common.LeftPadBytes(params.Key.Bytes(), 32)...)
	msg = append(msg, common.LeftPadBytes(params.Value.Bytes(), 32)...)

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
	msg = append(msg, common.LeftPadBytes(big.NewInt(int64(params.ChainID)).Bytes(), 32)...)
	msg = append(msg, common.LeftPadBytes(big.NewInt(int64(params.BlockNumber)).Bytes(), 32)...)
	msg = append(msg, common.LeftPadBytes(params.Key.Bytes(), 32)...)
	msg = append(msg, common.LeftPadBytes(params.Value.Bytes(), 32)...)

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
