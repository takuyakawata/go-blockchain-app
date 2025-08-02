package network

import (
	"fmt"
	"sync"
	"time"
)

// MempoolManager manages unconfirmed transactions
type MempoolManager struct {
	server       *Server
	transactions map[string]MempoolTransaction
	mu           sync.RWMutex
	maxSize      int
	timeout      time.Duration
}

// MempoolTransaction represents a transaction in the mempool
type MempoolTransaction struct {
	Transaction TransactionInterface
	Timestamp   time.Time
	Fees        int64
	Size        int
	Verified    bool
}

// NewMempoolManager creates a new mempool manager
func NewMempoolManager(server *Server) *MempoolManager {
	return &MempoolManager{
		server:       server,
		transactions: make(map[string]MempoolTransaction),
		maxSize:      1000,           // Maximum number of transactions
		timeout:      24 * time.Hour, // Transaction timeout
	}
}

// AddTransaction adds a transaction to the mempool
func (mm *MempoolManager) AddTransaction(tx TransactionInterface) error {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	txID := fmt.Sprintf("%x", tx.GetID())

	// Check if transaction already exists
	if _, exists := mm.transactions[txID]; exists {
		return fmt.Errorf("transaction %s already in mempool", txID)
	}

	// Check mempool size limit
	if len(mm.transactions) >= mm.maxSize {
		// Remove oldest transaction
		mm.evictOldestTransaction()
	}

	// Add transaction to mempool
	mempoolTx := MempoolTransaction{
		Transaction: tx,
		Timestamp:   time.Now(),
		Fees:        mm.calculateFees(tx),
		Size:        len(tx.Serialize()),
		Verified:    true, // Simplified - would verify transaction here
	}

	mm.transactions[txID] = mempoolTx

	fmt.Printf("Added transaction %s to mempool (size: %d)\n", txID[:8], len(mm.transactions))

	// Broadcast transaction to network
	mm.broadcastTransaction(tx)

	return nil
}

// RemoveTransaction removes a transaction from the mempool
func (mm *MempoolManager) RemoveTransaction(txID []byte) {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	txIDStr := fmt.Sprintf("%x", txID)
	delete(mm.transactions, txIDStr)

	fmt.Printf("Removed transaction %s from mempool\n", txIDStr[:8])
}

// GetTransaction gets a transaction from the mempool
func (mm *MempoolManager) GetTransaction(txID []byte) (TransactionInterface, bool) {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	txIDStr := fmt.Sprintf("%x", txID)
	mempoolTx, exists := mm.transactions[txIDStr]
	if !exists {
		return nil, false
	}

	return mempoolTx.Transaction, true
}

// GetAllTransactions returns all transactions in the mempool
func (mm *MempoolManager) GetAllTransactions() []TransactionInterface {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	transactions := make([]TransactionInterface, 0, len(mm.transactions))
	for _, mempoolTx := range mm.transactions {
		transactions = append(transactions, mempoolTx.Transaction)
	}

	return transactions
}

// GetTransactionsByFees returns transactions sorted by fees (highest first)
func (mm *MempoolManager) GetTransactionsByFees(limit int) []TransactionInterface {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	// Create a slice of mempool transactions
	mempoolTxs := make([]MempoolTransaction, 0, len(mm.transactions))
	for _, tx := range mm.transactions {
		mempoolTxs = append(mempoolTxs, tx)
	}

	// Sort by fees (simplified - would use proper sorting)
	// For now, just return first 'limit' transactions
	result := make([]TransactionInterface, 0, limit)
	count := 0
	for _, mempoolTx := range mempoolTxs {
		if count >= limit {
			break
		}
		result = append(result, mempoolTx.Transaction)
		count++
	}

	return result
}

// RemoveConfirmedTransactions removes transactions that have been confirmed in a block
func (mm *MempoolManager) RemoveConfirmedTransactions(confirmedTxs [][]byte) {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	removedCount := 0
	for _, txID := range confirmedTxs {
		txIDStr := fmt.Sprintf("%x", txID)
		if _, exists := mm.transactions[txIDStr]; exists {
			delete(mm.transactions, txIDStr)
			removedCount++
		}
	}

	if removedCount > 0 {
		fmt.Printf("Removed %d confirmed transactions from mempool\n", removedCount)
	}
}

