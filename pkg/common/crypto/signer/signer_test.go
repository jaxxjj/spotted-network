package signer_test

import (
	"math/big"
	"os"
	"path/filepath"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	p2pcrypto "github.com/libp2p/go-libp2p/core/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/galxe/spotted-network/pkg/common/crypto/signer"
)

func TestLocalSigner(t *testing.T) {
	// Create temporary directory for keystore files
	tmpDir, err := os.MkdirTemp("", "signer-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create test key
	signingKey, err := crypto.GenerateKey()
	require.NoError(t, err)

	// Create keystore and import key
	ks := keystore.NewKeyStore(tmpDir, keystore.StandardScryptN, keystore.StandardScryptP)
	password := "testpassword"

	// Import signing key
	signingAccount, err := ks.ImportECDSA(signingKey, password)
	require.NoError(t, err)

	// Get keystore path
	signingKeyPath := filepath.Join(tmpDir, signingAccount.URL.Path)

	// Create signer
	cfg := &signer.Config{
		SigningKeyPath: signingKeyPath,
		Password:       password,
	}
	s, err := signer.NewLocalSigner(cfg)
	require.NoError(t, err)

	t.Run("signing key functions", func(t *testing.T) {
		// Test signing address
		expectedSigningAddr := crypto.PubkeyToAddress(signingKey.PublicKey)
		assert.Equal(t, expectedSigningAddr, s.GetSigningAddress())

		// Test task response signing
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

	t.Run("error cases", func(t *testing.T) {
		// Test with non-existent key file
		_, err := signer.NewLocalSigner(&signer.Config{
			SigningKeyPath: "non-existent-file",
			Password:       password,
		})
		require.Error(t, err)

		// Test with wrong password
		_, err = signer.NewLocalSigner(&signer.Config{
			SigningKeyPath: signingKeyPath,
			Password:       "wrongpassword",
		})
		require.Error(t, err)

		// Test signature verification with modified parameters
		params := signer.TaskSignParams{
			User:        common.HexToAddress("0x1234567890123456789012345678901234567890"),
			ChainID:     1,
			BlockNumber: 12345,
			Key:         big.NewInt(1),
			Value:       big.NewInt(100),
		}

		sig, err := s.SignTaskResponse(params)
		require.NoError(t, err)

		// Modify params and verify should fail
		params.Value = big.NewInt(200)
		err = s.VerifyTaskResponse(params, sig, s.GetSigningAddress().Hex())
		require.Error(t, err)
	})
}

func TestLoadPrivateKeyFromFile(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "p2p-key-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	t.Run("successful key loading", func(t *testing.T) {
		// Generate Ed25519 key
		privKey, _, err := p2pcrypto.GenerateEd25519Key(nil)
		require.NoError(t, err)

		// Marshal the key
		keyBytes, err := p2pcrypto.MarshalPrivateKey(privKey)
		require.NoError(t, err)

		// Write to temporary file
		keyPath := filepath.Join(tmpDir, "valid.key")
		err = os.WriteFile(keyPath, keyBytes, 0600)
		require.NoError(t, err)

		// Test loading
		loadedKey, err := signer.LoadPrivateKeyFromFile(keyPath)
		require.NoError(t, err)
		assert.Equal(t, p2pcrypto.Ed25519, loadedKey.Type())
	})

	t.Run("error cases", func(t *testing.T) {
		// Test non-existent file
		_, err := signer.LoadPrivateKeyFromFile("non-existent-file")
		require.Error(t, err)

		// Test invalid key data
		invalidKeyPath := filepath.Join(tmpDir, "invalid.key")
		err = os.WriteFile(invalidKeyPath, []byte("invalid key data"), 0600)
		require.NoError(t, err)

		_, err = signer.LoadPrivateKeyFromFile(invalidKeyPath)
		require.Error(t, err)
	})
}
