package common

import (
	"bytes"
	"crypto/sha256"
	"log"
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
		return &MerkleTree{nodes: [][]byte{nil}}
	}

	// Get leaf hashes
	leaves := make([][]byte, len(items))
	for i, item := range items {
		leaves[i] = item.Hash()
	}

	// Initialize tree with leaves
	tree := &MerkleTree{
		nodes: make([][]byte, len(leaves)),
	}
	copy(tree.nodes, leaves)

	offset := 0
	levelSize := len(leaves)

	// Build tree level by level
	for levelSize > 1 {
		nextLevelSize := (levelSize + 1) / 2 // Handles odd number of nodes

		for i := 0; i < levelSize; i += 2 {
			// Get left child
			left := tree.nodes[offset+i]
			
			// Get right child if it exists, otherwise use left child
			var right []byte
			if i+1 < levelSize {
				right = tree.nodes[offset+i+1]
			} else {
				right = left
			}

			// Combine and hash
			combined := append(left, right...)
			hash := sha256.Sum256(combined)
			tree.nodes = append(tree.nodes, hash[:])
		}

		offset += levelSize
		levelSize = nextLevelSize
	}
	log.Printf("[MerkleTree] Final tree size: %d, Root hash: %x", len(tree.nodes), tree.nodes[len(tree.nodes)-1])
	return tree
}

// RootBytes returns the root hash of the tree
func (t *MerkleTree) RootBytes() []byte {
	if t == nil || len(t.nodes) == 0 {
		return nil
	}
	log.Printf("[MerkleTree] RootBytes: tree size=%d, root=%x", 
		len(t.nodes), t.nodes[len(t.nodes)-1])
	return t.nodes[len(t.nodes)-1]
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
