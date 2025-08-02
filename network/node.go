package network

import (
	"fmt"
	"log"
	"sync"
	"time"
)

// NodeManager manages peer nodes and bootstrap functionality
type NodeManager struct {
	server         *Server
	peers          map[string]*Peer
	bootstrapNodes []string
	maxPeers       int
	mu             sync.RWMutex
	running        bool
}

// Peer represents a connected peer node
type Peer struct {
	Address   string
	ID        string
	LastSeen  time.Time
	Height    int
	Status    PeerStatus
	Latency   time.Duration
	Connected bool
	Version   int32
}

// PeerStatus represents the status of a peer
type PeerStatus int

const (
	PeerStatusConnecting PeerStatus = iota
	PeerStatusConnected
	PeerStatusDisconnected
	PeerStatusBanned
)

// NewNodeManager creates a new node manager
func NewNodeManager(server *Server) *NodeManager {
	return &NodeManager{
		server: server,
		peers:  make(map[string]*Peer),
		bootstrapNodes: []string{
			"localhost:3001", // Default bootstrap nodes
			"localhost:3002",
		},
		maxPeers: 8, // Maximum number of peers
		running:  false,
	}
}

// Start starts the node manager
func (nm *NodeManager) Start() {
	nm.mu.Lock()
	nm.running = true
	nm.mu.Unlock()

	fmt.Println("Node manager started")

	// Start peer discovery routine
	go nm.startPeerDiscovery()

	// Start peer maintenance routine
	go nm.startPeerMaintenance()

	// Start health check routine
	go nm.startHealthCheck()
}

// Stop stops the node manager
func (nm *NodeManager) Stop() {
	nm.mu.Lock()
	nm.running = false
	nm.mu.Unlock()

	fmt.Println("Node manager stopped")
}

// Bootstrap connects to bootstrap nodes to join the network
func (nm *NodeManager) Bootstrap() error {
	fmt.Println("Starting bootstrap process...")

	if len(nm.bootstrapNodes) == 0 {
		return fmt.Errorf("no bootstrap nodes configured")
	}

	connectedCount := 0
	for _, node := range nm.bootstrapNodes {
		if node == nm.server.Address {
			continue // Skip self
		}

		fmt.Printf("Attempting to connect to bootstrap node: %s\n", node)

		err := nm.ConnectToPeer(node)
		if err != nil {
			log.Printf("Failed to connect to bootstrap node %s: %v", node, err)
			continue
		}

		connectedCount++
		fmt.Printf("Connected to bootstrap node: %s\n", node)

		// Try to connect to a few bootstrap nodes, not all
		if connectedCount >= 2 {
			break
		}
	}

	if connectedCount == 0 {
		return fmt.Errorf("failed to connect to any bootstrap nodes")
	}

	fmt.Printf("Bootstrap completed, connected to %d nodes\n", connectedCount)
	return nil
}

// ConnectToPeer connects to a specific peer
func (nm *NodeManager) ConnectToPeer(address string) error {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	// Check if already connected or trying to connect to self
	if address == nm.server.Address {
		return fmt.Errorf("cannot connect to self")
	}

	if peer, exists := nm.peers[address]; exists {
		if peer.Connected {
			return fmt.Errorf("already connected to %s", address)
		}
	}

	// Check max peers limit
	if nm.getConnectedPeerCount() >= nm.maxPeers {
		return fmt.Errorf("maximum peer limit reached (%d)", nm.maxPeers)
	}

	// Create peer entry
	peer := &Peer{
		Address:   address,
		LastSeen:  time.Now(),
		Status:    PeerStatusConnecting,
		Connected: false,
	}
	nm.peers[address] = peer

	// Attempt connection
	err := nm.server.ConnectToPeer(address)
	if err != nil {
		peer.Status = PeerStatusDisconnected
		return fmt.Errorf("failed to connect to %s: %v", address, err)
	}

	peer.Connected = true
	peer.Status = PeerStatusConnected
	peer.LastSeen = time.Now()

	fmt.Printf("Successfully connected to peer: %s\n", address)
	return nil
}

// DisconnectFromPeer disconnects from a specific peer
func (nm *NodeManager) DisconnectFromPeer(address string) {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	if peer, exists := nm.peers[address]; exists {
		peer.Connected = false
		peer.Status = PeerStatusDisconnected
		fmt.Printf("Disconnected from peer: %s\n", address)
	}
}

