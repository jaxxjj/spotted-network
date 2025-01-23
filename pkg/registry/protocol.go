package registry

import (
	"log"
)

// SetupProtocols sets up the protocol handlers for the registry node
func (n *Node) SetupProtocols() {
	// Set up join request handler
	n.host.SetStreamHandler("/spotted/1.0.0", n.handleStream)
	
	// Set up state sync handler
	n.host.SetStreamHandler("/state-sync/1.0.0", n.handleStateSync)
	
	log.Printf("Protocol handlers set up")
} 