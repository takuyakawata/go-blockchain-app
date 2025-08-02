// Pattern 7: P2P Network Layer
// This pattern demonstrates a peer-to-peer network layer for blockchain nodes
// Features:
// - TCP-based P2P communication
// - Message serialization and handling
// - Blockchain synchronization
// - Mempool management
// - Node discovery and management

package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"blockchain-app/network"
	"blockchain-app/transaction"
)

// P2PBlockchain implements the network.BlockchainInterface for P2P layer
type P2PBlockchain struct {
	*transaction.Blockchain
}

// GetBlockHashes returns all block hashes for the P2P layer
func (bc *P2PBlockchain) GetBlockHashes() [][]byte {
	var hashes [][]byte

	// Get all blocks from the blockchain
	iterator := bc.Blockchain.Iterator()

	for {
		block := iterator.Next()
		if block == nil {
			break
		}
		hashes = append(hashes, block.Hash)
	}

	return hashes
}

// GetBlock returns a block by its hash for the P2P layer
func (bc *P2PBlockchain) GetBlock(blockHash []byte) (network.BlockInterface, error) {
	// In a real implementation, this would efficiently look up blocks by hash
	// For now, iterate through all blocks to find the matching hash
	iterator := bc.Blockchain.Iterator()

	for {
		block := iterator.Next()
		if block == nil {
			break
		}

		// Compare hashes
		if string(block.Hash) == string(blockHash) {
			return &P2PBlock{Block: block}, nil
		}
	}

	return nil, fmt.Errorf("block not found")
}

// AddBlock adds a block to the blockchain for the P2P layer
func (bc *P2PBlockchain) AddBlock(block network.BlockInterface) {
	// In a real implementation, this would properly validate and add the block
	fmt.Printf("Would add block with hash: %x\n", block.GetHash())
}

// P2PBlock implements the network.BlockInterface
type P2PBlock struct {
	*transaction.Block
}

// GetHash returns the block hash
func (b *P2PBlock) GetHash() []byte {
	return b.Block.Hash
}

// GetHeight returns the block height
func (b *P2PBlock) GetHeight() int {
	// In a real implementation, blocks would have height information
	return 0
}

// Serialize serializes the block
func (b *P2PBlock) Serialize() []byte {
	return b.Block.Serialize()
}

// CLI functions for blockchain-seven pattern
func startNodeCommand(args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: startnode <port> [bootstrap_nodes...]")
		fmt.Println("Example: startnode 3000")
		fmt.Println("Example: startnode 3001 localhost:3000")
		return
	}

	port := args[1]
	address := "localhost:" + port
	nodeID := "node_" + port

	// Parse bootstrap nodes
	var bootstrapNodes []string
	if len(args) > 2 {
		bootstrapNodes = args[2:]
	}

	fmt.Printf("Starting node %s on %s\n", nodeID, address)
	if len(bootstrapNodes) > 0 {
		fmt.Printf("Bootstrap nodes: %s\n", strings.Join(bootstrapNodes, ", "))
	}

	// Create blockchain
	walletFile := fmt.Sprintf("wallet_%s.dat", port)
	bc := transaction.NewBlockchain(walletFile)
	// Note: In a production environment, we would need to properly close the database

	// Wrap blockchain for P2P interface
	p2pBlockchain := &P2PBlockchain{Blockchain: bc}

	// Create P2P server
	server := network.NewServer(address, nodeID, p2pBlockchain)

	// Set up graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start server in a goroutine
	go func() {
		err := server.Start()
		if err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait a moment for server to start
	time.Sleep(time.Second)

	// Bootstrap to network if nodes provided
	if len(bootstrapNodes) > 0 {
		fmt.Println("Connecting to bootstrap nodes...")
		err := server.Bootstrap(bootstrapNodes)
		if err != nil {
			log.Printf("Bootstrap failed: %v", err)
		}
	}

	// Start a goroutine to display network status
	go displayNetworkStatus(server)

	// Wait for shutdown signal
	<-sigChan
	fmt.Println("\nShutting down node...")

	err := server.Stop()
	if err != nil {
		log.Printf("Error stopping server: %v", err)
	}

	fmt.Println("Node stopped.")
}

// displayNetworkStatus displays network status periodically
func displayNetworkStatus(server *network.Server) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			nodeInfo := server.GetNodeInfo()
			fmt.Println("\n=== Node Status ===")
			fmt.Printf("Address: %s\n", nodeInfo.Address)
			fmt.Printf("Blockchain Height: %d\n", nodeInfo.Height)
			fmt.Printf("Connected Peers: %d/%d\n", nodeInfo.Network.ConnectedPeers, nodeInfo.Network.MaxPeers)
			fmt.Printf("Total Known Peers: %d\n", nodeInfo.Network.TotalPeers)
			fmt.Printf("Mempool Transactions: %d\n", nodeInfo.Mempool.TransactionCount)
			fmt.Printf("Sync Status: %v\n", nodeInfo.SyncStatus.IsSyncing)
			fmt.Println("==================")
		}
	}
}

