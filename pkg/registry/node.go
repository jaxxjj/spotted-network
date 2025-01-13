package registry

import (
	"bufio"
	"context"
	"io"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/galxe/spotted-network/pkg/p2p"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
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
	Addrs []multiaddr.Multiaddr
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

	// Set stream handler for operator connections
	host.SetStreamHandler("/spotted/1.0.0", func(s network.Stream) {
		// Get operator ID from the stream
		operatorID := s.Conn().RemotePeer()
		log.Printf("New stream from operator: %s", operatorID)
		
		// Read and process messages
		reader := bufio.NewReader(s)
		for {
			message, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					log.Printf("Stream closed by operator: %s", operatorID)
				} else {
					log.Printf("Error reading from stream: %v", err)
				}
				s.Close()
				return
			}

			message = strings.TrimSpace(message)
			log.Printf("Received message from %s: %s", operatorID, message)

			if !strings.HasPrefix(message, "ANNOUNCE ") {
				log.Printf("Invalid message format from %s, expected ANNOUNCE prefix: %s", operatorID, message)
				continue
			}

			parts := strings.Split(strings.TrimPrefix(message, "ANNOUNCE "), " ")
			if len(parts) != 2 {
				log.Printf("Invalid message format from %s: expected 2 parts, got %d", operatorID, len(parts))
				continue
			}

			// Parse operator addresses
			addrStrs := strings.Split(parts[1], ",")
			var addrs []multiaddr.Multiaddr
			for _, addrStr := range addrStrs {
				addr, err := multiaddr.NewMultiaddr(addrStr)
				if err != nil {
					log.Printf("Invalid address format from %s: %v", operatorID, err)
					continue
				}
				addrs = append(addrs, addr)
			}

			// Add the operator with its addresses
			node.operatorsMu.Lock()
			node.operators[operatorID] = &OperatorInfo{
				ID:       operatorID,
				Addrs:    addrs,
				LastSeen: time.Now(),
				Status:   "active",
			}
			node.operatorsMu.Unlock()

			log.Printf("Operator %s registered with %d addresses", operatorID, len(addrs))

			// Write acknowledgment back to the operator
			if _, err := s.Write([]byte("OK\n")); err != nil {
				log.Printf("Error sending acknowledgment to %s: %v", operatorID, err)
			}

			// After processing ANNOUNCE message, close the stream
			s.Close()
			return
		}
	})

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

	// Check if operator is already registered
	if _, exists := n.operators[id]; exists {
		return
	}

	// Get operator's addresses from the host
	peerInfo := n.host.Peerstore().PeerInfo(id)

	// Add new operator with addresses
	n.operators[id] = &OperatorInfo{
		ID: id,
		Addrs: peerInfo.Addrs,
		LastSeen: time.Now(),
		Status: "active",
	}
	log.Printf("New operator added: %s with addresses: %v\n", id, peerInfo.Addrs)

	// Broadcast new operator to all connected operators
	for opID := range n.operators {
		if opID != id {
			log.Printf("Broadcasting new operator %s to %s\n", id, opID)
			// Send operator info through p2p network
			if err := n.host.SendOperatorInfo(context.Background(), opID, id, peerInfo.Addrs); err != nil {
				log.Printf("Failed to broadcast operator info to %s: %v\n", opID, err)
			}
		}
	}

	// Send existing operators to the new operator
	for existingID, existingInfo := range n.operators {
		if existingID != id {
			log.Printf("Sending existing operator %s to new operator %s\n", existingID, id)
			if err := n.host.SendOperatorInfo(context.Background(), id, existingID, existingInfo.Addrs); err != nil {
				log.Printf("Failed to send existing operator info to %s: %v\n", id, err)
			}
		}
	}
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

// Get connected operators
func (n *Node) GetConnectedOperators() []peer.ID {
	n.operatorsMu.RLock()
	defer n.operatorsMu.RUnlock()

	operators := make([]peer.ID, 0, len(n.operators))
	for id := range n.operators {
		operators = append(operators, id)
	}
	return operators
}

// GetHostID returns the node's libp2p host ID
func (n *Node) GetHostID() string {
	return n.host.ID().String()
} 