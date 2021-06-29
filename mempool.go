package xblockchain

import (
	"fmt"
	"sync"
	"xblockchain/uint256"
)

type TxPool struct {
	eventBus *EventBus
	blockchain *BlockChain
	currentStateFn func()
	pending map[uint256.UInt256] *Transaction
	chainHeadEventSub *Subscription
	mu sync.RWMutex
}

type PendingTX struct {
	nonce int
	tx *Transaction
}

func NewTxPool(eventBus *EventBus,chain *BlockChain) *TxPool {
	txPool := &TxPool{
		eventBus: eventBus,
		blockchain: chain,
		pending: make(map[uint256.UInt256] *Transaction),
	}
	txPool.chainHeadEventSub = eventBus.Subscript(ChainHeadEvent{})
	go txPool.eventLoop()
	return txPool
}

func (pool *TxPool) AddTx(tx *Transaction) error {
	pool.mu.Lock()
	defer pool.mu.Unlock()
	hash := *uint256.NewUInt256BS(tx.Hash())
	if pool.pending[hash] != nil {
		return fmt.Errorf("know transaction (%s)", hash.Hex())
	}
	if pool.blockchain.VerifyTransaction(tx) {
		return fmt.Errorf("verify transaction err, hash: %s\n", hash.Hex())
	}
	pool.pending[hash] = tx
	pool.eventBus.Publish(TxPreEvent{tx})
	return nil
}

func (pool *TxPool) eventLoop() {
	for {
		select {
		case e := <-pool.chainHeadEventSub.Chan():
			event := e.(ChainHeadEvent)
			block := event.Block
			txs := block.Transactions
			for _,tx := range txs {
				if hash, have := pool.pending[*tx.ID]; have {
					delete(pool.pending, *hash.ID)
				}
			}
		}
	}
}


func (pool *TxPool) GetTransactions() []*Transaction {
	pool.mu.RLock()
	defer pool.mu.RUnlock()
	txs := make([]*Transaction, 0)
	for _, v := range pool.pending {
		txs = append(txs, v)
	}
	return txs
}


