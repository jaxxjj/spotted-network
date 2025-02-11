package signer_test

import (
	"encoding/hex"
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	p2pcrypto "github.com/libp2p/go-libp2p/core/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/sha3"

	"github.com/galxe/spotted-network/pkg/common/crypto/signer"
)

func setupSigner(t *testing.T) signer.Signer {
	signingKeyPath := "mock/dummy.key.json"
	password := "testpassword"

	s, err := signer.NewLocalSigner(&signer.Config{
		SigningKeyPath: signingKeyPath,
		Password:       password,
	})
	require.NoError(t, err)
	return s
}

func TestLocalSignerWithKeystore(t *testing.T) {
	s := setupSigner(t)

	// Test signing address matches expected
	expectedAddr := common.HexToAddress("0x9F0D8BAC11C5693a290527f09434b86651c66Bf2")
	assert.Equal(t, expectedAddr, s.GetSigningAddress())

	// Run other signing tests
	t.Run("signing functions", func(t *testing.T) {
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

		// Test signature verification with modified parameters
		modifiedParams := params
		modifiedParams.Value = big.NewInt(200)
		err = s.VerifyTaskResponse(modifiedParams, sig, s.GetSigningAddress().Hex())
		require.Error(t, err)
	})

	t.Run("generic signing", func(t *testing.T) {
		message := []byte("test message")
		sig, err := s.Sign(message)
		require.NoError(t, err)

		// Verify signature
		hash := crypto.Keccak256(message)
		pubKey, err := crypto.SigToPub(hash, sig)
		require.NoError(t, err)
		recoveredAddr := crypto.PubkeyToAddress(*pubKey)
		assert.Equal(t, s.GetSigningAddress(), recoveredAddr)
	})
}

func TestLocalSignerWithPrivateKey(t *testing.T) {
	// Create test key
	signingKey, err := crypto.GenerateKey()
	require.NoError(t, err)

	// Get private key hex
	privateKeyBytes := crypto.FromECDSA(signingKey)
	privateKeyHex := hex.EncodeToString(privateKeyBytes)

	// Create signer with private key
	cfg := &signer.Config{
		SigningKey: privateKeyHex,
	}
	s, err := signer.NewLocalSigner(cfg)
	require.NoError(t, err)

	// Test signing address matches expected
	expectedAddr := crypto.PubkeyToAddress(signingKey.PublicKey)
	assert.Equal(t, expectedAddr, s.GetSigningAddress())

	// Run other signing tests
	t.Run("signing functions", func(t *testing.T) {
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

		// Test signature verification with modified parameters
		modifiedParams := params
		modifiedParams.Value = big.NewInt(200)
		err = s.VerifyTaskResponse(modifiedParams, sig, s.GetSigningAddress().Hex())
		require.Error(t, err)
	})

	t.Run("generic signing", func(t *testing.T) {
		message := []byte("test message")
		sig, err := s.Sign(message)
		require.NoError(t, err)

		// Verify signature
		hash := crypto.Keccak256(message)
		pubKey, err := crypto.SigToPub(hash, sig)
		require.NoError(t, err)
		recoveredAddr := crypto.PubkeyToAddress(*pubKey)
		assert.Equal(t, s.GetSigningAddress(), recoveredAddr)
	})
}

func TestLocalSignerConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      *signer.Config
		expectError bool
	}{
		{
			name:        "nil config",
			config:      nil,
			expectError: true,
		},
		{
			name: "both keystore and private key provided",
			config: &signer.Config{
				SigningKeyPath: "path",
				SigningKey:     "key",
			},
			expectError: true,
		},
		{
			name: "neither keystore nor private key provided",
			config: &signer.Config{
				Password: "password",
			},
			expectError: true,
		},
		{
			name: "invalid private key hex",
			config: &signer.Config{
				SigningKey: "invalid-hex",
			},
			expectError: true,
		},
		{
			name: "non-existent keystore file",
			config: &signer.Config{
				SigningKeyPath: "non-existent-file",
				Password:       "password",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := signer.NewLocalSigner(tt.config)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSignTaskResponse(t *testing.T) {
	s := setupSigner(t)

	tests := []struct {
		name        string
		params      signer.TaskSignParams
		expectError bool
	}{
		{
			name: "valid parameters",
			params: signer.TaskSignParams{
				User:        common.HexToAddress("0x1234567890123456789012345678901234567890"),
				ChainID:     1,
				BlockNumber: 12345,
				Key:         big.NewInt(1),
				Value:       big.NewInt(100),
			},
			expectError: false,
		},
		{
			name: "zero values",
			params: signer.TaskSignParams{
				User:        common.HexToAddress("0x0000000000000000000000000000000000000000"),
				ChainID:     0,
				BlockNumber: 0,
				Key:         big.NewInt(0),
				Value:       big.NewInt(0),
			},
			expectError: false,
		},
		{
			name: "large numbers",
			params: signer.TaskSignParams{
				User:        common.HexToAddress("0xffffffffffffffffffffffffffffffffffffffff"),
				ChainID:     4294967295,
				BlockNumber: 18446744073709551615,
				Key: func() *big.Int {
					n, _ := new(big.Int).SetString("115792089237316195423570985008687907853269984665640564039457584007913129639935", 10)
					return n
				}(),
				Value: func() *big.Int {
					n, _ := new(big.Int).SetString("115792089237316195423570985008687907853269984665640564039457584007913129639935", 10)
					return n
				}(),
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 签名
			sig, err := s.SignTaskResponse(tt.params)
			if tt.expectError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, sig)

			// 验证签名
			err = s.VerifyTaskResponse(tt.params, sig, s.GetSigningAddress().Hex())
			require.NoError(t, err)

			// 验证参数修改后签名失效
			modifiedParams := tt.params
			modifiedParams.Value = new(big.Int).Add(tt.params.Value, big.NewInt(1))
			err = s.VerifyTaskResponse(modifiedParams, sig, s.GetSigningAddress().Hex())
			require.Error(t, err)
		})
	}
}

func TestSignAndVerify(t *testing.T) {
	s := setupSigner(t)

	tests := []struct {
		name        string
		message     []byte
		expectError bool
	}{
		{
			name:        "empty message",
			message:     []byte{},
			expectError: false,
		},
		{
			name:        "normal message",
			message:     []byte("test message"),
			expectError: false,
		},
		{
			name:        "long message",
			message:     []byte(strings.Repeat("long message ", 100)),
			expectError: false,
		},
		{
			name:        "binary data",
			message:     []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 签名
			sig, err := s.Sign(tt.message)
			if tt.expectError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, sig)

			// 验证签名
			hash := crypto.Keccak256(tt.message)
			pubKey, err := crypto.SigToPub(hash, sig)
			require.NoError(t, err)
			recoveredAddr := crypto.PubkeyToAddress(*pubKey)
			assert.Equal(t, s.GetSigningAddress(), recoveredAddr)

			// 验证消息修改后签名失效
			modifiedMessage := append(tt.message, byte(0))
			modifiedHash := crypto.Keccak256(modifiedMessage)
			pubKey, err = crypto.SigToPub(modifiedHash, sig)
			require.NoError(t, err)
			recoveredAddr = crypto.PubkeyToAddress(*pubKey)
			assert.NotEqual(t, s.GetSigningAddress(), recoveredAddr)
		})
	}
}

func TestBase64ToPrivKey(t *testing.T) {
	tests := []struct {
		name        string
		base64Key   string
		expectedP2P string
		expectError bool
	}{
		{
			name:        "operator1 key",
			base64Key:   "CAESQKW/y8x4MBT09AySrCDS1HXvsFEGoXLwqvWOQUifZ90TvdsBG0rSgcjJTH8qWwRYRysJaZ+7Z4egLxvShvBnQys=",
			expectedP2P: "0x310c8425b620980dcfcf756e46572bb6ac80eb07",
			expectError: false,
		},
		{
			name:        "operator2 key",
			base64Key:   "CAESQHGMebvS8Wf6IZZh40yacCPzXhRlKqJCGfPySZyCFid6EdbnbwgelZkcZbllzWAZFfrdV/dcf2poB1OySA2mV0I=",
			expectedP2P: "0x01078ffbf1de436d6f429f5ce6be8fd9d6e16165",
			expectError: false,
		},
		{
			name:        "operator3 key",
			base64Key:   "CAESQM5ltPHuttHq7/HHHHymN5A/XSDKt5EPOwGWor2H3k0PXckF23DDwxzmdOhEtOy5f8szIAYWqSFH8cIlICumemo=",
			expectedP2P: "0x67aa23adde2459a1620be2ea28982310597521b0",
			expectError: false,
		},
		{
			name:        "invalid base64 string",
			base64Key:   "invalid-base64",
			expectError: true,
		},
		{
			name:        "invalid key bytes",
			base64Key:   "aGVsbG8=", // "hello" in base64
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			privKey, err := signer.Base64ToPrivKey(tt.base64Key)
			if tt.expectError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, privKey)

			if tt.expectedP2P != "" {
				// 获取公钥并序列化
				pubKey := privKey.GetPublic()
				pubKeyBytes, err := p2pcrypto.MarshalPublicKey(pubKey)
				require.NoError(t, err)

				// 计算Keccak256哈希
				hasher := sha3.NewLegacyKeccak256()
				hasher.Write(pubKeyBytes)
				hash := hasher.Sum(nil)

				// 取最后20字节作为地址
				address := hash[len(hash)-20:]
				p2pAddr := "0x" + hex.EncodeToString(address)

				assert.Equal(t, tt.expectedP2P, strings.ToLower(p2pAddr))
			}
		})
	}
}
