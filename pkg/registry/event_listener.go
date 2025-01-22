package registry

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"os"
	"strconv"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/galxe/spotted-network/pkg/common/contracts"
	eth "github.com/galxe/spotted-network/pkg/common/contracts/ethereum"
	"github.com/galxe/spotted-network/pkg/repos/registry/operators"
	"github.com/jackc/pgx/v5/pgtype"
)

// getStartBlock returns the block number to start listening from
func getStartBlock() uint64 {
	startBlockStr := os.Getenv("START_BLOCK")
	if startBlockStr == "" {
		log.Printf("[EventListener] START_BLOCK not set, using default value 0")
		return 0 // Default to 0 if not set
	}
	startBlock, err := strconv.ParseUint(startBlockStr, 10, 64)
	if err != nil {
		log.Printf("[EventListener] Invalid START_BLOCK value: %v, using 0", err)
		return 0
	}
	log.Printf("[EventListener] Using START_BLOCK: %d", startBlock)
	return startBlock
}

type EventListener struct {
	chainClients *eth.ChainClients     // Chain clients manager
	operatorQueries *operators.Queries     // Database queries
}

func NewEventListener(chainClients *eth.ChainClients, operatorQueries *operators.Queries) *EventListener {
	log.Printf("[EventListener] Creating new EventListener instance")
	return &EventListener{
		chainClients: chainClients,
		operatorQueries: operatorQueries,
	}
}

// StartListening starts listening for operator events
func (el *EventListener) StartListening(ctx context.Context) error {
	log.Printf("[EventListener] Starting event listener...")
	
	// Get mainnet client (which is actually anvil in our case)
	client := el.chainClients.GetMainnetClient()
	if client == nil {
		return fmt.Errorf("mainnet client is nil")
	}
	log.Printf("[EventListener] Got mainnet client successfully")

	startBlock := getStartBlock()

	// Create filter options
	filterOpts := &bind.FilterOpts{
		Start: startBlock,
		Context: ctx,
	}
	log.Printf("[EventListener] Created filter options: start=%d", filterOpts.Start)

	// Create event channels
	registeredEventChan := make(chan *contracts.OperatorRegisteredEvent)
	deregisteredEventChan := make(chan *contracts.OperatorDeregisteredEvent)
	log.Printf("[EventListener] Created event channels")

	// Subscribe to registration events
	log.Printf("[EventListener] Subscribing to OperatorRegistered events...")
	regSub, err := client.WatchOperatorRegistered(filterOpts, registeredEventChan)
	if err != nil {
		log.Printf("[EventListener] Failed to subscribe to registration events: %v", err)
		return fmt.Errorf("[EventListener] failed to subscribe to registration events: %w", err)
	}
	log.Printf("[EventListener] Successfully subscribed to OperatorRegistered events")

	// Subscribe to deregistration events
	log.Printf("[EventListener] Subscribing to OperatorDeregistered events...")
	deregSub, err := client.WatchOperatorDeregistered(filterOpts, deregisteredEventChan)
	if err != nil {
		regSub.Unsubscribe() // Clean up registration subscription
		log.Printf("[EventListener] Failed to subscribe to deregistration events: %v", err)
		return fmt.Errorf("[EventListener] failed to subscribe to deregistration events: %w", err)
	}
	log.Printf("[EventListener] Successfully subscribed to OperatorDeregistered events")

	// Start a goroutine to handle events
	go func() {
		defer regSub.Unsubscribe()
		defer deregSub.Unsubscribe()
		log.Printf("[EventListener] Started event handling goroutine")

		for {
			select {
			case <-ctx.Done():
				log.Printf("[EventListener] Context done, stopping event listener")
				return
			case err := <-regSub.Err():
				log.Printf("[EventListener] Registration event subscription error: %v", err)
				return
			case err := <-deregSub.Err():
				log.Printf("[EventListener] Deregistration event subscription error: %v", err)
				return
			case event := <-registeredEventChan:
				log.Printf("[EventListener] Received OperatorRegistered event: operator=%s, signingKey=%s, blockNumber=%s",
					event.Operator.Hex(), event.SigningKey.Hex(), event.BlockNumber.String())
				if err := el.handleOperatorRegistered(ctx, event); err != nil {
					log.Printf("[EventListener] Error handling registration event: %v", err)
				}
			case event := <-deregisteredEventChan:
				log.Printf("[EventListener] Received OperatorDeregistered event: operator=%s, blockNumber=%s",
					event.Operator.Hex(), event.BlockNumber.String())
				if err := el.handleOperatorDeregistered(ctx, event); err != nil {
					log.Printf("[EventListener] Error handling deregistration event: %v", err)
				}
			}
		}
	}()

	log.Printf("[EventListener] Started listening for operator events from block %d\n", startBlock)
	return nil
}

