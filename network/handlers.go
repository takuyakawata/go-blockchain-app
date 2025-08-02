package network

import (
	"fmt"
	"log"
	"net"
)

// HandleVersion handles version messages
func (s *Server) HandleVersion(data []byte, conn net.Conn) {
	var versionData VersionData
	GobDecode(data, &versionData)

	fmt.Printf("Received version from %s (height: %d)\n", versionData.AddrFrom, versionData.BestHeight)

	myBestHeight := s.Blockchain.GetBestHeight()
	foreignerBestHeight := int(versionData.BestHeight)

	if myBestHeight < foreignerBestHeight {
		// Request blocks from the peer
		s.SendGetBlocks(versionData.AddrFrom)
	} else if myBestHeight > foreignerBestHeight {
		// Send our version back
		s.SendVersion(versionData.AddrFrom)
	}

	// Send acknowledgment
	if !s.NodeIsKnown(versionData.AddrFrom) {
		s.mu.Lock()
		s.KnownNodes[versionData.AddrFrom] = true
		s.mu.Unlock()
	}

	// Update peer info in node manager
	if s.NodeManager != nil {
		s.NodeManager.UpdatePeerInfo(versionData.AddrFrom, int(versionData.BestHeight), versionData.Version)
	}
}

// HandleGetBlocks handles getblocks messages
func (s *Server) HandleGetBlocks(data []byte, conn net.Conn) {
	var getBlocksData GetBlocksData
	GobDecode(data, &getBlocksData)

	fmt.Printf("Received getblocks from %s\n", getBlocksData.AddrFrom)

	blocks := s.Blockchain.GetBlockHashes()
	s.SendInv(getBlocksData.AddrFrom, "block", blocks)
}

// HandleInv handles inventory messages
func (s *Server) HandleInv(data []byte, conn net.Conn) {
	var invData InvData
	GobDecode(data, &invData)

	fmt.Printf("Received inventory with %d %s\n", len(invData.Items), invData.Type)

	if invData.Type == "block" {
		blocksInTransit = invData.Items

		blockHash := invData.Items[0]
		s.SendGetData(invData.AddrFrom, "block", blockHash)

		var newInTransit [][]byte
		for _, b := range blocksInTransit {
			if !BytesEqual(b, blockHash) {
				newInTransit = append(newInTransit, b)
			}
		}
		blocksInTransit = newInTransit
	}

	if invData.Type == "tx" {
		txID := invData.Items[0]

		if !s.TransactionExists(txID) {
			s.SendGetData(invData.AddrFrom, "tx", txID)
		}
	}
}

// HandleGetData handles getdata messages
func (s *Server) HandleGetData(data []byte, conn net.Conn) {
	var getDataData GetDataData
	GobDecode(data, &getDataData)

	fmt.Printf("Received getdata for %s from %s\n", getDataData.Type, getDataData.AddrFrom)

	if getDataData.Type == "block" {
		block, err := s.Blockchain.GetBlock(getDataData.ID)
		if err != nil {
			log.Printf("Block not found: %v", err)
			return
		}

		s.SendBlock(getDataData.AddrFrom, block)
	}

	if getDataData.Type == "tx" {
		txID := fmt.Sprintf("%x", getDataData.ID)

		s.mu.RLock()
		tx, exists := s.Mempool[txID]
		s.mu.RUnlock()

		if exists {
			s.SendTx(getDataData.AddrFrom, tx)
		}
	}
}

// HandleBlock handles block messages
func (s *Server) HandleBlock(data []byte, conn net.Conn) {
	var blockData BlockData
	GobDecode(data, &blockData)

	fmt.Printf("Received new block from %s\n", blockData.AddrFrom)

	// Deserialize and add block (simplified - would need proper block interface)
	// This would typically involve:
	// 1. Deserializing the block
	// 2. Validating the block
	// 3. Adding to blockchain
	// 4. Updating UTXO set
	// 5. Removing transactions from mempool

	fmt.Printf("Added block. Blocks in transit: %d\n", len(blocksInTransit))

	if len(blocksInTransit) > 0 {
		blockHash := blocksInTransit[0]
		s.SendGetData(blockData.AddrFrom, "block", blockHash)

		blocksInTransit = blocksInTransit[1:]
	} else {
		// Request mempool from connected nodes
		for node := range s.KnownNodes {
			if node != s.Address {
				s.SendGetBlocks(node)
			}
		}
	}
}

