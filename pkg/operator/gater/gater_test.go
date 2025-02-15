package gater

import (
	"context"
	"fmt"
	"testing"
	"time"

	"sync"

	"github.com/coocood/freecache"
	"github.com/galxe/spotted-network/pkg/repos/blacklist"
	"github.com/galxe/spotted-network/pkg/repos/consensus_responses"
	"github.com/galxe/spotted-network/pkg/repos/operators"
	"github.com/galxe/spotted-network/pkg/repos/tasks"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
	"github.com/stumble/dcache"
	"github.com/stumble/wpgx/testsuite"
)

type operatorStatesTableSerde struct {
	operatorStatesQuerier *operators.Queries
}

func (o operatorStatesTableSerde) Load(data []byte) error {
	if err := o.operatorStatesQuerier.Load(context.Background(), data); err != nil {
		return fmt.Errorf("failed to load operator states data: %w", err)
	}
	return nil
}

func (o operatorStatesTableSerde) Dump() ([]byte, error) {
	return o.operatorStatesQuerier.Dump(context.Background(), func(m *operators.Operators) {
		m.CreatedAt = time.Unix(0, 0).UTC()
		m.UpdatedAt = time.Unix(0, 0).UTC()
	})
}

type blacklistStatesTableSerde struct {
	blacklistQuerier *blacklist.Queries
}

func (o blacklistStatesTableSerde) Load(data []byte) error {
	if err := o.blacklistQuerier.Load(context.Background(), data); err != nil {
		return fmt.Errorf("failed to load blacklist states data: %w", err)
	}
	return nil
}

func (o blacklistStatesTableSerde) Dump() ([]byte, error) {
	return o.blacklistQuerier.Dump(context.Background(), func(m *blacklist.Blacklist) {
		m.CreatedAt = time.Unix(0, 0).UTC()
		m.UpdatedAt = time.Unix(0, 0).UTC()
	})
}

type OperatorTestSuite struct {
	*testsuite.WPgxTestSuite
	gater *connectionGater

	operatorRepo  OperatorRepo
	blacklistRepo BlacklistRepo

	RedisConn redis.UniversalClient
	FreeCache *freecache.Cache
	DCache    *dcache.DCache
}

func TestOperatorSuite(t *testing.T) {
	suite.Run(t, newOperatorTestSuite())
}

func (s *OperatorTestSuite) SetupTest() {
	s.WPgxTestSuite.SetupTest()
	s.Require().NoError(s.RedisConn.FlushAll(context.Background()).Err())
	s.FreeCache.Clear()

	pool := s.GetPool()
	s.Require().NotNil(pool, "Database pool should not be nil")
	conn := pool.WConn()
	s.Require().NotNil(conn, "Database connection should not be nil")

	operatorRepo := operators.New(conn, s.DCache)
	s.Require().NotNil(operatorRepo, "Operator repository should not be nil")
	s.operatorRepo = operatorRepo
	blacklistRepo := blacklist.New(conn, s.DCache)
	s.Require().NotNil(blacklistRepo, "Blacklist repository should not be nil")
	s.blacklistRepo = blacklistRepo

	gater, err := NewConnectionGater(&Config{
		BlacklistRepo: s.blacklistRepo,
		OperatorRepo:  s.operatorRepo,
	})
	s.Require().NoError(err, "Failed to create event listener")
	s.gater = gater.(*connectionGater)
}

func newOperatorTestSuite() *OperatorTestSuite {
	redisClient := redis.NewClient(&redis.Options{
		Addr:        "127.0.0.1:6379",
		ReadTimeout: 3 * time.Second,
		PoolSize:    50,
		Password:    "",
	})

	if redisClient.Ping(context.Background()).Err() != nil {
		panic(fmt.Errorf("redis connection failed to ping"))
	}

	memCache := freecache.NewCache(100 * 1024 * 1024)

	dCache, err := dcache.NewDCache(
		"test", redisClient, memCache, 100*time.Millisecond, true, true)
	if err != nil {
		panic(err)
	}

	s := &OperatorTestSuite{
		WPgxTestSuite: testsuite.NewWPgxTestSuiteFromEnv("operator_test", []string{
			tasks.Schema,
			consensus_responses.Schema,
			blacklist.Schema,
			operators.Schema,
		}),
		RedisConn: redisClient,
		FreeCache: memCache,
		DCache:    dCache,
	}

	return s
}

