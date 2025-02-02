package signer_test

import (
	"bytes"
	"math/big"
	"os"
	"path/filepath"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/galxe/spotted-network/pkg/common/crypto/signer"
)

func TestLocalSigner(t *testing.T) {
	// Create temporary directory for keystore files
	tmpDir, err := os.MkdirTemp("", "signer-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create test keys
	operatorKey, err := crypto.GenerateKey()
	require.NoError(t, err)
	signingKey, err := crypto.GenerateKey()
	require.NoError(t, err)

	// Create keystore and import keys
	ks := keystore.NewKeyStore(tmpDir, keystore.StandardScryptN, keystore.StandardScryptP)
	password := "testpassword"

	// Import keys
	operatorAccount, err := ks.ImportECDSA(operatorKey, password)
	require.NoError(t, err)
	signingAccount, err := ks.ImportECDSA(signingKey, password)
	require.NoError(t, err)

	// Get keystore paths
	operatorKeyPath := filepath.Join(tmpDir, operatorAccount.URL.Path)
	signingKeyPath := filepath.Join(tmpDir, signingAccount.URL.Path)

	// Create signer
	s, err := signer.NewLocalSigner(operatorKeyPath, signingKeyPath, password)
	require.NoError(t, err)

	t.Run("operator key functions", func(t *testing.T) {
		// Test operator address
		expectedOperatorAddr := crypto.PubkeyToAddress(operatorKey.PublicKey)
		assert.Equal(t, expectedOperatorAddr, s.GetOperatorAddress())

		// Test join request signing
		joinMsg := []byte("join request")
		sig, err := s.Sign(joinMsg)
		require.NoError(t, err)

		// Verify join signature
		hash := crypto.Keccak256(joinMsg)
		recoveredPub, err := crypto.SigToPub(hash, sig)
		require.NoError(t, err)
		recoveredAddr := crypto.PubkeyToAddress(*recoveredPub)
		assert.Equal(t, expectedOperatorAddr, recoveredAddr)
	})

	t.Run("signing key functions", func(t *testing.T) {
		// Test signing address
		expectedSigningAddr := crypto.PubkeyToAddress(signingKey.PublicKey)
		assert.Equal(t, expectedSigningAddr, s.GetSigningAddress())

		// Test message signing
		message := []byte("test message")
		sig, err := s.Sign(message)
		require.NoError(t, err)

		// Verify signature using package function
		assert.True(t, signer.VerifySignature(expectedSigningAddr, message, sig))
	})

	t.Run("task response signing", func(t *testing.T) {
		params := signer.TaskSignParams{
			User:        common.HexToAddress("0x1234567890123456789012345678901234567890"),
			ChainID:     1,
			BlockNumber: 12345,
			Key:         big.NewInt(1),
			Value:       big.NewInt(100),
		}

		// Sign task response
		sig, err := s.SignTaskResponse(params)
		require.NoError(t, err)

		// Verify task response
		err = s.VerifyTaskResponse(params, sig, s.GetSigningAddress().Hex())
		require.NoError(t, err)

		// Test invalid signer address
		wrongAddr := common.HexToAddress("0x0000000000000000000000000000000000000000")
		err = s.VerifyTaskResponse(params, sig, wrongAddr.Hex())
		require.Error(t, err)
	})
}

func TestAggregateSignatures(t *testing.T) {
	// Create temporary directory for keystore files
	tmpDir, err := os.MkdirTemp("", "signer-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create test keys
	operatorKey1, err := crypto.GenerateKey()
	require.NoError(t, err)
	operatorKey2, err := crypto.GenerateKey()
	require.NoError(t, err)
	operatorKey3, err := crypto.GenerateKey()
	require.NoError(t, err)
	
	signingKey1, err := crypto.GenerateKey()
	require.NoError(t, err)
	signingKey2, err := crypto.GenerateKey()
	require.NoError(t, err)
	signingKey3, err := crypto.GenerateKey()
	require.NoError(t, err)

	// Create keystore files
	password := "testpassword"
	
	// Create keystore and import keys
	ks := keystore.NewKeyStore(tmpDir, keystore.StandardScryptN, keystore.StandardScryptP)
	
	// Import operator keys
	acc1, err := ks.ImportECDSA(operatorKey1, password)
	require.NoError(t, err)
	acc2, err := ks.ImportECDSA(operatorKey2, password)
	require.NoError(t, err)
	acc3, err := ks.ImportECDSA(operatorKey3, password)
	require.NoError(t, err)
	
	// Import signing keys
	signAcc1, err := ks.ImportECDSA(signingKey1, password)
	require.NoError(t, err)
	signAcc2, err := ks.ImportECDSA(signingKey2, password)
	require.NoError(t, err)
	signAcc3, err := ks.ImportECDSA(signingKey3, password)
	require.NoError(t, err)

	// Create signers
	signer1, err := signer.NewLocalSigner(
		filepath.Join(tmpDir, acc1.URL.Path),
		filepath.Join(tmpDir, signAcc1.URL.Path),
		password,
	)
	require.NoError(t, err)
	signer2, err := signer.NewLocalSigner(
		filepath.Join(tmpDir, acc2.URL.Path),
		filepath.Join(tmpDir, signAcc2.URL.Path),
		password,
	)
	require.NoError(t, err)
	signer3, err := signer.NewLocalSigner(
		filepath.Join(tmpDir, acc3.URL.Path),
		filepath.Join(tmpDir, signAcc3.URL.Path),
		password,
	)
	require.NoError(t, err)

	message := []byte("test message")
	
	sig1, err := signer1.Sign(message)
	require.NoError(t, err)
	sig2, err := signer2.Sign(message)
	require.NoError(t, err)
	sig3, err := signer3.Sign(message)
	require.NoError(t, err)

	t.Run("normal aggregation", func(t *testing.T) {
		sigs := map[string][]byte{
			signer1.GetSigningAddress().Hex(): sig1,
			signer2.GetSigningAddress().Hex(): sig2,
			signer3.GetSigningAddress().Hex(): sig3,
		}

		// Aggregate signatures
		aggregated := signer1.AggregateSignatures(sigs)

		// Verify length
		expectedLen := len(sig1) + len(sig2) + len(sig3)
		assert.Equal(t, expectedLen, len(aggregated))

		// Verify deterministic ordering
		aggregated2 := signer2.AggregateSignatures(sigs)
		assert.True(t, bytes.Equal(aggregated, aggregated2))
	})

	t.Run("empty input", func(t *testing.T) {
		sigs := map[string][]byte{}
		aggregated := signer1.AggregateSignatures(sigs)
		assert.Empty(t, aggregated)
	})

	t.Run("single signature", func(t *testing.T) {
		sigs := map[string][]byte{
			signer1.GetSigningAddress().Hex(): sig1,
		}
		aggregated := signer1.AggregateSignatures(sigs)
		assert.True(t, bytes.Equal(sig1, aggregated))
	})

	t.Run("verify individual signatures in aggregated", func(t *testing.T) {
		sigs := map[string][]byte{
			signer1.GetSigningAddress().Hex(): sig1,
			signer2.GetSigningAddress().Hex(): sig2,
		}
		
		aggregated := signer1.AggregateSignatures(sigs)
		
		// Extract and verify individual signatures
		sig1Len := len(sig1)
		sig2Len := len(sig2)
		
		// Get individual signatures from aggregated
		extractedSig1 := aggregated[:sig1Len]
		extractedSig2 := aggregated[sig1Len:sig1Len+sig2Len]
		
		// Verify each signature
		assert.True(t, signer.VerifySignature(signer1.GetSigningAddress(), message, extractedSig1))
		assert.True(t, signer.VerifySignature(signer2.GetSigningAddress(), message, extractedSig2))
	})
}
