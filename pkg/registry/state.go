package registry

import (
	"context"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	utils "github.com/galxe/spotted-network/pkg/common"
	"github.com/galxe/spotted-network/pkg/common/types"
	"github.com/galxe/spotted-network/pkg/repos/registry/operators"
	pb "github.com/galxe/spotted-network/proto"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/libp2p/go-libp2p/core/peer"
)

//  syncs operator states from database and handles inactive operators
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
			n.DisconnectPeer(peerID)
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
	status, logMsg := DetermineOperatorStatus(currentBlock, activeEpoch, operator.ExitEpoch)
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

// getActiveOperators returns list of active operators
func (n *Node) getActiveOperators() []*pb.OperatorPeerState {
	// get the registry node info
	registryInfo := peer.AddrInfo{
		ID:    n.host.ID(),
		Addrs: n.host.Addrs(),
	}

	// create the active operators list, include the registry
	activeOperators := make([]*pb.OperatorPeerState, 0)
	
	// add the registry info
	registryOp := &pb.OperatorPeerState{
		PeerId:     registryInfo.ID.String(),
		Multiaddrs: utils.MultiaddrsToStrings(registryInfo.Addrs),
	}
	activeOperators = append(activeOperators, registryOp)

	// add all active operators
	n.activeOperators.mu.RLock()
	for _, state := range n.activeOperators.active {
		op := &pb.OperatorPeerState{
			PeerId:     state.PeerID.String(),
			Multiaddrs: utils.MultiaddrsToStrings(state.Multiaddrs),
			Address:    state.Address,
		}
		activeOperators = append(activeOperators, op)
	}
	n.activeOperators.mu.RUnlock()

	return activeOperators
}