// AddPeer adds a peer to the peer list
func (nm *NodeManager) AddPeer(address string) {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	if address == nm.server.Address {
		return // Don't add self
	}

	if _, exists := nm.peers[address]; !exists {
		peer := &Peer{
			Address:   address,
			LastSeen:  time.Now(),
			Status:    PeerStatusDisconnected,
			Connected: false,
		}
		nm.peers[address] = peer
		fmt.Printf("Added new peer: %s\n", address)
	}
}

// RemovePeer removes a peer from the peer list
func (nm *NodeManager) RemovePeer(address string) {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	if _, exists := nm.peers[address]; exists {
		delete(nm.peers, address)
		fmt.Printf("Removed peer: %s\n", address)
	}
}

// GetConnectedPeers returns all connected peers
func (nm *NodeManager) GetConnectedPeers() []*Peer {
	nm.mu.RLock()
	defer nm.mu.RUnlock()

	var connectedPeers []*Peer
	for _, peer := range nm.peers {
		if peer.Connected && peer.Status == PeerStatusConnected {
			connectedPeers = append(connectedPeers, peer)
		}
	}
	return connectedPeers
}

// GetAllPeers returns all known peers
func (nm *NodeManager) GetAllPeers() []*Peer {
	nm.mu.RLock()
	defer nm.mu.RUnlock()

	var allPeers []*Peer
	for _, peer := range nm.peers {
		allPeers = append(allPeers, peer)
	}
	return allPeers
}

// GetPeerCount returns the total number of known peers
func (nm *NodeManager) GetPeerCount() int {
	nm.mu.RLock()
	defer nm.mu.RUnlock()
	return len(nm.peers)
}

// GetConnectedPeerCount returns the number of connected peers
func (nm *NodeManager) GetConnectedPeerCount() int {
	nm.mu.RLock()
	defer nm.mu.RUnlock()
	return nm.getConnectedPeerCount()
}

// getConnectedPeerCount returns connected peer count (internal, no lock)
func (nm *NodeManager) getConnectedPeerCount() int {
	count := 0
	for _, peer := range nm.peers {
		if peer.Connected && peer.Status == PeerStatusConnected {
			count++
		}
	}
	return count
}

// UpdatePeerInfo updates peer information
func (nm *NodeManager) UpdatePeerInfo(address string, height int, version int32) {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	if peer, exists := nm.peers[address]; exists {
		peer.Height = height
		peer.Version = version
		peer.LastSeen = time.Now()
		peer.Connected = true
		peer.Status = PeerStatusConnected
	} else {
		// Add new peer
		peer := &Peer{
			Address:   address,
			Height:    height,
			Version:   version,
			LastSeen:  time.Now(),
			Status:    PeerStatusConnected,
			Connected: true,
		}
		nm.peers[address] = peer
	}
}

// BanPeer bans a peer for misbehavior
func (nm *NodeManager) BanPeer(address string, reason string) {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	if peer, exists := nm.peers[address]; exists {
		peer.Status = PeerStatusBanned
		peer.Connected = false
		fmt.Printf("Banned peer %s: %s\n", address, reason)
	}
}

// IsPeerBanned checks if a peer is banned
func (nm *NodeManager) IsPeerBanned(address string) bool {
	nm.mu.RLock()
	defer nm.mu.RUnlock()

	if peer, exists := nm.peers[address]; exists {
		return peer.Status == PeerStatusBanned
	}
	return false
}

// GetBestPeers returns peers with highest blockchain height
func (nm *NodeManager) GetBestPeers(limit int) []*Peer {
	nm.mu.RLock()
	defer nm.mu.RUnlock()

	var bestPeers []*Peer
	for _, peer := range nm.peers {
		if peer.Connected && peer.Status == PeerStatusConnected {
			bestPeers = append(bestPeers, peer)
		}
	}

	// Simple sorting by height (in reality would use proper sorting)
	// For now, just return first 'limit' connected peers
	if len(bestPeers) > limit {
		bestPeers = bestPeers[:limit]
	}

	return bestPeers
}

// startPeerDiscovery starts the peer discovery routine
func (nm *NodeManager) startPeerDiscovery() {
	ticker := time.NewTicker(30 * time.Second) // Discover peers every 30 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if !nm.running {
				return
			}
			nm.discoverPeers()
		}
	}
}

