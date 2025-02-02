package types

import (
	"bytes"
	"crypto/sha256"
	"math/big"
	"sort"

	"github.com/galxe/spotted-network/pkg/common"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

// OperatorState represents the complete state of an operator
type OperatorState struct {
	// Peer Info
	PeerID     peer.ID
	Multiaddrs []multiaddr.Multiaddr

	// Business State
	Address    string
	SigningKey string
	Weight     *big.Int
}

// OperatorStates is a sortable slice of OperatorState
type OperatorStates []*OperatorState

func (s OperatorStates) Len() int { return len(s) }
func (s OperatorStates) Less(i, j int) bool {
	return s[i].PeerID.String() < s[j].PeerID.String()
}
func (s OperatorStates) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

// Hash implements the Hashable interface
func (s *OperatorState) Hash() []byte {
	var buf bytes.Buffer

	// Write fields in a fixed order
	buf.WriteString(s.PeerID.String())
	
	// Sort multiaddrs for consistency
	addrs := make([]string, len(s.Multiaddrs))
	for i, addr := range s.Multiaddrs {
		addrs[i] = addr.String()
	}
	sort.Strings(addrs)
	for _, addr := range addrs {
		buf.WriteString(addr)
	}

	buf.WriteString(s.Address)
	buf.WriteString(s.SigningKey)
	buf.WriteString(s.Weight.String())

	// Compute SHA256
	hash := sha256.Sum256(buf.Bytes())
	return hash[:]
}

// ComputeStateRoot computes the merkle root hash of a list of operators
func ComputeStateRoot(operators []*OperatorState) []byte {
	// Sort operators by PeerID
	ops := OperatorStates(operators)
	sort.Sort(ops)

	// Convert to []Hashable
	hashables := make([]common.Hashable, len(ops))
	for i, op := range ops {
		hashables[i] = op
	}

	// Create merkle tree
	tree := common.NewMerkleTree(hashables)
	return tree.RootBytes()
}

// VerifyStateRoot verifies if the given root matches the computed root
func VerifyStateRoot(operators []*OperatorState, root []byte) bool {
	return bytes.Equal(ComputeStateRoot(operators), root)
} 