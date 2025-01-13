package registry

import (
	"context"
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/galxe/spotted-network/pkg/common/contracts"
	"github.com/galxe/spotted-network/pkg/common/contracts/ethereum"
	"github.com/galxe/spotted-network/pkg/repos/registry"
	"github.com/jackc/pgx/v5/pgtype"
)

const (
	// StartBlockNumber is the block number to start listening from
	StartBlockNumber uint64 = 21618385
)

type EventListener struct {
    chainClients *ethereum.ChainClients  // Chain clients manager
    db          *registry.Queries        // Database queries
}

func NewEventListener(chainClients *ethereum.ChainClients, db *registry.Queries) *EventListener {
    return &EventListener{
        chainClients: chainClients,
        db:          db,
    }
}

// StartListening starts listening for OperatorRegistered events
func (el *EventListener) StartListening(ctx context.Context) error {
    // Get mainnet client
    mainnetClient := el.chainClients.GetMainnetClient()
    
    // Create a filter for OperatorRegistered events
    filterOpts := &bind.FilterOpts{
        Start:   StartBlockNumber, // Start from specified block
        Context: ctx,
    }

    // Watch for new events
    eventCh := make(chan *contracts.OperatorRegisteredEvent)
    sub, err := mainnetClient.WatchOperatorRegistered(filterOpts, eventCh)
    if err != nil {
        return fmt.Errorf("failed to subscribe to OperatorRegistered events: %w", err)
    }

    go func() {
        defer sub.Unsubscribe()

        for {
            select {
            case err := <-sub.Err():
                log.Printf("Error in event subscription: %v", err)
                return
            case event := <-eventCh:
                if err := el.handleOperatorRegistered(ctx, event); err != nil {
                    log.Printf("Error handling OperatorRegistered event: %v", err)
                }
            case <-ctx.Done():
                return
            }
        }
    }()

    log.Printf("Started listening for OperatorRegistered events from block %d\n", StartBlockNumber)
    return nil
}

// handleOperatorRegistered processes a new OperatorRegistered event
func (el *EventListener) handleOperatorRegistered(ctx context.Context, event *contracts.OperatorRegisteredEvent) error {
    // Get mainnet client
    mainnetClient := el.chainClients.GetMainnetClient()
    
    // Get the active epoch for the registration block
    activeEpoch, err := mainnetClient.GetEffectiveEpochForBlock(ctx, event.BlockNumber.Uint64())
    if err != nil {
        return fmt.Errorf("failed to get active epoch: %w", err)
    }

    // Create numeric values for database
    blockNumber := pgtype.Numeric{}
    if err := blockNumber.Scan(event.BlockNumber.String()); err != nil {
        return fmt.Errorf("failed to convert block number: %w", err)
    }

    timestamp := pgtype.Numeric{}
    if err := timestamp.Scan(event.Timestamp.String()); err != nil {
        return fmt.Errorf("failed to convert timestamp: %w", err)
    }

    weight := pgtype.Numeric{}
    if err := weight.Scan("0"); err != nil {
        return fmt.Errorf("failed to convert weight: %w", err)
    }

    // Create operator record with waitingJoin status (status is set in SQL)
    params := registry.CreateOperatorParams{
        Address:                 event.Operator.Hex(),
        SigningKey:             event.SigningKey.Hex(),
        RegisteredAtBlockNumber: blockNumber,
        RegisteredAtTimestamp:   timestamp,
        ActiveEpoch:            int32(activeEpoch),
        Weight:                 weight,
    }

    // Create operator in database
    _, err = el.db.CreateOperator(ctx, params)
    if err != nil {
        return fmt.Errorf("failed to create operator: %w", err)
    }

    log.Printf("Operator %s registered and marked as waitingJoin", event.Operator.Hex())
    return nil
} 