package registry

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/event"
	"github.com/galxe/spotted-network/pkg/common/contracts/ethereum"
	"github.com/galxe/spotted-network/pkg/repos/registry/operators"
	"github.com/jackc/pgx/v5/pgtype"
)



type EventListenerChainClient interface {
	GetEffectiveEpochForBlock(ctx context.Context, blockNumber uint64) (uint32, error)
	GetOperatorWeight(ctx context.Context, address ethcommon.Address) (*big.Int, error)
	WatchOperatorRegistered(filterOpts *bind.FilterOpts, sink chan<- *ethereum.OperatorRegisteredEvent) (event.Subscription, error)
	WatchOperatorDeregistered(filterOpts *bind.FilterOpts, sink chan<- *ethereum.OperatorDeregisteredEvent) (event.Subscription, error) 
}

type EventListenerConfig struct {
	node *Node
	mainnetClient EventListenerChainClient
	operators OperatorsQuerier
}

type EventListener struct {
	node *Node
	mainnetClient EventListenerChainClient
	operators OperatorsQuerier
	
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func NewEventListener(ctx context.Context, cfg *EventListenerConfig) *EventListener {
	if cfg.node == nil {
		log.Fatal("[EventListener] node not initialized")
	}
	if cfg.mainnetClient == nil {
		log.Fatal("[EventListener] mainnet client not initialized")
	}
	if cfg.operators == nil {
		log.Fatal("[EventListener] operators querier not initialized")
	}

	// 创建带 cancel 的 context
	ctx, cancel := context.WithCancel(ctx)
	
	el := &EventListener{
		node:          cfg.node,
		mainnetClient: cfg.mainnetClient,
		operators:     cfg.operators,
		cancel:        cancel,
	}
	
	// 启动监听
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
	return el
}

// StartListening starts listening for operator events
func (el *EventListener) start(ctx context.Context) error {
	log.Printf("[EventListener] Starting event listener...")

	// 创建 filter options
	filterOpts := &bind.FilterOpts{
		Start:   GenesisBlock,
		Context: ctx,
	}

	// 创建事件通道
	registeredChan := make(chan *ethereum.OperatorRegisteredEvent)
	deregisteredChan := make(chan *ethereum.OperatorDeregisteredEvent)

	// 订阅事件
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

	// 事件处理循环
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

// 简单的事件订阅函数，包含基本重试
func (el *EventListener) subscribeToEvents(
	ctx context.Context,
	filterOpts *bind.FilterOpts,
	registeredChan chan<- *ethereum.OperatorRegisteredEvent,
	deregisteredChan chan<- *ethereum.OperatorDeregisteredEvent,
) (event.Subscription, event.Subscription, error) {
	var regSub, deregSub event.Subscription
	var err error

	// 简单重试 3 次
	for i := 0; i < 3; i++ {
		regSub, err = el.mainnetClient.WatchOperatorRegistered(filterOpts, registeredChan)
		if err == nil {
			deregSub, err = el.mainnetClient.WatchOperatorDeregistered(filterOpts, deregisteredChan)
			if err == nil {
				return regSub, deregSub, nil
			}
			regSub.Unsubscribe()
		}

		log.Printf("[EventListener] Subscription attempt %d failed: %v", i+1, err)
		
		// 简单延迟后重试
		select {
		case <-ctx.Done():
			return nil, nil, ctx.Err()
		case <-time.After(time.Second):
		}
	}

	return nil, nil, fmt.Errorf("failed to subscribe after retries: %w", err)
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
	weight, err := el.mainnetClient.GetOperatorWeight(ctx, event.Operator)
	if err != nil {
		log.Printf("[EventListener] Failed to get operator weight: %v", err)
		return fmt.Errorf("[EventListener] failed to get operator weight: %w", err)
	}
	// Create or update operator record
	params := operators.UpsertOperatorParams{
		Address:     event.Operator.Hex(),
		SigningKey:  event.SigningKey.Hex(),
		RegisteredAtBlockNumber: blockNumber,
		RegisteredAtTimestamp: timestamp,
		ActiveEpoch: activeEpoch,
		Weight: pgtype.Numeric{
			Int:    weight,
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
	if err := el.node.updateSingleOperatorState(ctx, event.Operator.Hex()); err != nil {
		log.Printf("[EventListener] Failed to update operator status: %v", err)
		return fmt.Errorf("[EventListener] failed to update operator status: %w", err)
	}

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
	if err := el.node.updateSingleOperatorState(ctx, event.Operator.Hex()); err != nil {
		log.Printf("[EventListener] Failed to update operator status: %v", err)
		return fmt.Errorf("[EventListener] failed to update operator status: %w", err)
	}

	return nil
}

func (el *EventListener) Stop() {
	log.Printf("[EventListener] Stopping event listener...")
	
	// Cancel context
	el.cancel()
	
	// Wait for goroutines
	el.wg.Wait()
	
	log.Printf("[EventListener] Event listener stopped")
}