// CleanExpiredTransactions removes expired transactions from mempool
func (mm *MempoolManager) CleanExpiredTransactions() {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	now := time.Now()
	expiredTxs := make([]string, 0)

	for txID, mempoolTx := range mm.transactions {
		if now.Sub(mempoolTx.Timestamp) > mm.timeout {
			expiredTxs = append(expiredTxs, txID)
		}
	}

	// Remove expired transactions
	for _, txID := range expiredTxs {
		delete(mm.transactions, txID)
	}

	if len(expiredTxs) > 0 {
		fmt.Printf("Cleaned %d expired transactions from mempool\n", len(expiredTxs))
	}
}

// GetMempoolInfo returns information about the mempool
func (mm *MempoolManager) GetMempoolInfo() MempoolInfo {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	totalSize := 0
	totalFees := int64(0)

	for _, mempoolTx := range mm.transactions {
		totalSize += mempoolTx.Size
		totalFees += mempoolTx.Fees
	}

	return MempoolInfo{
		TransactionCount: len(mm.transactions),
		TotalSize:        totalSize,
		TotalFees:        totalFees,
		MaxSize:          mm.maxSize,
		Timeout:          mm.timeout,
	}
}

// MempoolInfo contains mempool statistics
type MempoolInfo struct {
	TransactionCount int
	TotalSize        int
	TotalFees        int64
	MaxSize          int
	Timeout          time.Duration
}

// StartCleanupRoutine starts a routine to periodically clean expired transactions
func (mm *MempoolManager) StartCleanupRoutine() {
	ticker := time.NewTicker(10 * time.Minute) // Clean every 10 minutes

	go func() {
		for range ticker.C {
			mm.CleanExpiredTransactions()
		}
	}()

	fmt.Println("Mempool cleanup routine started")
}

// broadcastTransaction broadcasts a transaction to all connected peers
func (mm *MempoolManager) broadcastTransaction(tx TransactionInterface) {
	txID := tx.GetID()
	invData := InvData{
		AddrFrom: mm.server.Address,
		Type:     "tx",
		Items:    [][]byte{txID},
	}

	msg := Message{
		Command: CmdInv,
		Data:    GobEncode(invData),
	}

	// Broadcast to all known nodes
	mm.server.BroadcastMessage(msg)

	fmt.Printf("Broadcasted transaction %x to network\n", txID)
}

// calculateFees calculates transaction fees (simplified)
func (mm *MempoolManager) calculateFees(tx TransactionInterface) int64 {
	// Simplified fee calculation based on transaction size
	size := len(tx.Serialize())
	return int64(size) * 10 // 10 satoshis per byte
}

// evictOldestTransaction removes the oldest transaction to make space
func (mm *MempoolManager) evictOldestTransaction() {
	var oldestTxID string
	var oldestTime time.Time

	for txID, mempoolTx := range mm.transactions {
		if oldestTxID == "" || mempoolTx.Timestamp.Before(oldestTime) {
			oldestTxID = txID
			oldestTime = mempoolTx.Timestamp
		}
	}

	if oldestTxID != "" {
		delete(mm.transactions, oldestTxID)
		fmt.Printf("Evicted oldest transaction %s from mempool\n", oldestTxID[:8])
	}
}

// ValidateTransaction validates a transaction before adding to mempool
func (mm *MempoolManager) ValidateTransaction(tx TransactionInterface) error {
	// Simplified validation - in reality would check:
	// 1. Transaction format and structure
	// 2. Digital signatures
	// 3. Input/output validity
	// 4. Double-spending prevention
	// 5. Fee adequacy
	// 6. Script execution

	txID := tx.GetID()
	if len(txID) == 0 {
		return fmt.Errorf("invalid transaction ID")
	}

	serialized := tx.Serialize()
	if len(serialized) == 0 {
		return fmt.Errorf("transaction serialization failed")
	}

	return nil
}

// HasTransaction checks if a transaction exists in the mempool
func (mm *MempoolManager) HasTransaction(txID []byte) bool {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	txIDStr := fmt.Sprintf("%x", txID)
	_, exists := mm.transactions[txIDStr]
	return exists
}

// Size returns the number of transactions in the mempool
func (mm *MempoolManager) Size() int {
	mm.mu.RLock()
	defer mm.mu.RUnlock()
	return len(mm.transactions)
}

// Clear removes all transactions from the mempool
func (mm *MempoolManager) Clear() {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	count := len(mm.transactions)
	mm.transactions = make(map[string]MempoolTransaction)

	fmt.Printf("Cleared %d transactions from mempool\n", count)
}
