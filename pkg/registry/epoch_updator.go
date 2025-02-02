package registry

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"time"

	ethcommon "github.com/ethereum/go-ethereum/common"
	utils "github.com/galxe/spotted-network/pkg/common"
	"github.com/galxe/spotted-network/pkg/repos/registry/operators"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stumble/wpgx"
)

const (

	GenesisBlock = 0     
	EpochPeriod  = 12   

	epochMonitorInterval = 5 * time.Second
)

type EpochUpdatorNode interface {
	syncPeerInfo(ctx context.Context) error 
}
type EpochUpdatorQuerier interface {
	GetOperatorByAddress(ctx context.Context, address string) (*operators.Operators, error)
	ListAllOperators(ctx context.Context) ([]operators.Operators, error)
	UpdateOperatorState(ctx context.Context, arg operators.UpdateOperatorStateParams, getOperatorByAddress *string) (*operators.Operators, error) 
	WithTx(tx *wpgx.WTx) *operators.Queries
}

type EpochUpdatorChainClient interface {
	BlockNumber(ctx context.Context) (uint64, error)
	GetOperatorWeight(ctx context.Context, address ethcommon.Address) (*big.Int, error)
	GetOperatorSigningKey(ctx context.Context, address ethcommon.Address, epoch uint32) (ethcommon.Address, error)
}

type EpochUpdatorSP interface {
	BroadcastStateUpdate() error
}

// TransactionManager abstracts database transaction operations
type TxManager interface {
	Transact(ctx context.Context, txOptions pgx.TxOptions, f func(context.Context, *wpgx.WTx) (any, error)) (any, error)
}

type EpochUpdatorConfig struct {
	node EpochUpdatorNode
	opQuerier EpochUpdatorQuerier
	pubsub PubSubService
	sp EpochUpdatorSP
	mainnetClient EpochUpdatorChainClient
	txManager TxManager // Replace dbPool with txManager
}

type EpochUpdator struct {
	node EpochUpdatorNode
	opQuerier EpochUpdatorQuerier
	lastProcessedEpoch uint32
	pubsub PubSubService
	sp EpochUpdatorSP
	mainnetClient EpochUpdatorChainClient
	txManager TxManager // Replace dbPool with txManager
}

func NewEpochUpdator(ctx context.Context, cfg *EpochUpdatorConfig) (*EpochUpdator, error) {
	e := &EpochUpdator{
		node: cfg.node,
		opQuerier: cfg.opQuerier,
		pubsub: cfg.pubsub,
		mainnetClient: cfg.mainnetClient,
		sp: cfg.sp,
		txManager: cfg.txManager, // Use txManager instead of dbPool
	}
	
	go e.start(ctx)
	log.Printf("[Epoch] Epoch monitoring started")
	
	return e, nil
}

// starts monitoring epoch updates
func (e *EpochUpdator) start(ctx context.Context) error {
	ticker := time.NewTicker(epochMonitorInterval)
	defer ticker.Stop()

	log.Printf("[Epoch] Starting epoch monitoring...")

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			// Get latest block number
			blockNumber, err := e.mainnetClient.BlockNumber(ctx)
			if err != nil {
				log.Printf("[Epoch] Failed to get latest block number: %v", err)
				continue
			}

			// Calculate current epoch
			currentEpoch := calculateEpochNumber(blockNumber)

			// If we're in a new epoch, update operator states
			if currentEpoch > e.lastProcessedEpoch {
				if err := e.handleEpochUpdate(ctx, currentEpoch); err != nil {
					log.Printf("[Epoch] Failed to update operator states for epoch %d: %v", currentEpoch, err)
					continue
				}
				e.lastProcessedEpoch = currentEpoch
				log.Printf("[Epoch] Updated operator states for epoch %d", currentEpoch)
			}
		}
	}
}

func (e *EpochUpdator) handleEpochUpdate(ctx context.Context, currentEpoch uint32) error {
	// Use txManager instead of dbPool
	currentBlock, err := e.mainnetClient.BlockNumber(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current block number: %w", err)
	}

	_, err = e.txManager.Transact(ctx, pgx.TxOptions{}, func(ctx context.Context, tx *wpgx.WTx) (any, error) {
		txQuerier := e.opQuerier.WithTx(tx)

		// Get all operators using transaction
		allOperators, err := txQuerier.ListAllOperators(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get all operators: %w", err)
		}

		log.Printf("[Epoch] Processing %d operators for epoch %d (block %d)", len(allOperators), currentEpoch, currentBlock)

		// Update operator states in DB
		for _, operator := range allOperators {
			activeEpoch := operator.ActiveEpoch
			exitEpoch := operator.ExitEpoch
			
			// Use helper function to determine status
			newStatus, logMsg := DetermineOperatorStatus(currentBlock, activeEpoch, exitEpoch)
			log.Printf("[Node] %s", logMsg)

			var currentWeight *big.Int
			var weightNum pgtype.Numeric
			var currentSigningKey ethcommon.Address

			// Only get weight and signing key if operator will be active
			if newStatus == "active" {
				currentWeight, err = e.mainnetClient.GetOperatorWeight(ctx, ethcommon.HexToAddress(operator.Address))
				if err != nil {
					log.Printf("[Epoch] Failed to get weight for operator %s: %v", operator.Address, err)
					return nil, fmt.Errorf("failed to get weight for operator %s: %w", operator.Address, err)
				}
				
				currentSigningKey, err = e.mainnetClient.GetOperatorSigningKey(ctx, ethcommon.HexToAddress(operator.Address), currentEpoch)
				if err != nil {
					log.Printf("[Epoch] Failed to get signing key for operator %s: %v", operator.Address, err)
					return nil, fmt.Errorf("failed to get signing key for operator %s: %w", operator.Address, err)
				}

				weightNum = utils.BigIntToNumeric(currentWeight)
			} else {
				// If not active, set weight to 0
				weightNum = utils.BigIntToNumeric(big.NewInt(0))
			}

			// Update operator state in database using transaction
			_, err = txQuerier.UpdateOperatorState(ctx, operators.UpdateOperatorStateParams{
				Address: operator.Address,
				Status:  newStatus,
				Weight:  weightNum,
				SigningKey: currentSigningKey.Hex(),
			}, &operator.Address)

			if err != nil {
				log.Printf("[Epoch] Failed to update operator %s state: %v", operator.Address, err)
				return nil, fmt.Errorf("failed to update operator %s state: %w", operator.Address, err)
			}
		}

		// Sync peer info and broadcast state update within transaction
		if err := e.node.syncPeerInfo(ctx); err != nil {
			return nil, fmt.Errorf("failed to sync peer info: %w", err)
		}

		if err := e.sp.BroadcastStateUpdate(); err != nil {
			return nil, fmt.Errorf("failed to broadcast state update: %w", err)
		}

		return nil, nil
	})

	if err != nil {
		return fmt.Errorf("transaction failed: %w", err)
	}

	return nil
}	

