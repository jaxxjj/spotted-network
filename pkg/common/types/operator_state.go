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
	if s.Multiaddrs != nil {
		validAddrs := make([]string, 0, len(s.Multiaddrs))
		for _, addr := range s.Multiaddrs {
			if addr != nil {
				validAddrs = append(validAddrs, addr.String())
			}
		}
		sort.Strings(validAddrs)
		for _, addr := range validAddrs {
			buf.WriteString(addr)
		}
	}

	buf.WriteString(s.Address)
	buf.WriteString(s.SigningKey)
	
	if s.Weight != nil && s.Weight.Sign() >= 0 {
		buf.WriteString(s.Weight.String())
	}

	// Compute SHA256
	hash := sha256.Sum256(buf.Bytes())
	return hash[:]
}

// ComputeStateRoot computes the merkle root hash of a list of operators
func ComputeStateRoot(operators []*OperatorState) []byte {
	if len(operators) == 0 {
		return nil
	}

	// Sort operators by PeerID
	ops := OperatorStates(operators)
	sort.Sort(ops)

	// Convert to []Hashable and filter invalid operators
	hashables := make([]common.Hashable, 0, len(ops))
	for _, op := range ops {
		// Skip invalid operators
		if op == nil || op.Weight == nil || op.Weight.Sign() < 0 {
			continue
		}
		if err := op.PeerID.Validate(); err != nil {
			continue
		}
		
		hashables = append(hashables, op)
	}

	if len(hashables) == 0 {
		return nil
	}

	// Create merkle tree and return root
	tree := common.NewMerkleTree(hashables)
	if tree == nil {
		return nil
	}

	return tree.RootBytes()
}

// VerifyStateRoot verifies if the given root matches the computed root
func VerifyStateRoot(operators []*OperatorState, root []byte) bool {
	return bytes.Equal(ComputeStateRoot(operators), root)
} 