// handleOperatorRegistered processes a new OperatorRegistered event
func (el *EventListener) handleOperatorRegistered(ctx context.Context, event *contracts.OperatorRegisteredEvent) error {
	log.Printf("[EventListener] Processing OperatorRegistered event...")
	
	// Get mainnet client (which is actually anvil in our case)
	client := el.chainClients.GetMainnetClient()
	log.Printf("[EventListener] Got mainnet client for event processing")
	
	// Get the active epoch for the registration block
	activeEpoch, err := client.GetEffectiveEpochForBlock(ctx, event.BlockNumber.Uint64())
	if err != nil {
		log.Printf("[EventListener] Failed to get active epoch: %v", err)
		return fmt.Errorf("[EventListener] failed to get active epoch: %w", err)
	}
	log.Printf("[EventListener] Got active epoch: %d", activeEpoch)

	blockNumber := event.BlockNumber.Uint64()
	timestamp := event.Timestamp.Uint64()
	log.Printf("[EventListener] Event values - Block Number: %d, Timestamp: %d", 
		blockNumber, timestamp)

	// Create or update operator record
	params := operators.UpsertOperatorParams{
		Address:     event.Operator.Hex(),
		SigningKey:  event.SigningKey.Hex(),
		RegisteredAtBlockNumber: pgtype.Numeric{
			Int:    new(big.Int).SetUint64(blockNumber),
			Exp:    0,
			Valid:  true,
		},
		RegisteredAtTimestamp: pgtype.Numeric{
			Int:    new(big.Int).SetUint64(timestamp),
			Exp:    0,
			Valid:  true,
		},
		ActiveEpoch: pgtype.Numeric{
			Int:    new(big.Int).SetInt64(int64(activeEpoch)),
			Exp:    0,
			Valid:  true,
		},
		Weight: pgtype.Numeric{
			Int:    new(big.Int).SetInt64(0),
			Exp:    0,
			Valid:  true,
		},
		ExitEpoch: pgtype.Numeric{
			Int:    new(big.Int).SetUint64(4294967295), // max uint32 value to indicate no exit epoch
			Exp:    0,
			Valid:  true,
		},
	}
	log.Printf("Created operator params: %+v", params)

	// Upsert operator in database
	_, err = el.operatorQueries.UpsertOperator(ctx, params)
	if err != nil {
		log.Printf("[EventListener] Failed to upsert operator: %v", err)
		return fmt.Errorf("[EventListener] failed to upsert operator: %w", err)
	}

	log.Printf("Successfully upserted operator %s in database", event.Operator.Hex())

	// Update operator status after creation/update
	if err := el.updateStatusAfterOperations(ctx, event.Operator.String()); err != nil {
		log.Printf("[EventListener] Failed to update operator status: %v", err)
		return fmt.Errorf("[EventListener] failed to update operator status: %w", err)
	}

	return nil
}