// Test cases for gater
func (s *OperatorTestSuite) TestGaterInitialization() {
	// Test successful initialization
	gater, err := NewConnectionGater(&Config{
		BlacklistRepo: s.blacklistRepo,
		OperatorRepo:  s.operatorRepo,
	})
	s.Require().NoError(err)
	s.NotNil(gater)

	// Test initialization with nil config
	gater, err = NewConnectionGater(nil)
	s.Error(err)
	s.Nil(gater)

	// Test initialization with missing BlacklistRepo
	gater, err = NewConnectionGater(&Config{
		OperatorRepo: s.operatorRepo,
	})
	s.Error(err)
	s.Nil(gater)

	// Test initialization with missing OperatorRepo
	gater, err = NewConnectionGater(&Config{
		BlacklistRepo: s.blacklistRepo,
	})
	s.Error(err)
	s.Nil(gater)
}

// TestPeerPermissions tests the peer permission checking functionality
func (s *OperatorTestSuite) TestPeerPermissions() {
	// Create test peer IDs
	testCases := []struct {
		name          string
		setupState    func(s *OperatorTestSuite)
		peerID        string
		expectedAllow bool
		validateState func(s *OperatorTestSuite)
		goldenFile    string
	}{
		{
			name: "active_operator_not_blocked",
			setupState: func(s *OperatorTestSuite) {
				operatorSerde := operatorStatesTableSerde{
					operatorStatesQuerier: s.operatorRepo.(*operators.Queries),
				}
				s.LoadState("TestOperatorSuite/TestPeerPermissions/active_operator.json", operatorSerde)
			},
			peerID:        "12D3KooWNbUurxoy5Qn7hSRi5dvMdaeEFZQavacg253npoiuSJ9p",
			expectedAllow: true,
		},
		{
			name: "inactive_operator_not_blocked",
			setupState: func(s *OperatorTestSuite) {
				operatorSerde := operatorStatesTableSerde{
					operatorStatesQuerier: s.operatorRepo.(*operators.Queries),
				}
				s.LoadState("TestOperatorSuite/TestPeerPermissions/inactive_operator.json", operatorSerde)
			},
			peerID:        "12D3KooWNbUurxoy5Qn7hSRi5dvMdaeEFZQavacg253npoiuSJ9p",
			expectedAllow: false,
		},
		{
			name: "active_operator_blocked",
			setupState: func(s *OperatorTestSuite) {
				operatorSerde := operatorStatesTableSerde{
					operatorStatesQuerier: s.operatorRepo.(*operators.Queries),
				}
				s.LoadState("TestOperatorSuite/TestPeerPermissions/blocked_operator.json", operatorSerde)

				blacklistSerde := blacklistStatesTableSerde{
					blacklistQuerier: s.blacklistRepo.(*blacklist.Queries),
				}
				s.LoadState("TestOperatorSuite/TestPeerPermissions/blacklist.json", blacklistSerde)
			},
			peerID:        "12D3KooWNbUurxoy5Qn7hSRi5dvMdaeEFZQavacg253npoiuSJ9p",
			expectedAllow: false,
		},
		{
			name: "inactive_operator_blocked",
			setupState: func(s *OperatorTestSuite) {
				operatorSerde := operatorStatesTableSerde{
					operatorStatesQuerier: s.operatorRepo.(*operators.Queries),
				}
				s.LoadState("TestOperatorSuite/TestPeerPermissions/inactive_blocked_operator.json", operatorSerde)

				blacklistSerde := blacklistStatesTableSerde{
					blacklistQuerier: s.blacklistRepo.(*blacklist.Queries),
				}
				s.LoadState("TestOperatorSuite/TestPeerPermissions/blacklist.json", blacklistSerde)
			},
			peerID:        "12D3KooWNbUurxoy5Qn7hSRi5dvMdaeEFZQavacg253npoiuSJ9p",
			expectedAllow: false,
		},
		{
			name: "active_operator_blocked_expires",
			setupState: func(s *OperatorTestSuite) {
				operatorSerde := operatorStatesTableSerde{
					operatorStatesQuerier: s.operatorRepo.(*operators.Queries),
				}
				s.LoadState("TestOperatorSuite/TestPeerPermissions/blocked_operator.json", operatorSerde)

				blacklistSerde := blacklistStatesTableSerde{
					blacklistQuerier: s.blacklistRepo.(*blacklist.Queries),
				}
				s.LoadState("TestOperatorSuite/TestPeerPermissions/blacklist_expires.json", blacklistSerde)
			},
			peerID:        "12D3KooWNbUurxoy5Qn7hSRi5dvMdaeEFZQavacg253npoiuSJ9p",
			expectedAllow: true,
		},
		{
			name: "inactive_operator_blocked_expires",
			setupState: func(s *OperatorTestSuite) {
				operatorSerde := operatorStatesTableSerde{
					operatorStatesQuerier: s.operatorRepo.(*operators.Queries),
				}
				s.LoadState("TestOperatorSuite/TestPeerPermissions/inactive_blocked_operator.json", operatorSerde)

				blacklistSerde := blacklistStatesTableSerde{
					blacklistQuerier: s.blacklistRepo.(*blacklist.Queries),
				}
				s.LoadState("TestOperatorSuite/TestPeerPermissions/blacklist_expires.json", blacklistSerde)
			},
			peerID:        "12D3KooWNbUurxoy5Qn7hSRi5dvMdaeEFZQavacg253npoiuSJ9p",
			expectedAllow: false,
		},
		{
			name:          "non_existent_operator",
			setupState:    func(s *OperatorTestSuite) {},
			peerID:        "12D3KooWNbUurxoy5Qn7hSRi5dvMdaeEFZQavacg253npoiuSJ9p",
			expectedAllow: false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			// Setup test state
			tc.setupState(s)

			// Create peer ID
			peerID, err := peer.Decode(tc.peerID)
			s.Require().NoError(err)

			// Test all gating methods
			s.Equal(tc.expectedAllow, s.gater.InterceptPeerDial(peerID))
			s.Equal(tc.expectedAllow, s.gater.InterceptAddrDial(peerID, nil))
			s.Equal(tc.expectedAllow, s.gater.InterceptSecured(network.DirInbound, peerID, nil))

			// InterceptUpgraded always returns true
			allow, _ := s.gater.InterceptUpgraded(nil)
			s.True(allow)

			if tc.validateState != nil {
				tc.validateState(s)
			}

			if tc.goldenFile != "" {
				operatorSerde := operatorStatesTableSerde{
					operatorStatesQuerier: s.operatorRepo.(*operators.Queries),
				}
				s.Golden(tc.goldenFile, operatorSerde)
			}
		})
	}
}

