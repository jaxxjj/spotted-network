package common

import (
	"bytes"
	"crypto/sha256"
)

// Hashable is an interface for types that can be hashed in a merkle tree
type Hashable interface {
	// Hash returns a deterministic hash of the object
	Hash() []byte
}

// MerkleTree represents a merkle tree data structure
type MerkleTree struct {
	// nodes[0] is empty
	// nodes[1] is the root
	// nodes[2*i] and nodes[2*i+1] are children of nodes[i]
	nodes [][]byte
}

// NewMerkleTree creates a new merkle tree from a list of hashable items
func NewMerkleTree(items []Hashable) *MerkleTree {
	if len(items) == 0 {
		return &MerkleTree{nodes: [][]byte{nil, nil}}
	}

	// Get leaf hashes
	leaves := make([][]byte, len(items))
	for i, item := range items {
		leaves[i] = item.Hash()
	}

	// Calculate tree size
	size := 1 // for empty first element
	for l := len(leaves); l > 0; l = (l + 1) / 2 {
		size += l
	}

	// Initialize tree
	tree := &MerkleTree{
		nodes: make([][]byte, size),
	}

	// Copy leaves
	offset := size - len(leaves)
	copy(tree.nodes[offset:], leaves)

	// Build internal nodes
	for i := offset - 1; i > 0; i-- {
		left := tree.nodes[i*2]
		right := tree.nodes[i*2+1]
		if right == nil {
			right = left
		}
		combined := append(left, right...)
		hash := sha256.Sum256(combined)
		tree.nodes[i] = hash[:]
	}

	return tree
}

// RootBytes returns the merkle root hash as bytes
func (t *MerkleTree) RootBytes() []byte {
	if len(t.nodes) < 2 {
		return nil
	}
	return t.nodes[1]
}

// Verify verifies if the given hash matches the merkle root
func (t *MerkleTree) Verify(hash []byte) bool {
	return bytes.Equal(t.RootBytes(), hash)
}

// HashableBytes is a wrapper for []byte to implement Hashable
type HashableBytes []byte

func (h HashableBytes) Hash() []byte {
	hash := sha256.Sum256(h)
	return hash[:]
}
