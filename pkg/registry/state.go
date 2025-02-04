package registry

import (
	"context"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/galxe/spotted-network/pkg/common/types"
	"github.com/galxe/spotted-network/pkg/repos/registry/operators"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/libp2p/go-libp2p/core/peer"
)

// syncs operator states from database and handles inactive operators
func (n *Node) syncPeerInfo(ctx context.Context) error {
	n.activeOperators.mu.RLock()
	defer n.activeOperators.mu.RUnlock()

	log.Printf("[State] Starting to sync operator states for %d peers", len(n.activeOperators.active))

	for peerID, info := range n.activeOperators.active {
		// Skip if no address associated
		if info.Address == "" {
			log.Printf("[State] Skipping peer %s: no address associated", peerID)
			continue
		}

		// Query operator state from database
		operator, err := n.opQuerier.GetOperatorByAddress(ctx, info.Address)
		if err != nil {
			log.Printf("[State] Failed to query operator %s state: %v", info.Address, err)
			continue
		}

		// If operator is inactive, disconnect
		if operator.Status != "active" {
			log.Printf("[State] Operator %s (peer %s) is inactive, disconnecting", info.Address, peerID)
			n.disconnectPeer(peerID)
			continue
		}

	}

	return nil
}

// updateStatusAfterOperations updates operator status based on current block number and epochs
func (n *Node) updateSingleOperatorState(ctx context.Context, operatorAddr string) error {

	// Get current block number
	currentBlock, err := n.mainnetClient.BlockNumber(ctx)
	if err != nil {
		return fmt.Errorf("[EventListener] failed to get current block number: %w", err)
	}

	// Get operator from database to get active_epoch and exit_epoch
	operator, err := n.opQuerier.GetOperatorByAddress(ctx, operatorAddr)
	if err != nil {
		return fmt.Errorf("[EventListener] failed to get operator from database: %w", err)
	}

	activeEpoch := operator.ActiveEpoch

	// Determine operator status using helper
	status, logMsg := determineOperatorStatus(currentBlock, activeEpoch, operator.ExitEpoch)
	log.Printf("[EventListener] %s", logMsg)

	// Get weight if status is active
	var weight *big.Int
	if status == types.OperatorStatusActive {
		weight, err = n.mainnetClient.GetOperatorWeight(ctx, common.HexToAddress(operatorAddr))
		if err != nil {
			return fmt.Errorf("[EventListener] failed to get operator weight: %w", err)
		}
		log.Printf("[EventListener] Got weight for operator %s with status %s: %s", operatorAddr, status, weight.String())
	} else {
		weight = big.NewInt(0)
		log.Printf("[EventListener] Set weight to 0 for operator %s with status %s", operatorAddr, status)
	}

	// Update operator state in database
	_, err = n.opQuerier.UpdateOperatorState(ctx, operators.UpdateOperatorStateParams{
		Address: operatorAddr,
		Status:  status,
		Weight: pgtype.Numeric{
			Int:    weight,
			Valid:  true,
			Exp:    0,
		},
	}, &operatorAddr)
	if err != nil {
		return fmt.Errorf("failed to update operator state: %w", err)
	}

	log.Printf("[EventListener] Updated operator %s status to %s with weight %s", operatorAddr, status, weight.String())
	return nil
}


// GetOperatorState returns the state of a connected operator
func (n *Node) GetOperatorState(id peer.ID) *OperatorPeerInfo {
	n.activeOperators.mu.RLock()
	defer n.activeOperators.mu.RUnlock()
	return n.activeOperators.active[id]
}

// GetConnectedOperators returns all connected operator IDs
func (n *Node) getActivePeerIDs() []peer.ID {
	n.activeOperators.mu.RLock()
	defer n.activeOperators.mu.RUnlock()

	operators := make([]peer.ID, 0, len(n.activeOperators.active))
	for id := range n.activeOperators.active {
		operators = append(operators, id)
	}
	return operators
}

// UpdateOperatorState updates an operator's state
func (n *Node) UpdateOperatorState(id peer.ID, state *OperatorPeerInfo) {
	n.activeOperators.mu.Lock()
	defer n.activeOperators.mu.Unlock()
	n.activeOperators.active[id] = state
}

// RemoveOperator removes an operator from the active set
func (n *Node) RemoveOperator(id peer.ID) {
	n.activeOperators.mu.Lock()
	defer n.activeOperators.mu.Unlock()
	delete(n.activeOperators.active, id)
}