// updateStatusAfterOperations updates operator status based on current block number and epochs
func (el *EventListener) updateStatusAfterOperations(ctx context.Context, operatorAddr string) error {
	// Get mainnet client
	client := el.chainClients.GetMainnetClient()

	// Get current block number
	currentBlock, err := client.GetLatestBlockNumber(ctx)
	if err != nil {
		return fmt.Errorf("[EventListener] failed to get current block number: %w", err)
	}

	// Get operator from database to get active_epoch and exit_epoch
	operator, err := el.operatorQueries.GetOperatorByAddress(ctx, operatorAddr)
	if err != nil {
		return fmt.Errorf("[EventListener] failed to get operator from database: %w", err)
	}

	activeEpoch := uint32(operator.ActiveEpoch.Int.Int64())
	var exitEpoch uint32
	if operator.ExitEpoch.Valid {
		exitEpoch = uint32(operator.ExitEpoch.Int.Int64())
	} else {
		exitEpoch = 4294967295 // max uint32 value to indicate no exit epoch
	}

	// Determine operator status
	var status string
	var weight *big.Int

	if exitEpoch == 4294967295 {
		// Case 1: Only has active epoch, no exit epoch
		if currentBlock >= uint64(activeEpoch) {
			status = "active"
			log.Printf("[EventListener] Operator %s marked as active", operatorAddr)
		} else {
			status = "waitingActive"
			log.Printf("[EventListener] Operator %s marked as waitingActive", operatorAddr)
		}
	} else if exitEpoch > activeEpoch {
		// Case 2: Has exit epoch and exit epoch > active epoch
		if currentBlock < uint64(activeEpoch) {
			status = "waitingActive"
			log.Printf("[EventListener] Operator %s marked as waitingActive", operatorAddr)
		} else if currentBlock >= uint64(activeEpoch) && currentBlock < uint64(exitEpoch) {
			status = "active"
			log.Printf("[EventListener] Operator %s marked as active", operatorAddr)
		} else {
			status = "inactive"
			log.Printf("[EventListener] Operator %s marked as inactive", operatorAddr)
		}
	} else {
		// Case 3: Has exit epoch and exit epoch <= active epoch
		if currentBlock < uint64(activeEpoch) {
			status = "waitingActive"
			log.Printf("[EventListener] Operator %s marked as waitingActive", operatorAddr)
		} else {
			status = "active"
			log.Printf("[EventListener] Operator %s marked as active", operatorAddr)
		}
	}

	// Get weight if status is active
	if status == "active" {
		weight, err = client.GetOperatorWeight(ctx, common.HexToAddress(operatorAddr))
		if err != nil {
			return fmt.Errorf("[EventListener] failed to get operator weight: %w", err)
		}
	} else {
		weight = big.NewInt(0)
	}

	// Update operator state in database
	_, err = el.operatorQueries.UpdateOperatorState(ctx, operators.UpdateOperatorStateParams{
		Address: operatorAddr,
		Status:  status,
		Weight: pgtype.Numeric{
			Int:    weight,
			Valid:  true,
			Exp:    0,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to update operator state: %w", err)
	}

	log.Printf("[EventListener] Updated operator %s status to %s with weight %s", operatorAddr, status, weight.String())
	return nil
}

// handleOperatorDeregistered handles operator deregistration events
func (el *EventListener) handleOperatorDeregistered(ctx context.Context, event *contracts.OperatorDeregisteredEvent) error {
	log.Printf("[EventListener] Processing OperatorDeregistered event...")
	
	// Get mainnet client (which is actually anvil in our case)
	client := el.chainClients.GetMainnetClient()
	log.Printf("[EventListener] Got mainnet client for event processing")
	
	// Get the exit epoch for the deregistration block
	exitEpoch, err := client.GetEffectiveEpochForBlock(ctx, event.BlockNumber.Uint64())
	if err != nil {
		log.Printf("[EventListener] Failed to get exit epoch: %v", err)
		return fmt.Errorf("[EventListener] failed to get exit epoch: %w", err)
	}
	log.Printf("[EventListener] Got exit epoch: %d", exitEpoch)

	// Update operator status in database
	_, err = el.operatorQueries.UpdateOperatorExitEpoch(ctx, operators.UpdateOperatorExitEpochParams{
		Address:   event.Operator.Hex(),
		ExitEpoch: pgtype.Numeric{
			Int:    new(big.Int).SetInt64(int64(exitEpoch)),
			Exp:    0,
			Valid:  true,
		},
		Status:    "waitingExit",
	})
	if err != nil {
		log.Printf("[EventListener] Failed to update operator exit epoch: %v", err)
		return fmt.Errorf("[EventListener] failed to update operator exit epoch: %w", err)
	}
	log.Printf("[EventListener] Updated operator %s exit epoch to %d", event.Operator.Hex(), exitEpoch)

	// Update operator status after setting exit epoch
	if err := el.updateStatusAfterOperations(ctx, event.Operator.String()); err != nil {
		log.Printf("[EventListener] Failed to update operator status: %v", err)
		return fmt.Errorf("[EventListener] failed to update operator status: %w", err)
	}

	return nil
}