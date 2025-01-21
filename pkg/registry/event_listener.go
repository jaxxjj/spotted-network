package registry

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"os"
	"strconv"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/galxe/spotted-network/pkg/common/contracts"
	eth "github.com/galxe/spotted-network/pkg/common/contracts/ethereum"
	"github.com/galxe/spotted-network/pkg/repos/registry/operators"
	"github.com/jackc/pgx/v5/pgtype"
)

// getStartBlock returns the block number to start listening from
func getStartBlock() uint64 {
	startBlockStr := os.Getenv("START_BLOCK")
	if startBlockStr == "" {
		log.Printf("START_BLOCK not set, using default value 0")
		return 0 // Default to 0 if not set
	}
	startBlock, err := strconv.ParseUint(startBlockStr, 10, 64)
	if err != nil {
		log.Printf("Invalid START_BLOCK value: %v, using 0", err)
		return 0
	}
	log.Printf("Using START_BLOCK: %d", startBlock)
	return startBlock
}

type EventListener struct {
	chainClients *eth.ChainClients     // Chain clients manager
	db          *operators.Queries     // Database queries
}

func NewEventListener(chainClients *eth.ChainClients, db *operators.Queries) *EventListener {
	log.Printf("Creating new EventListener instance")
	return &EventListener{
		chainClients: chainClients,
		db:          db,
	}
}

// StartListening starts listening for OperatorRegistered events
func (el *EventListener) StartListening(ctx context.Context) error {
	log.Printf("Starting event listener...")
	
	// Get mainnet client (which is actually anvil in our case)
	client := el.chainClients.GetMainnetClient()
	if client == nil {
		return fmt.Errorf("mainnet client is nil")
	}
	log.Printf("Got mainnet client successfully")

	startBlock := getStartBlock()

	// Create filter options
	filterOpts := &bind.FilterOpts{
		Start: startBlock,
		Context: ctx,
	}
	log.Printf("Created filter options: start=%d", filterOpts.Start)

	// Create event channel
	eventChan := make(chan *contracts.OperatorRegisteredEvent)
	log.Printf("Created event channel")

	// Subscribe to events
	log.Printf("Subscribing to OperatorRegistered events...")
	sub, err := client.WatchOperatorRegistered(filterOpts, eventChan)
	if err != nil {
		log.Printf("Failed to subscribe to events: %v", err)
		return fmt.Errorf("failed to subscribe to events: %w", err)
	}
	log.Printf("Successfully subscribed to OperatorRegistered events")

	// Start a goroutine to handle events
	go func() {
		defer sub.Unsubscribe()
		log.Printf("Started event handling goroutine")

		for {
			select {
			case <-ctx.Done():
				log.Printf("Context done, stopping event listener")
				return
			case err := <-sub.Err():
				log.Printf("Event subscription error: %v", err)
				return
			case event := <-eventChan:
				log.Printf("Received OperatorRegistered event: operator=%s, signingKey=%s, blockNumber=%s",
					event.Operator.Hex(), event.SigningKey.Hex(), event.BlockNumber.String())
				if err := el.handleOperatorRegistered(ctx, event); err != nil {
					log.Printf("Error handling event: %v", err)
				}
			}
		}
	}()

	log.Printf("Started listening for OperatorRegistered events from block %d\n", startBlock)
	return nil
}

// handleOperatorRegistered processes a new OperatorRegistered event
func (el *EventListener) handleOperatorRegistered(ctx context.Context, event *contracts.OperatorRegisteredEvent) error {
	log.Printf("Processing OperatorRegistered event...")
	
	// Get mainnet client (which is actually anvil in our case)
	client := el.chainClients.GetMainnetClient()
	log.Printf("Got mainnet client for event processing")
	
	// Get the active epoch for the registration block
	activeEpoch, err := client.GetEffectiveEpochForBlock(ctx, event.BlockNumber.Uint64())
	if err != nil {
		log.Printf("Failed to get active epoch: %v", err)
		return fmt.Errorf("failed to get active epoch: %w", err)
	}
	log.Printf("Got active epoch: %d", activeEpoch)

	// Create operator record with waitingJoin status (status is set in SQL)
	params := operators.CreateOperatorParams{
		Address:                 event.Operator.Hex(),
		SigningKey:             event.SigningKey.Hex(),
		RegisteredAtBlockNumber: event.BlockNumber.Uint64(),
		RegisteredAtTimestamp:   event.Timestamp.Uint64(),
		ActiveEpoch:            uint32(activeEpoch),
		Weight: pgtype.Numeric{
			Int:    new(big.Int),  // Initialize with 0
			Valid:  true,          // This is a valid (non-NULL) value
			Exp:    0,             // No decimal places
		},
	}
	log.Printf("Created operator params: %+v", params)

	// Create operator in database
	_, err = el.db.CreateOperator(ctx, params)
	if err != nil {
		log.Printf("Failed to create operator: %v", err)
		return fmt.Errorf("failed to create operator: %w", err)
	}

	log.Printf("Successfully created operator %s in database", event.Operator.Hex())
	return nil
} 