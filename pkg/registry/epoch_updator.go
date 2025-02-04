package registry

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"runtime/debug"
	"sync"
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


// TransactionManager abstracts database transaction operations
type TxManager interface {
	Transact(ctx context.Context, txOptions pgx.TxOptions, f func(context.Context, *wpgx.WTx) (any, error)) (any, error)
}

type EpochUpdatorConfig struct {
	node *Node

	opQuerier OperatorsQuerier
	mainnetClient MainnetClient
	txManager TxManager 
	pubsub PubSubService
}

type EpochUpdator struct {
	node *Node

	opQuerier OperatorsQuerier
	mainnetClient MainnetClient
	txManager TxManager 
	pubsub PubSubService

	lastProcessedEpoch uint32

	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func NewEpochUpdator(ctx context.Context, cfg *EpochUpdatorConfig) (*EpochUpdator, error) {
	if cfg.node == nil {
		log.Fatal("node is nil")
	}
	if cfg.opQuerier == nil {
		log.Fatal("opQuerier is nil")
	}
	if cfg.pubsub == nil {
		log.Fatal("pubsub is nil")
	}
	if cfg.mainnetClient == nil {
		log.Fatal("mainnetClient is nil")
	}
	if cfg.txManager == nil {
		log.Fatal("txManager is nil")
	}
	e := &EpochUpdator{
		node: cfg.node,
		opQuerier: cfg.opQuerier,
		pubsub: cfg.pubsub,
		mainnetClient: cfg.mainnetClient,
		txManager: cfg.txManager, 
	}

	ctx, cancel := context.WithCancel(ctx)
	e.cancel = cancel
	e.wg.Add(1)
	
	go func() {
		defer e.wg.Done()
		if err := e.start(ctx); err != nil {
			log.Printf("[Epoch] Epoch monitoring failed: %v", err)
		}
	}()
	
	log.Printf("[Epoch] Epoch monitoring started")
	
	return e, nil
}

// starts monitoring epoch updates
func (e *EpochUpdator) start(ctx context.Context) error {
	ticker := time.NewTicker(epochMonitorInterval)
	defer ticker.Stop()

	log.Printf("[Epoch] Starting epoch monitoring with interval %v...", epochMonitorInterval)

	defer func() {
		if r := recover(); r != nil {
			log.Printf("[Epoch] Recovered from panic: %v\nStack: %s", r, debug.Stack())
		}
	}()

	for {
		select {
		case <-ctx.Done():
			log.Printf("[Epoch] Stopping epoch monitoring: %v", ctx.Err())
			return nil
		case <-ticker.C:
			// add timeout control for each update
			updateCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
			
			func() {
				defer cancel() 

				// Get latest block number
				blockNumber, err := e.mainnetClient.BlockNumber(updateCtx)
				if err != nil {
					log.Printf("[Epoch] Failed to get latest block number: %v", err)
					return
				}

				// Calculate current epoch
				currentEpoch := utils.CalculateEpochNumber(blockNumber)

				// If we're in a new epoch, update operator states
				if currentEpoch > e.lastProcessedEpoch {
					log.Printf("[Epoch] Starting update for epoch %d (current: %d)", 
						currentEpoch, e.lastProcessedEpoch)

					if err := e.handleEpochUpdate(updateCtx, currentEpoch); err != nil {
						log.Printf("[Epoch] Failed to update operator states for epoch %d: %v", 
							currentEpoch, err)
						return
					}
					
					e.lastProcessedEpoch = currentEpoch
					log.Printf("[Epoch] Successfully updated operator states for epoch %d", currentEpoch)
				}
			}()
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

		if err := e.node.sp.broadcastStateUpdate(&currentEpoch); err != nil {
			return nil, fmt.Errorf("failed to broadcast state update: %w", err)
		}

		return nil, nil
	})

	if err != nil {
		return fmt.Errorf("transaction failed: %w", err)
	}

	return nil
}	

func (e *EpochUpdator) Stop() {
	e.cancel()
	
	e.wg.Wait()
	
	log.Printf("[Epoch] Epoch monitoring stopped")
}	