func nodeInfoCommand(args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: nodeinfo <port>")
		fmt.Println("Example: nodeinfo 3000")
		return
	}

	port := args[1]
	address := "localhost:" + port

	fmt.Printf("Getting node info for %s\n", address)

	// Create a temporary client to connect and get info
	// In a real implementation, this would make a proper API call
	fmt.Printf("Node Address: %s\n", address)
	fmt.Println("Note: Connect to the running node to get detailed info")
}

func connectPeerCommand(args []string) {
	if len(args) < 3 {
		fmt.Println("Usage: connectpeer <local_port> <peer_address>")
		fmt.Println("Example: connectpeer 3000 localhost:3001")
		return
	}

	localPort := args[1]
	peerAddress := args[2]
	localAddress := "localhost:" + localPort

	fmt.Printf("Connecting node %s to peer %s\n", localAddress, peerAddress)

	// This would typically involve sending a message to the running node
	// For now, we'll just display the command
	fmt.Println("Note: This command would send a connect request to the running node")
}

func listPeersCommand(args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: listpeers <port>")
		fmt.Println("Example: listpeers 3000")
		return
	}

	port := args[1]
	address := "localhost:" + port

	fmt.Printf("Listing peers for node %s\n", address)
	fmt.Println("Note: This would query the running node for its peer list")
}

func sendTxCommand(args []string) {
	if len(args) < 5 {
		fmt.Println("Usage: sendtx <port> <from> <to> <amount>")
		fmt.Println("Example: sendtx 3000 alice bob 10")
		return
	}

	port := args[1]
	from := args[2]
	to := args[3]
	amountStr := args[4]

	amount, err := strconv.Atoi(amountStr)
	if err != nil {
		fmt.Printf("Invalid amount: %s\n", amountStr)
		return
	}

	address := "localhost:" + port
	fmt.Printf("Sending transaction through node %s: %s -> %s (%d)\n", address, from, to, amount)
	fmt.Println("Note: This would create and broadcast a transaction through the P2P network")
}

func mineBlockCommand(args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: mineblock <port>")
		fmt.Println("Example: mineblock 3000")
		return
	}

	port := args[1]
	address := "localhost:" + port

	fmt.Printf("Requesting block mining on node %s\n", address)
	fmt.Println("Note: This would request the running node to mine a new block")
}

func syncStatusCommand(args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: syncstatus <port>")
		fmt.Println("Example: syncstatus 3000")
		return
	}

	port := args[1]
	address := "localhost:" + port

	fmt.Printf("Getting sync status for node %s\n", address)
	fmt.Println("Note: This would query the running node's synchronization status")
}

// runBlockchainSeven demonstrates P2P blockchain functionality
func runBlockchainSeven() {
	fmt.Println("=== Blockchain Pattern 7: P2P Network Layer ===")
	fmt.Println("Available commands:")
	fmt.Println("  startnode <port> [bootstrap_nodes...] - Start a P2P node")
	fmt.Println("  nodeinfo <port>                       - Get node information")
	fmt.Println("  connectpeer <local_port> <peer_addr>  - Connect to a peer")
	fmt.Println("  listpeers <port>                      - List connected peers")
	fmt.Println("  sendtx <port> <from> <to> <amount>    - Send transaction")
	fmt.Println("  mineblock <port>                      - Mine a new block")
	fmt.Println("  syncstatus <port>                     - Get sync status")
	fmt.Println("  help                                  - Show this help")
	fmt.Println("  exit                                  - Exit program")
	fmt.Println()
	fmt.Println("Example workflow:")
	fmt.Println("1. startnode 3000")
	fmt.Println("2. In another terminal: startnode 3001 localhost:3000")
	fmt.Println("3. In another terminal: startnode 3002 localhost:3000 localhost:3001")
	fmt.Println()

	for {
		fmt.Print("blockchain7> ")
		var input string
		fmt.Scanln(&input)

		args := strings.Fields(input)
		if len(args) == 0 {
			continue
		}

		command := strings.ToLower(args[0])
		switch command {
		case "startnode":
			startNodeCommand(args)
		case "nodeinfo":
			nodeInfoCommand(args)
		case "connectpeer":
			connectPeerCommand(args)
		case "listpeers":
			listPeersCommand(args)
		case "sendtx":
			sendTxCommand(args)
		case "mineblock":
			mineBlockCommand(args)
		case "syncstatus":
			syncStatusCommand(args)
		case "help":
			fmt.Println("Available commands:")
			fmt.Println("  startnode <port> [bootstrap_nodes...] - Start a P2P node")
			fmt.Println("  nodeinfo <port>                       - Get node information")
			fmt.Println("  connectpeer <local_port> <peer_addr>  - Connect to a peer")
			fmt.Println("  listpeers <port>                      - List connected peers")
			fmt.Println("  sendtx <port> <from> <to> <amount>    - Send transaction")
			fmt.Println("  mineblock <port>                      - Mine a new block")
			fmt.Println("  syncstatus <port>                     - Get sync status")
			fmt.Println("  help                                  - Show this help")
			fmt.Println("  exit                                  - Exit program")
		case "exit":
			fmt.Println("Goodbye!")
			return
		default:
			fmt.Printf("Unknown command: %s (type 'help' for available commands)\n", command)
		}
	}
}
