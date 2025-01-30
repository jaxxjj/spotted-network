package registry

import (
	"context"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/galxe/spotted-network/pkg/common/contracts/ethereum"
	"github.com/galxe/spotted-network/pkg/common/types"
	"github.com/galxe/spotted-network/pkg/repos/registry/operators"
	"github.com/jackc/pgx/v5/pgtype"
)

type EventListener struct {
	node *Node
	mainnetClient MainnetClient
	operators OperatorsQuerier
}

func NewEventListener(node *Node, mainnetClient MainnetClient, operators OperatorsQuerier) *EventListener {
	log.Printf("[EventListener] Creating new EventListener instance")
	return &EventListener{
		node: node,
		mainnetClient: mainnetClient,
		operators: operators,
	}
}

// StartListening starts listening for operator events
func (el *EventListener) StartListening(ctx context.Context) error {
	log.Printf("[EventListener] Starting event listener...")

	// Create filter options
	filterOpts := &bind.FilterOpts{
		Start: GenesisBlock,
		Context: ctx,
	}
	log.Printf("[EventListener] Created filter options: start=%d", filterOpts.Start)

	// Create event channels
	registeredEventChan := make(chan *ethereum.OperatorRegisteredEvent)
	deregisteredEventChan := make(chan *ethereum.OperatorDeregisteredEvent)
	log.Printf("[EventListener] Created event channels")

	// Subscribe to registration events
	log.Printf("[EventListener] Subscribing to OperatorRegistered events...")
	regSub, err := el.mainnetClient.WatchOperatorRegistered(filterOpts, registeredEventChan)
	if err != nil {
		log.Printf("[EventListener] Failed to subscribe to registration events: %v", err)
		return fmt.Errorf("[EventListener] failed to subscribe to registration events: %w", err)
	}
	log.Printf("[EventListener] Successfully subscribed to OperatorRegistered events")

	// Subscribe to deregistration events
	log.Printf("[EventListener] Subscribing to OperatorDeregistered events...")
	deregSub, err := el.mainnetClient.WatchOperatorDeregistered(filterOpts, deregisteredEventChan)
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

	log.Printf("[EventListener] Started listening for operator events from block %d\n", GenesisBlock)
	return nil
}

// handleOperatorRegistered processes a new OperatorRegistered event
func (el *EventListener) handleOperatorRegistered(ctx context.Context, event *ethereum.OperatorRegisteredEvent) error {
	log.Printf("[EventListener] Processing OperatorRegistered event...")
	blockNumber := event.BlockNumber.Uint64()
	timestamp := event.Timestamp.Uint64()
	log.Printf("[EventListener] Event values - Block Number: %d, Timestamp: %d", 
		blockNumber, timestamp)

	// Get the active epoch for the registration block
	activeEpoch, err := el.mainnetClient.GetEffectiveEpochForBlock(ctx, blockNumber)
	if err != nil {
		log.Printf("[EventListener] Failed to get active epoch: %v", err)
		return fmt.Errorf("[EventListener] failed to get active epoch: %w", err)
	}
	log.Printf("[EventListener] Got active epoch: %d", activeEpoch)

	// Create or update operator record
	params := operators.UpsertOperatorParams{
		Address:     event.Operator.Hex(),
		SigningKey:  event.SigningKey.Hex(),
		RegisteredAtBlockNumber: blockNumber,
		RegisteredAtTimestamp: timestamp,
		ActiveEpoch: activeEpoch,
		Weight: pgtype.Numeric{
			Int:    big.NewInt(0),
			Exp:    0,
			Valid:  true,
		},
		ExitEpoch: 4294967295,
	}
	log.Printf("Created operator params: %+v", params)

	// Update operator in database
	_, err = el.operators.UpsertOperator(ctx, params, &params.Address)
	if err != nil {
		return fmt.Errorf("failed to upsert operator: %w", err)
	}

	log.Printf("Successfully upserted operator %s in database", event.Operator.Hex())

	// Update operator status after creation/update
	if err := el.node.updateStatusAfterOperations(ctx, event.Operator.String()); err != nil {
		log.Printf("[EventListener] Failed to update operator status: %v", err)
		return fmt.Errorf("[EventListener] failed to update operator status: %w", err)
	}

	return nil
}

// updateStatusAfterOperations updates operator status based on current block number and epochs
func (n *Node) updateStatusAfterOperations(ctx context.Context, operatorAddr string) error {

	// Get current block number
	currentBlock, err := n.mainnetClient.BlockNumber(ctx)
	if err != nil {
		return fmt.Errorf("[EventListener] failed to get current block number: %w", err)
	}

	// Get operator from database to get active_epoch and exit_epoch
	operator, err := n.operators.GetOperatorByAddress(ctx, operatorAddr)
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
	_, err = n.operators.UpdateOperatorState(ctx, operators.UpdateOperatorStateParams{
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

// handleOperatorDeregistered handles operator deregistration events
func (el *EventListener) handleOperatorDeregistered(ctx context.Context, event *ethereum.OperatorDeregisteredEvent) error {
	log.Printf("[EventListener] Processing OperatorDeregistered event...")
	
	// Get the exit epoch for the deregistration block
	exitEpoch, err := el.mainnetClient.GetEffectiveEpochForBlock(ctx, event.BlockNumber.Uint64())
	if err != nil {
		log.Printf("[EventListener] Failed to get exit epoch: %v", err)
		return fmt.Errorf("[EventListener] failed to get exit epoch: %w", err)
	}
	log.Printf("[EventListener] Got exit epoch: %d", exitEpoch)

	// Update operator exit epoch
	operatorAddr := event.Operator.Hex()
	_, err = el.operators.UpdateOperatorExitEpoch(ctx, operators.UpdateOperatorExitEpochParams{
		Address:   operatorAddr,
		ExitEpoch: exitEpoch,
	}, &operatorAddr)
	if err != nil {
		return fmt.Errorf("failed to update operator exit epoch: %w", err)
	}
	log.Printf("[EventListener] Updated operator %s exit epoch to %d", operatorAddr, exitEpoch)

	// Update operator status after setting exit epoch
	if err := el.node.updateStatusAfterOperations(ctx, event.Operator.String()); err != nil {
		log.Printf("[EventListener] Failed to update operator status: %v", err)
		return fmt.Errorf("[EventListener] failed to update operator status: %w", err)
	}

	return nil
}