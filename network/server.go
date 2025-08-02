package network

import (
	"fmt"
	"io"
	"log"
	"net"
	"sync"
)

// Server represents a P2P network server
type Server struct {
	Address     string
	NodeID      string
	Blockchain  BlockchainInterface
	KnownNodes  map[string]bool
	Mempool     map[string]TransactionInterface
	NodeManager *NodeManager
	MempoolMgr  *MempoolManager
	SyncMgr     *SyncManager
	mu          sync.RWMutex
	running     bool
	listener    net.Listener
}

// BlockchainInterface defines required blockchain methods
type BlockchainInterface interface {
	GetBestHeight() int
	GetBlockHashes() [][]byte
	GetBlock(blockHash []byte) (BlockInterface, error)
	AddBlock(block BlockInterface)
}

// BlockInterface defines required block methods
type BlockInterface interface {
	GetHash() []byte
	GetHeight() int
	Serialize() []byte
}

// TransactionInterface defines required transaction methods
type TransactionInterface interface {
	GetID() []byte
	Serialize() []byte
}

// NewServer creates a new P2P server
func NewServer(address, nodeID string, blockchain BlockchainInterface) *Server {
	server := &Server{
		Address:    address,
		NodeID:     nodeID,
		Blockchain: blockchain,
		KnownNodes: make(map[string]bool),
		Mempool:    make(map[string]TransactionInterface),
		running:    false,
	}

	// Initialize managers
	server.NodeManager = NewNodeManager(server)
	server.MempoolMgr = NewMempoolManager(server)
	server.SyncMgr = NewSyncManager(server)

	return server
}

// Start starts the P2P server
func (s *Server) Start() error {
	listener, err := net.Listen("tcp", s.Address)
	if err != nil {
		return fmt.Errorf("failed to start server: %v", err)
	}

	s.listener = listener
	s.running = true

	// Start the managers
	s.NodeManager.Start()
	s.MempoolMgr.StartCleanupRoutine()
	s.SyncMgr.StartPeriodicSync()

	fmt.Printf("Node %s listening on %s\n", s.NodeID, s.Address)

	for s.running {
		conn, err := listener.Accept()
		if err != nil {
			if s.running {
				log.Printf("Failed to accept connection: %v", err)
			}
			continue
		}

		go s.HandleConnection(conn)
	}

	return nil
}

// Stop stops the P2P server
func (s *Server) Stop() error {
	s.running = false

	// Stop the managers
	if s.NodeManager != nil {
		s.NodeManager.Stop()
	}

	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

// HandleConnection handles incoming connections
func (s *Server) HandleConnection(conn net.Conn) {
	defer conn.Close()

	remoteAddr := conn.RemoteAddr().String()
	fmt.Printf("New connection from %s\n", remoteAddr)

	// Add to known nodes and node manager
	s.mu.Lock()
	s.KnownNodes[remoteAddr] = true
	s.mu.Unlock()

	if s.NodeManager != nil {
		s.NodeManager.AddPeer(remoteAddr)
	}

	for {
		message, err := ReadMessage(conn)
		if err != nil {
			if err != io.EOF {
				log.Printf("Error reading message: %v", err)
			}
			break
		}

		s.ProcessMessage(message, conn)
	}

	// Remove from known nodes and node manager
	s.mu.Lock()
	delete(s.KnownNodes, remoteAddr)
	s.mu.Unlock()

	if s.NodeManager != nil {
		s.NodeManager.DisconnectFromPeer(remoteAddr)
	}

	fmt.Printf("Connection from %s closed\n", remoteAddr)
}

// ProcessMessage processes incoming messages
func (s *Server) ProcessMessage(msg Message, conn net.Conn) {
	fmt.Printf("Received %s message\n", msg.Command)

	switch msg.Command {
	case CmdVersion:
		s.HandleVersion(msg.Data, conn)
	case CmdGetBlocks:
		s.HandleGetBlocks(msg.Data, conn)
	case CmdInv:
		s.HandleInv(msg.Data, conn)
	case CmdGetData:
		s.HandleGetData(msg.Data, conn)
	case CmdBlock:
		s.HandleBlock(msg.Data, conn)
	case CmdTx:
		s.HandleTx(msg.Data, conn)
	case CmdPing:
		s.HandlePing(msg.Data, conn)
	default:
		fmt.Printf("Unknown command: %s\n", msg.Command)
	}
}

// SendMessage sends a message to a specific address
func (s *Server) SendMessage(address string, msg Message) error {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %v", address, err)
	}
	defer conn.Close()

	return WriteMessage(conn, msg)
}

// BroadcastMessage broadcasts a message to all known nodes
func (s *Server) BroadcastMessage(msg Message) {
	s.mu.RLock()
	nodes := make([]string, 0, len(s.KnownNodes))
	for node := range s.KnownNodes {
		nodes = append(nodes, node)
	}
	s.mu.RUnlock()

	for _, node := range nodes {
		go func(addr string) {
			err := s.SendMessage(addr, msg)
			if err != nil {
				log.Printf("Failed to send message to %s: %v", addr, err)
				// Remove failed node
				s.mu.Lock()
				delete(s.KnownNodes, addr)
				s.mu.Unlock()
			}
		}(node)
	}
}

// ConnectToPeer connects to a peer node
func (s *Server) ConnectToPeer(address string) error {
	// Send version message to establish connection
	versionData := VersionData{
		Version:    1,
		BestHeight: int32(s.Blockchain.GetBestHeight()),
		AddrFrom:   s.Address,
	}

	msg := Message{
		Command: CmdVersion,
		Data:    GobEncode(versionData),
	}

	return s.SendMessage(address, msg)
}

// GetKnownNodes returns the list of known nodes
func (s *Server) GetKnownNodes() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	nodes := make([]string, 0, len(s.KnownNodes))
	for node := range s.KnownNodes {
		nodes = append(nodes, node)
	}
	return nodes
}

// Bootstrap connects to bootstrap nodes to join the network
func (s *Server) Bootstrap(bootstrapNodes []string) error {
	if s.NodeManager != nil {
		if len(bootstrapNodes) > 0 {
			s.NodeManager.SetBootstrapNodes(bootstrapNodes)
		}
		return s.NodeManager.Bootstrap()
	}
	return fmt.Errorf("node manager not initialized")
}

// GetNodeInfo returns node information including peer statistics
func (s *Server) GetNodeInfo() NodeInfo {
	var networkInfo NetworkInfo
	if s.NodeManager != nil {
		networkInfo = s.NodeManager.GetNetworkInfo()
	}

	var mempoolInfo MempoolInfo
	if s.MempoolMgr != nil {
		mempoolInfo = s.MempoolMgr.GetMempoolInfo()
	}

	var syncStatus SyncStatus
	if s.SyncMgr != nil {
		syncStatus = s.SyncMgr.GetSyncStatus()
	}

	return NodeInfo{
		Address:    s.Address,
		NodeID:     s.NodeID,
		IsRunning:  s.running,
		Height:     s.Blockchain.GetBestHeight(),
		Network:    networkInfo,
		Mempool:    mempoolInfo,
		SyncStatus: syncStatus,
	}
}

// NodeInfo contains comprehensive node information
type NodeInfo struct {
	Address    string
	NodeID     string
	IsRunning  bool
	Height     int
	Network    NetworkInfo
	Mempool    MempoolInfo
	SyncStatus SyncStatus
}
