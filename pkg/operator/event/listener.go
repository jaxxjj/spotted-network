package event

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/event"
	utils "github.com/galxe/spotted-network/pkg/common"
	"github.com/galxe/spotted-network/pkg/common/contracts/ethereum"
	"github.com/galxe/spotted-network/pkg/operator/constants"
	"github.com/galxe/spotted-network/pkg/repos/operators"
	"github.com/jackc/pgx/v5/pgtype"
)

type ChainClient interface {
	WatchOperatorRegistered(filterOpts *bind.FilterOpts, sink chan<- *ethereum.OperatorRegisteredEvent) (event.Subscription, error)
	WatchOperatorDeregistered(filterOpts *bind.FilterOpts, sink chan<- *ethereum.OperatorDeregisteredEvent) (event.Subscription, error)
	Close() error
}

type OperatorRepo interface {
	UpdateOperatorExitEpoch(ctx context.Context, arg operators.UpdateOperatorExitEpochParams, getOperatorByAddress *string, getOperatorBySigningKey *string, getOperatorByP2PKey *string, isActiveOperator *string) (*operators.Operators, error)
	GetOperatorByAddress(ctx context.Context, address string) (*operators.Operators, error)
	UpsertOperator(ctx context.Context, arg operators.UpsertOperatorParams, getOperatorByAddress *string, getOperatorBySigningKey *string, getOperatorByP2PKey *string, isActiveOperator *string) (*operators.Operators, error)
}

type Config struct {
	MainnetClient ChainClient
	OperatorRepo  OperatorRepo
}

type EventListenerService interface {
	Stop()
}

type eventListener struct {
	mainnetClient ChainClient
	operatorRepo  OperatorRepo

	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func NewEventListener(ctx context.Context, cfg *Config) (EventListenerService, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is nil")
	}
	if cfg.MainnetClient == nil {
		return nil, fmt.Errorf("mainnet client not initialized")
	}
	if cfg.OperatorRepo == nil {
		return nil, fmt.Errorf("operator repo not initialized")
	}

	ctx, cancel := context.WithCancel(ctx)

	el := &eventListener{
		mainnetClient: cfg.MainnetClient,
		operatorRepo:  cfg.OperatorRepo,
		cancel:        cancel,
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[EventListener] Recovered from panic: %v", r)
			}
		}()

		if err := el.start(ctx); err != nil {
			log.Printf("[EventListener] Event listener stopped with error: %v", err)
		}
	}()

	log.Printf("[EventListener] Event listener started")
	return el, nil
}

// starts listening for operator events
func (el *eventListener) start(ctx context.Context) error {
	log.Printf("[EventListener] Starting event listener...")

	filterOpts := &bind.FilterOpts{
		Start:   constants.GenesisBlock,
		Context: ctx,
	}

	registeredChan := make(chan *ethereum.OperatorRegisteredEvent)
	deregisteredChan := make(chan *ethereum.OperatorDeregisteredEvent)

	regSub, deregSub, err := el.subscribeToEvents(ctx, filterOpts, registeredChan, deregisteredChan)
	if err != nil {
		return fmt.Errorf("failed to subscribe to events: %w", err)
	}
	defer func() {
		regSub.Unsubscribe()
		deregSub.Unsubscribe()
		log.Printf("[EventListener] Unsubscribed from all events")
	}()

	log.Printf("[EventListener] Successfully subscribed to all events")

	for {
		select {
		case <-ctx.Done():
			log.Printf("[EventListener] Stopping event listener: %v", ctx.Err())
			return nil

		case err := <-regSub.Err():
			log.Printf("[EventListener] Registration subscription error: %v", err)
			return fmt.Errorf("registration subscription error: %w", err)

		case err := <-deregSub.Err():
			log.Printf("[EventListener] Deregistration subscription error: %v", err)
			return fmt.Errorf("deregistration subscription error: %w", err)

		case event := <-registeredChan:
			log.Printf("[EventListener] Processing registration for operator %s", event.Operator.Hex())
			if err := el.handleOperatorRegistered(ctx, event); err != nil {
				log.Printf("[EventListener] Failed to handle registration: %v", err)
			}

		case event := <-deregisteredChan:
			log.Printf("[EventListener] Processing deregistration for operator %s", event.Operator.Hex())
			if err := el.handleOperatorDeregistered(ctx, event); err != nil {
				log.Printf("[EventListener] Failed to handle deregistration: %v", err)
			}
		}
	}
}