// TestAddressValidation tests the multiaddr validation functionality
func (s *OperatorTestSuite) TestAddressValidation() {
	testCases := []struct {
		name        string
		addr        string
		expectValid bool
	}{
		{
			name:        "valid_ipv4_tcp",
			addr:        "/ip4/127.0.0.1/tcp/1234",
			expectValid: true,
		},
		{
			name:        "valid_ipv4_tcp_min_port",
			addr:        "/ip4/127.0.0.1/tcp/1",
			expectValid: true,
		},
		{
			name:        "valid_ipv4_tcp_max_port",
			addr:        "/ip4/127.0.0.1/tcp/65535",
			expectValid: true,
		},
		{
			name:        "valid_ipv6_tcp",
			addr:        "/ip6/::1/tcp/1234",
			expectValid: false,
		},
		{
			name:        "invalid_ip",
			addr:        "/ip4/256.256.256.256/tcp/1234",
			expectValid: false,
		},
		{
			name:        "invalid_port_zero",
			addr:        "/ip4/127.0.0.1/tcp/0",
			expectValid: false,
		},
		{
			name:        "invalid_port_negative",
			addr:        "/ip4/127.0.0.1/tcp/-1",
			expectValid: false,
		},
		{
			name:        "invalid_port_too_large",
			addr:        "/ip4/127.0.0.1/tcp/65536",
			expectValid: false,
		},
		{
			name:        "missing_tcp",
			addr:        "/ip4/127.0.0.1",
			expectValid: false,
		},
		{
			name:        "missing_ip",
			addr:        "/tcp/1234",
			expectValid: false,
		},
		{
			name:        "additional_protocol",
			addr:        "/ip4/127.0.0.1/udp/1234/tcp/1234",
			expectValid: true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			addr, err := ma.NewMultiaddr(tc.addr)
			if err != nil {
				s.False(tc.expectValid)
				return
			}

			s.Equal(tc.expectValid, s.gater.InterceptAccept(&mockConnMultiaddrs{
				remoteAddr: addr,
			}))
		})
	}
}

