package signer

import (
	"crypto/ecdsa"
	"fmt"
	"io/ioutil"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// Signer interface defines methods for signing messages
type Signer interface {
	// Sign signs the message with the private key
	Sign(message []byte) ([]byte, error)
	// GetAddress returns the ethereum address associated with the signer
	GetAddress() common.Address
	// GetPublicKey returns the public key associated with the signer
	GetPublicKey() *ecdsa.PublicKey
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
func VerifySignature(message []byte, signature []byte, address common.Address) bool {
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
