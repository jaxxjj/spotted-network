package signer_test

import (
	"bytes"
	"math/big"
	"os"
	"path/filepath"
	"testing"

	"github.com/Layr-Labs/eigensdk-go/signerv2"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/galxe/spotted-network/pkg/common/crypto/signer"
)

func TestPrivateKeySigner(t *testing.T) {
	privateKeyHex := "ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	require.NoError(t, err)
	chainID := big.NewInt(1)

	signer, err := signerv2.PrivateKeySignerFn(privateKey, chainID)
	require.NoError(t, err)

	address := crypto.PubkeyToAddress(privateKey.PublicKey)
	tx := types.NewTx(&types.DynamicFeeTx{
		Nonce:   0,
		Value:   big.NewInt(0),
		ChainID: chainID,
		Data:    common.Hex2Bytes("6057361d00000000000000000000000000000000000000000000000000000000000f4240"),
	})
	signedTx, err := signer(address, tx)
	require.NoError(t, err)

	// Verify the sender address of the signed transaction
	from, err := types.Sender(types.LatestSignerForChainID(chainID), signedTx)
	require.NoError(t, err)
	require.Equal(t, address, from)
}

func TestKeystoreSigner(t *testing.T) {
	keystorePath := "mockdata/dummy.key.json"
	keystorePassword := "testpassword"
	chainID := big.NewInt(1)
	signer, err := signerv2.KeyStoreSignerFn(keystorePath, keystorePassword, chainID)
	require.NoError(t, err)

	privateKey, err := crypto.HexToECDSA("7a28b5ba57c53603b0b07b56bba752f7784bf506fa95edc395f5cf6c7514fe9d")
	require.NoError(t, err)

	address := crypto.PubkeyToAddress(privateKey.PublicKey)
	tx := types.NewTx(&types.DynamicFeeTx{
		Nonce:   0,
		Value:   big.NewInt(0),
		ChainID: chainID,
		Data:    common.Hex2Bytes("6057361d00000000000000000000000000000000000000000000000000000000000f4240"),
	})
	signedTx, err := signer(address, tx)
	require.NoError(t, err)

	// Verify the sender address of the signed transaction
	from, err := types.Sender(types.LatestSignerForChainID(chainID), signedTx)
	require.NoError(t, err)
	require.Equal(t, address, from)
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
	password := "testpass"
	
	// Create keystore and import keys
	ks := keystore.NewKeyStore(tmpDir, keystore.StandardScryptN, keystore.StandardScryptP)
	
	// Import operator keys
	_, err = ks.ImportECDSA(operatorKey1, password)
	require.NoError(t, err)
	_, err = ks.ImportECDSA(operatorKey2, password)
	require.NoError(t, err)
	_, err = ks.ImportECDSA(operatorKey3, password)
	require.NoError(t, err)
	
	// Import signing keys
	_, err = ks.ImportECDSA(signingKey1, password)
	require.NoError(t, err)
	_, err = ks.ImportECDSA(signingKey2, password)
	require.NoError(t, err)
	_, err = ks.ImportECDSA(signingKey3, password)
	require.NoError(t, err)

	// Get keystore file paths
	files, err := os.ReadDir(tmpDir)
	require.NoError(t, err)
	require.Len(t, files, 6) // 3 operator keys + 3 signing keys

	// Get operator key paths
	operatorKs1 := filepath.Join(tmpDir, files[0].Name())
	operatorKs2 := filepath.Join(tmpDir, files[1].Name())
	operatorKs3 := filepath.Join(tmpDir, files[2].Name())

	// Get signing key paths
	signingKs1 := filepath.Join(tmpDir, files[3].Name())
	signingKs2 := filepath.Join(tmpDir, files[4].Name())
	signingKs3 := filepath.Join(tmpDir, files[5].Name())

	// Create signers
	signer1, err := signer.NewLocalSigner(operatorKs1, signingKs1, password)
	require.NoError(t, err)
	signer2, err := signer.NewLocalSigner(operatorKs2, signingKs2, password)
	require.NoError(t, err)
	signer3, err := signer.NewLocalSigner(operatorKs3, signingKs3, password)
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
			signer1.Address(): sig1,
			signer2.Address(): sig2,
			signer3.Address(): sig3,
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
			signer1.Address(): sig1,
		}
		aggregated := signer1.AggregateSignatures(sigs)
		assert.True(t, bytes.Equal(sig1, aggregated))
	})

	t.Run("deterministic ordering", func(t *testing.T) {
		// Create signatures in different orders
		sigs1 := map[string][]byte{
			signer1.Address(): sig1,
			signer2.Address(): sig2,
			signer3.Address(): sig3,
		}
		sigs2 := map[string][]byte{
			signer3.Address(): sig3,
			signer1.Address(): sig1,
			signer2.Address(): sig2,
		}

		// Aggregate in different orders
		aggregated1 := signer1.AggregateSignatures(sigs1)
		aggregated2 := signer1.AggregateSignatures(sigs2)

		// Should be identical due to address sorting
		assert.True(t, bytes.Equal(aggregated1, aggregated2))
	})

	t.Run("verify individual signatures in aggregated", func(t *testing.T) {
		sigs := map[string][]byte{
			signer1.Address(): sig1,
			signer2.Address(): sig2,
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