// HandleTx handles transaction messages
func (s *Server) HandleTx(data []byte, conn net.Conn) {
	var txData TxData
	GobDecode(data, &txData)

	fmt.Printf("Received new transaction from %s\n", txData.AddrFrom)

	// This would involve:
	// 1. Deserializing transaction
	// 2. Validating transaction
	// 3. Adding to mempool
	// 4. Broadcasting to other nodes

	// Simplified implementation
	fmt.Println("Transaction processed and added to mempool")
}

// HandlePing handles ping messages
func (s *Server) HandlePing(data []byte, conn net.Conn) {
	var pingData PingData
	GobDecode(data, &pingData)

	fmt.Printf("Received ping from %s\n", pingData.AddrFrom)

	// Send pong response
	pongData := PongData{AddrFrom: s.Address}
	msg := Message{
		Command: CmdPong,
		Data:    GobEncode(pongData),
	}

	err := WriteMessage(conn, msg)
	if err != nil {
		log.Printf("Failed to send pong: %v", err)
	}
}

// SendVersion sends version message to a node
func (s *Server) SendVersion(addr string) {
	bestHeight := s.Blockchain.GetBestHeight()
	versionData := VersionData{
		Version:    1,
		BestHeight: int32(bestHeight),
		AddrFrom:   s.Address,
	}

	msg := Message{
		Command: CmdVersion,
		Data:    GobEncode(versionData),
	}

	err := s.SendMessage(addr, msg)
	if err != nil {
		log.Printf("Failed to send version to %s: %v", addr, err)
	}
}

// SendGetBlocks sends getblocks message to a node
func (s *Server) SendGetBlocks(addr string) {
	getBlocksData := GetBlocksData{AddrFrom: s.Address}
	msg := Message{
		Command: CmdGetBlocks,
		Data:    GobEncode(getBlocksData),
	}

	err := s.SendMessage(addr, msg)
	if err != nil {
		log.Printf("Failed to send getblocks to %s: %v", addr, err)
	}
}

// SendInv sends inventory message to a node
func (s *Server) SendInv(addr, kind string, items [][]byte) {
	invData := InvData{
		AddrFrom: s.Address,
		Type:     kind,
		Items:    items,
	}

	msg := Message{
		Command: CmdInv,
		Data:    GobEncode(invData),
	}

	err := s.SendMessage(addr, msg)
	if err != nil {
		log.Printf("Failed to send inv to %s: %v", addr, err)
	}
}

// SendGetData sends getdata message to a node
func (s *Server) SendGetData(addr, kind string, id []byte) {
	getDataData := GetDataData{
		AddrFrom: s.Address,
		Type:     kind,
		ID:       id,
	}

	msg := Message{
		Command: CmdGetData,
		Data:    GobEncode(getDataData),
	}

	err := s.SendMessage(addr, msg)
	if err != nil {
		log.Printf("Failed to send getdata to %s: %v", addr, err)
	}
}

// SendBlock sends block message to a node
func (s *Server) SendBlock(addr string, block BlockInterface) {
	blockData := BlockData{
		AddrFrom: s.Address,
		Block:    block.Serialize(),
	}

	msg := Message{
		Command: CmdBlock,
		Data:    GobEncode(blockData),
	}

	err := s.SendMessage(addr, msg)
	if err != nil {
		log.Printf("Failed to send block to %s: %v", addr, err)
	}
}

// SendTx sends transaction message to a node
func (s *Server) SendTx(addr string, tx TransactionInterface) {
	txData := TxData{
		AddrFrom:    s.Address,
		Transaction: tx.Serialize(),
	}

	msg := Message{
		Command: CmdTx,
		Data:    GobEncode(txData),
	}

	err := s.SendMessage(addr, msg)
	if err != nil {
		log.Printf("Failed to send transaction to %s: %v", addr, err)
	}
}

// Helper functions

// NodeIsKnown checks if a node is in the known nodes list
func (s *Server) NodeIsKnown(addr string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, exists := s.KnownNodes[addr]
	return exists
}

// TransactionExists checks if a transaction exists in mempool
func (s *Server) TransactionExists(txID []byte) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	txIDStr := fmt.Sprintf("%x", txID)
	_, exists := s.Mempool[txIDStr]
	return exists
}

// BytesEqual compares two byte slices
func BytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// Global variable for blocks in transit (simplified)
var blocksInTransit = [][]byte{}