// mockConnMultiaddrs implements network.ConnMultiaddrs for testing
type mockConnMultiaddrs struct {
	remoteAddr ma.Multiaddr
}

func (m *mockConnMultiaddrs) RemoteMultiaddr() ma.Multiaddr {
	return m.remoteAddr
}

func (m *mockConnMultiaddrs) LocalMultiaddr() ma.Multiaddr {
	// For testing purposes, we return the same address
	return m.remoteAddr
}

// TestConcurrentConnections tests concurrent connection attempts
func (s *OperatorTestSuite) TestConcurrentConnections() {
	const numConnections = 10
	var wg sync.WaitGroup

	// Load active operator state
	operatorSerde := operatorStatesTableSerde{
		operatorStatesQuerier: s.operatorRepo.(*operators.Queries),
	}
	s.LoadState("TestOperatorSuite/TestConcurrentConnections/active_operator.json", operatorSerde)

	peerID, err := peer.Decode("12D3KooWNbUurxoy5Qn7hSRi5dvMdaeEFZQavacg253npoiuSJ9p")
	s.Require().NoError(err)

	for i := 0; i < numConnections; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.True(s.gater.InterceptPeerDial(peerID))
		}()
	}

	wg.Wait()
}

// TestInvalidPeerID tests handling of invalid peer IDs
func (s *OperatorTestSuite) TestInvalidPeerID() {
	// Test with invalid peer ID format
	invalidPeerID := "invalid-peer-id"
	peerID, err := peer.Decode(invalidPeerID)
	s.Error(err)

	// Even with error in peer.Decode, test the gater methods
	s.False(s.gater.InterceptPeerDial(peerID))
	s.False(s.gater.InterceptAddrDial(peerID, nil))
	s.False(s.gater.InterceptSecured(network.DirInbound, peerID, nil))
}

// TestNilMultiaddr tests handling of nil multiaddr
func (s *OperatorTestSuite) TestNilMultiaddr() {
	s.False(s.gater.InterceptAccept(&mockConnMultiaddrs{
		remoteAddr: nil,
	}))
}

// TestRaceConditions tests for race conditions in concurrent access
func (s *OperatorTestSuite) TestRaceConditions() {
	const numGoroutines = 100
	var wg sync.WaitGroup

	// Setup test data
	operatorSerde := operatorStatesTableSerde{
		operatorStatesQuerier: s.operatorRepo.(*operators.Queries),
	}
	s.LoadState("TestOperatorSuite/TestPeerPermissions/active_operator.json", operatorSerde)

	peerID, err := peer.Decode("12D3KooWNbUurxoy5Qn7hSRi5dvMdaeEFZQavacg253npoiuSJ9p")
	s.Require().NoError(err)

	// Test concurrent access to all gater methods
	for i := 0; i < numGoroutines; i++ {
		wg.Add(3)
		go func() {
			defer wg.Done()
			s.True(s.gater.InterceptPeerDial(peerID))
		}()
		go func() {
			defer wg.Done()
			s.True(s.gater.InterceptAddrDial(peerID, nil))
		}()
		go func() {
			defer wg.Done()
			s.True(s.gater.InterceptSecured(network.DirInbound, peerID, nil))
		}()
	}

	wg.Wait()
}