// discoverPeers discovers new peers through connected peers
func (nm *NodeManager) discoverPeers() {
	connectedPeers := nm.GetConnectedPeers()

	if len(connectedPeers) < nm.maxPeers/2 {
		fmt.Println("Discovering new peers...")

		// In a real implementation, this would:
		// 1. Ask connected peers for their peer lists
		// 2. Try to connect to new peers
		// 3. Maintain a diverse set of connections

		// Simplified: try to maintain minimum connections
		if len(connectedPeers) < 2 && len(nm.bootstrapNodes) > 0 {
			for _, bootstrap := range nm.bootstrapNodes {
				if bootstrap != nm.server.Address {
					err := nm.ConnectToPeer(bootstrap)
					if err == nil {
						break
					}
				}
			}
		}
	}
}

// startPeerMaintenance starts the peer maintenance routine
func (nm *NodeManager) startPeerMaintenance() {
	ticker := time.NewTicker(60 * time.Second) // Maintain peers every minute
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if !nm.running {
				return
			}
			nm.maintainPeers()
		}
	}
}

// maintainPeers maintains peer connections and removes stale peers
func (nm *NodeManager) maintainPeers() {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	now := time.Now()
	staleThreshold := 5 * time.Minute

	var stalePeers []string
	for address, peer := range nm.peers {
		if now.Sub(peer.LastSeen) > staleThreshold && peer.Status != PeerStatusBanned {
			stalePeers = append(stalePeers, address)
		}
	}

	// Remove stale peers
	for _, address := range stalePeers {
		delete(nm.peers, address)
		fmt.Printf("Removed stale peer: %s\n", address)
	}

	if len(stalePeers) > 0 {
		fmt.Printf("Removed %d stale peers\n", len(stalePeers))
	}
}

// startHealthCheck starts the health check routine
func (nm *NodeManager) startHealthCheck() {
	ticker := time.NewTicker(30 * time.Second) // Health check every 30 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if !nm.running {
				return
			}
			nm.performHealthCheck()
		}
	}
}

// performHealthCheck performs health checks on connected peers
func (nm *NodeManager) performHealthCheck() {
	connectedPeers := nm.GetConnectedPeers()

	for _, peer := range connectedPeers {
		// Send ping to check if peer is alive
		go nm.pingPeer(peer.Address)
	}
}

// pingPeer sends a ping message to a peer
func (nm *NodeManager) pingPeer(address string) {
	pingData := PingData{
		AddrFrom: nm.server.Address,
	}

	msg := Message{
		Command: CmdPing,
		Data:    GobEncode(pingData),
	}

	err := nm.server.SendMessage(address, msg)
	if err != nil {
		log.Printf("Failed to ping peer %s: %v", address, err)
		nm.DisconnectFromPeer(address)
	}
}

// GetNetworkInfo returns network information
func (nm *NodeManager) GetNetworkInfo() NetworkInfo {
	nm.mu.RLock()
	defer nm.mu.RUnlock()

	connectedCount := nm.getConnectedPeerCount()

	return NetworkInfo{
		TotalPeers:     len(nm.peers),
		ConnectedPeers: connectedCount,
		MaxPeers:       nm.maxPeers,
		IsRunning:      nm.running,
		BootstrapNodes: nm.bootstrapNodes,
	}
}

// NetworkInfo contains network statistics
type NetworkInfo struct {
	TotalPeers     int
	ConnectedPeers int
	MaxPeers       int
	IsRunning      bool
	BootstrapNodes []string
}

// SetBootstrapNodes sets the bootstrap nodes
func (nm *NodeManager) SetBootstrapNodes(nodes []string) {
	nm.mu.Lock()
	defer nm.mu.Unlock()
	nm.bootstrapNodes = nodes
	fmt.Printf("Set bootstrap nodes: %v\n", nodes)
}

// AddBootstrapNode adds a bootstrap node
func (nm *NodeManager) AddBootstrapNode(node string) {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	// Check if already exists
	for _, existing := range nm.bootstrapNodes {
		if existing == node {
			return
		}
	}

	nm.bootstrapNodes = append(nm.bootstrapNodes, node)
	fmt.Printf("Added bootstrap node: %s\n", node)
}

// String returns string representation of peer status
func (ps PeerStatus) String() string {
	switch ps {
	case PeerStatusConnecting:
		return "connecting"
	case PeerStatusConnected:
		return "connected"
	case PeerStatusDisconnected:
		return "disconnected"
	case PeerStatusBanned:
		return "banned"
	default:
		return "unknown"
	}
}
