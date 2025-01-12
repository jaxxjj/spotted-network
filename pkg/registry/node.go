package registry

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/galxe/spotted-network/pkg/p2p"
	"github.com/libp2p/go-libp2p/core/peer"
)

type Node struct {
	host *p2p.Host
	
	// Connected operators
	operators map[peer.ID]*OperatorInfo
	operatorsMu sync.RWMutex
	
	// Health check interval
	healthCheckInterval time.Duration
}

type OperatorInfo struct {
	ID peer.ID
	LastSeen time.Time
	Status string
}

func NewNode(ctx context.Context, cfg *p2p.Config) (*Node, error) {
	host, err := p2p.NewHost(ctx, cfg)
	if err != nil {
		return nil, err
	}

	node := &Node{
		host:               host,
		operators:          make(map[peer.ID]*OperatorInfo),
		healthCheckInterval: 5 * time.Second,
	}

	// Start health check
	go node.startHealthCheck(ctx)

	log.Printf("Registry Node started. %s\n", host.GetHostInfo())
	return node, nil
}

func (n *Node) startHealthCheck(ctx context.Context) {
	ticker := time.NewTicker(n.healthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			n.checkOperators(ctx)
		}
	}
}

func (n *Node) checkOperators(ctx context.Context) {
	n.operatorsMu.Lock()
	defer n.operatorsMu.Unlock()

	for id, info := range n.operators {
		// Ping the operator
		if err := n.host.PingPeer(ctx, id); err != nil {
			log.Printf("Operator %s is unreachable: %v\n", id, err)
			info.Status = "unreachable"
		} else {
			info.LastSeen = time.Now()
			info.Status = "active"
			log.Printf("Operator %s is healthy\n", id)
		}
	}
}

func (n *Node) AddOperator(id peer.ID) {
	n.operatorsMu.Lock()
	defer n.operatorsMu.Unlock()

	n.operators[id] = &OperatorInfo{
		ID: id,
		LastSeen: time.Now(),
		Status: "active",
	}
	log.Printf("New operator added: %s\n", id)
}

func (n *Node) RemoveOperator(id peer.ID) {
	n.operatorsMu.Lock()
	defer n.operatorsMu.Unlock()

	delete(n.operators, id)
	log.Printf("Operator removed: %s\n", id)
}

func (n *Node) GetOperatorCount() int {
	n.operatorsMu.RLock()
	defer n.operatorsMu.RUnlock()
	return len(n.operators)
}

func (n *Node) Stop() error {
	return n.host.Close()
} 