func (el *eventListener) subscribeToEvents(
	ctx context.Context,
	filterOpts *bind.FilterOpts,
	registeredChan chan<- *ethereum.OperatorRegisteredEvent,
	deregisteredChan chan<- *ethereum.OperatorDeregisteredEvent,
) (event.Subscription, event.Subscription, error) {
	var regSub, deregSub event.Subscription
	var err error

	for i := 0; i < 3; i++ {
		regSub, err = el.mainnetClient.WatchOperatorRegistered(filterOpts, registeredChan)
		if err == nil {
			deregSub, err = el.mainnetClient.WatchOperatorDeregistered(filterOpts, deregisteredChan)
			if err == nil {
				return regSub, deregSub, nil
			}
			regSub.Unsubscribe()
		}

		select {
		case <-ctx.Done():
			return nil, nil, ctx.Err()
		case <-time.After(time.Second):
		}
	}

	return nil, nil, fmt.Errorf("failed to subscribe after retries: %w", err)
}

// handleOperatorRegistered processes a new OperatorRegistered event
func (el *eventListener) handleOperatorRegistered(ctx context.Context, event *ethereum.OperatorRegisteredEvent) error {
	log.Printf("[EventListener] Processing OperatorRegistered event...")
	blockNumber := event.BlockNumber.Uint64()

	// Get the active epoch for the registration block
	activeEpoch := utils.GetEffectiveEpochByBlockNumber(blockNumber)
	log.Printf("[EventListener] Got active epoch: %d", activeEpoch)

	// Create or update operator record
	params := operators.UpsertOperatorParams{
		Address:                 event.Operator.Hex(),
		SigningKey:              event.SigningKey.Hex(),
		P2pKey:                  event.P2PKey.Hex(),
		RegisteredAtBlockNumber: blockNumber,
		ActiveEpoch:             activeEpoch,
		Weight: pgtype.Numeric{
			Int:   big.NewInt(0),
			Exp:   0,
			Valid: true,
		},
		ExitEpoch: 4294967295,
	}

	// Update operator in database
	_, err := el.operatorRepo.UpsertOperator(ctx, params, &params.Address, &params.SigningKey, &params.P2pKey, &params.P2pKey)
	if err != nil {
		return fmt.Errorf("failed to upsert operator: %w", err)
	}

	log.Printf("Successfully upserted operator %s in database", event.Operator.Hex())

	return nil
}

// handleOperatorDeregistered handles operator deregistration events
func (el *eventListener) handleOperatorDeregistered(ctx context.Context, event *ethereum.OperatorDeregisteredEvent) error {
	log.Printf("[EventListener] Processing OperatorDeregistered event...")
	blockNumber := event.BlockNumber.Uint64()

	exitEpoch := utils.GetEffectiveEpochByBlockNumber(blockNumber)
	log.Printf("[EventListener] Got exit epoch: %d", exitEpoch)

	operatorAddress := event.Operator.Hex()
	operator, err := el.operatorRepo.GetOperatorByAddress(ctx, operatorAddress)
	if err != nil {
		return fmt.Errorf("failed to get operator: %w", err)
	}
	operatorSigningKey := operator.SigningKey
	operatorP2PKey := operator.P2pKey

	_, err = el.operatorRepo.UpdateOperatorExitEpoch(ctx, operators.UpdateOperatorExitEpochParams{
		Address:   operatorAddress,
		ExitEpoch: exitEpoch,
	}, &operatorAddress, &operatorSigningKey, &operatorP2PKey, &operatorP2PKey)
	if err != nil {
		return fmt.Errorf("failed to update operator exit epoch: %w", err)
	}
	log.Printf("[EventListener] Updated operator %s exit epoch to %d", operatorAddress, exitEpoch)

	return nil
}

func (el *eventListener) Stop() {
	log.Printf("[EventListener] Stopping event listener...")
	el.mainnetClient.Close()
	// Cancel context
	el.cancel()

	// Wait for goroutines
	el.wg.Wait()

	log.Printf("[EventListener] Event listener stopped")
}
