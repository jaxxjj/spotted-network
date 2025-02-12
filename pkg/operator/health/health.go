package health

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/protocol/ping"
)

const (
	HealthCheckInterval = 20 * time.Second
)

type PingService interface {
	Ping(ctx context.Context, p peer.ID) <-chan ping.Result
}

type Node interface {
	DisconnectPeer(peer.ID) error
	GetConnectedPeers() []peer.ID
	PrintConnectedPeers()
}

// HealthService 定义健康检查服务的接口
type HealthService interface {
	// Stop 停止健康检查服务
	Stop()
	// GetStatus 获取当前健康状态
	GetStatus() map[peer.ID]time.Duration
	// TriggerCheck 手动触发一次健康检查
	TriggerCheck(ctx context.Context) error
	// SetCheckInterval 设置健康检查间隔
	SetCheckInterval(d time.Duration)
}

// HealthChecker 实现 HealthService 接口
type healthChecker struct {
	node          Node
	pingService   PingService
	cancel        context.CancelFunc
	checkInterval time.Duration
	statusMu      sync.RWMutex
	peerStatus    map[peer.ID]time.Duration
}

// NewHealthChecker 创建并启动一个新的健康检查器
func NewHealthChecker(ctx context.Context, node Node, pingService PingService) (HealthService, error) {
	if node == nil {
		return nil, fmt.Errorf("node is nil")
	}
	if pingService == nil {
		return nil, fmt.Errorf("ping service is nil")
	}

	ctx, cancel := context.WithCancel(ctx)

	hc := &healthChecker{
		node:          node,
		pingService:   pingService,
		cancel:        cancel,
		checkInterval: HealthCheckInterval,
		peerStatus:    make(map[peer.ID]time.Duration),
	}

	go hc.start(ctx)
	log.Printf("[Health] Health check service started with interval %v", hc.checkInterval)

	return hc, nil
}

// Stop implements HealthService
func (hc *healthChecker) Stop() {
	if hc.cancel != nil {
		hc.cancel()
	}
}

// GetStatus implements HealthService
func (hc *healthChecker) GetStatus() map[peer.ID]time.Duration {
	hc.statusMu.RLock()
	defer hc.statusMu.RUnlock()

	status := make(map[peer.ID]time.Duration)
	for k, v := range hc.peerStatus {
		status[k] = v
	}
	return status
}

// TriggerCheck implements HealthService
func (hc *healthChecker) TriggerCheck(ctx context.Context) error {
	hc.checkPeers(ctx)
	return nil
}

// SetCheckInterval implements HealthService
func (hc *healthChecker) SetCheckInterval(d time.Duration) {
	hc.checkInterval = d
}

// --- 内部方法 ---

func (hc *healthChecker) start(ctx context.Context) {
	ticker := time.NewTicker(hc.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			hc.checkPeers(ctx)
			hc.node.PrintConnectedPeers()
		}
	}
}

func (hc *healthChecker) checkPeers(ctx context.Context) {
	peers := hc.node.GetConnectedPeers()
	for _, peerID := range peers {
		if err := hc.pingPeer(ctx, peerID); err != nil {
			log.Printf("[Health] Peer %s failed health check: %v", peerID, err)
			hc.node.DisconnectPeer(peerID)
		}
	}
}

func (hc *healthChecker) pingPeer(ctx context.Context, p peer.ID) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	result := <-hc.pingService.Ping(ctx, p)
	if result.Error != nil {
		return result.Error
	}

	hc.statusMu.Lock()
	hc.peerStatus[p] = result.RTT
	hc.statusMu.Unlock()

	log.Printf("[Health] Successfully pinged peer %s (RTT: %v)", p, result.RTT)
	return nil
}
