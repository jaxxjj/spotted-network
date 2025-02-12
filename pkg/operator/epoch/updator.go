package epoch

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
	"github.com/galxe/spotted-network/pkg/repos/operators"
	"github.com/jackc/pgx/v5"
	"github.com/stumble/wpgx"
)

const (
	epochMonitorInterval = 5 * time.Second
)

// TransactionManager abstracts database transaction operations
type TxManager interface {
	Transact(ctx context.Context, txOptions pgx.TxOptions, f func(context.Context, *wpgx.WTx) (any, error)) (any, error)
}

type OperatorRepo interface {
	ListAllOperators(ctx context.Context) ([]operators.Operators, error)
	WithTx(tx *wpgx.WTx) *operators.Queries
	UpdateOperatorState(ctx context.Context, arg operators.UpdateOperatorStateParams, getOperatorByAddress *string, getOperatorBySigningKey *string, getOperatorByP2PKey *string, isActiveOperator *string) (*operators.Operators, error)
}

type MainnetClient interface {
	BlockNumber(ctx context.Context) (uint64, error)
	GetOperatorWeight(ctx context.Context, address ethcommon.Address) (*big.Int, error)
	GetOperatorSigningKey(ctx context.Context, operator ethcommon.Address, epoch uint32) (ethcommon.Address, error)
	GetOperatorP2PKey(ctx context.Context, operator ethcommon.Address, epoch uint32) (ethcommon.Address, error)
	GetMinimumWeight(ctx context.Context) (*big.Int, error)
	GetTotalWeight(ctx context.Context) (*big.Int, error)
	GetThresholdWeight(ctx context.Context) (*big.Int, error)
	Close() error
}

type EpochState struct {
	EpochNumber     uint32
	MinimumWeight   *big.Int
	TotalWeight     *big.Int
	ThresholdWeight *big.Int
}

type Config struct {
	OperatorRepo  OperatorRepo
	MainnetClient MainnetClient
	TxManager     TxManager
}

// epochUpdator implements EpochStateQuerier interface
type epochUpdator struct {
	operatorRepo      OperatorRepo
	mainnetClient     MainnetClient
	txManager         TxManager
	currentEpochState EpochState

	lastProcessedEpoch uint32

	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NewEpochUpdator creates a new epoch updator
func NewEpochUpdator(ctx context.Context, cfg *Config) (EpochStateQuerier, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is nil")
	}
	if cfg.OperatorRepo == nil {
		return nil, fmt.Errorf("operatorRepo is nil")
	}
	if cfg.MainnetClient == nil {
		return nil, fmt.Errorf("mainnetClient is nil")
	}
	if cfg.TxManager == nil {
		return nil, fmt.Errorf("txManager is nil")
	}
	e := &epochUpdator{
		operatorRepo:  cfg.OperatorRepo,
		mainnetClient: cfg.MainnetClient,
		txManager:     cfg.TxManager,
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
func (e *epochUpdator) start(ctx context.Context) error {
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
				currentEpoch := utils.CalculateCurrentEpochNumber(blockNumber)

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

func (e *epochUpdator) handleEpochUpdate(ctx context.Context, currentEpoch uint32) error {
	currentBlock, err := e.mainnetClient.BlockNumber(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current block number: %w", err)
	}

	_, err = e.txManager.Transact(ctx, pgx.TxOptions{}, func(ctx context.Context, tx *wpgx.WTx) (any, error) {
		txQuerier := e.operatorRepo.WithTx(tx)

		err = e.updateEpochState(ctx, currentEpoch)
		if err != nil {
			return nil, fmt.Errorf("failed to update epoch state: %w", err)
		}

		allOperators, err := txQuerier.ListAllOperators(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get all operators: %w", err)
		}

		log.Printf("[Epoch] Processing %d operators for epoch %d (block %d)", len(allOperators), currentEpoch, currentBlock)

		for _, operator := range allOperators {
			activeEpoch := operator.ActiveEpoch
			exitEpoch := operator.ExitEpoch

			isActive := IsOperatorActive(currentBlock, activeEpoch, exitEpoch)
			currentWeight, err := e.mainnetClient.GetOperatorWeight(ctx, ethcommon.HexToAddress(operator.Address))
			if err != nil {
				log.Printf("[Epoch] Failed to get weight for operator %s: %v", operator.Address, err)
				return nil, fmt.Errorf("failed to get weight for operator %s: %w", operator.Address, err)
			}

			currentSigningKey, err := e.mainnetClient.GetOperatorSigningKey(ctx, ethcommon.HexToAddress(operator.Address), currentEpoch)
			if err != nil {
				log.Printf("[Epoch] Failed to get signing key for operator %s: %v", operator.Address, err)
				return nil, fmt.Errorf("failed to get signing key for operator %s: %w", operator.Address, err)
			}

			currentP2PKey, err := e.mainnetClient.GetOperatorP2PKey(ctx, ethcommon.HexToAddress(operator.Address), currentEpoch)
			if err != nil {
				log.Printf("[Epoch] Failed to get P2P key for operator %s: %v", operator.Address, err)
				return nil, fmt.Errorf("failed to get P2P key for operator %s: %w", operator.Address, err)
			}

			weightNum := utils.BigIntToNumeric(currentWeight)
			params := operators.UpdateOperatorStateParams{
				Address:    operator.Address,
				IsActive:   isActive,
				Weight:     weightNum,
				SigningKey: currentSigningKey.String(),
				P2pKey:     currentP2PKey.String(),
			}
			_, err = txQuerier.UpdateOperatorState(ctx, params, &operator.Address, &operator.SigningKey, &operator.P2pKey, &operator.P2pKey)

			if err != nil {
				log.Printf("[Epoch] Failed to update operator %s state: %v", operator.Address, err)
				return nil, fmt.Errorf("failed to update operator %s state: %w", operator.Address, err)
			}
		}
		return nil, nil
	})

	if err != nil {
		return fmt.Errorf("transaction failed: %w", err)
	}

	return nil
}

func (e *epochUpdator) updateEpochState(ctx context.Context, currentEpoch uint32) error {
	minimumStake, err := e.mainnetClient.GetMinimumWeight(ctx)
	if err != nil {
		return fmt.Errorf("failed to get minimum stake: %w", err)
	}

	totalWeight, err := e.mainnetClient.GetTotalWeight(ctx)
	if err != nil {
		return fmt.Errorf("failed to get total stake: %w", err)
	}

	thresholdWeight, err := e.mainnetClient.GetThresholdWeight(ctx)
	if err != nil {
		return fmt.Errorf("failed to get threshold weight: %w", err)
	}

	e.currentEpochState = EpochState{
		EpochNumber:     currentEpoch,
		MinimumWeight:   minimumStake,
		TotalWeight:     totalWeight,
		ThresholdWeight: thresholdWeight,
	}
	log.Printf("[Epoch] Epoch state: %+v", e.currentEpochState)
	return nil
}

func (e *epochUpdator) Stop() {
	e.mainnetClient.Close()
	e.cancel()
	e.wg.Wait()
	log.Printf("[Epoch] Epoch monitoring stopped")
}
