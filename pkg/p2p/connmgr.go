package p2p

import (
	"github.com/libp2p/go-libp2p/p2p/net/connmgr"
)

func NewConnectionManager() (*connmgr.BasicConnMgr, error) {
	// Configure the connection manager
	return connmgr.NewConnManager(
		20, // Low water mark
		60, // High water mark
		connmgr.WithGracePeriod(20), // Grace period in seconds
	)
} 