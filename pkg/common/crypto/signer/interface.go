package signer

import ethcommon "github.com/ethereum/go-ethereum/common"

// Signer interface defines methods for signing messages
type Signer interface {
	// Sign signs the message with the signing key
	Sign(message []byte) ([]byte, error)
	// GetSigningAddress returns the address derived from signing key
	GetSigningAddress() ethcommon.Address
	// SignTaskResponse signs a task response
	SignTaskResponse(params TaskSignParams) ([]byte, error)
	// VerifyTaskResponse verifies a task response signature
	VerifyTaskResponse(params TaskSignParams, signature []byte, signerAddr string) error
}
