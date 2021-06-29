package xblockchain

import (
	"fmt"
	"log"

	"github.com/sirupsen/logrus"
)

var MaxBlockTxCount = 10

type Miner struct {
	blockChain    *BlockChain
	wallets       *Wallets
	running       bool
	poll          *TxPool
	eventBus      *EventBus
	TxPendingPool *TXPendingPool
	txPreEventSub *Subscription
	txQueue       []*Transaction
}

func NewMiner(
	blockChain *BlockChain,
	wallets *Wallets,
	pool *TxPool,
	eventBus *EventBus) *Miner {
	m := &Miner{
		blockChain: blockChain,
		wallets:    wallets,
		poll:       pool,
		eventBus:   eventBus,
		txQueue:    make([]*Transaction, 0),
	}
	m.txPreEventSub = eventBus.Subscript(TxPreEvent{})
	return m
}

func (m *Miner) Run() error {
	if err := m.PreRun(); err != nil {
		return err
	}
	if m.running {
		return fmt.Errorf("miner is running, unable to start again")
	}
	m.running = true
	go m.run()
	return nil
}
func (m *Miner) commitTransactions() {

}
func (m *Miner) run() {
	go func() {
		for m.running {
			select {
			case _ = <-m.txPreEventSub.Chan():
				txs := m.poll.GetTransactions()
				m.txQueue = append(m.txQueue, txs...)
			}
		}
	}()
	for m.running {
		var defaultAddr string
		if defaultAddr = m.wallets.GetDefault(); defaultAddr == "" {
			logrus.Warn("not found default address !")
		}
		txs := make([]*Transaction, 0)
		coinbaseTx, err := NewCoinBaseTransaction(defaultAddr, "")
		if err != nil {
			logrus.Warnf("create coinbase transaction err: %s", err)
		}
		txs = append(txs, coinbaseTx)
		maxCount := MaxBlockTxCount - 1
		if len(m.txQueue) > 0 && len(m.txQueue) < maxCount {
			txs = append(txs, m.txQueue[:maxCount]...)
			m.txQueue = m.txQueue[maxCount:]
		} else if len(m.txQueue) == 0 {
			txs = append(txs, m.txQueue[:maxCount]...)
		}
		block, err := m.blockChain.AddBlock(txs)
		if err != nil {
			logrus.Warnf("Miner block errors: %s", err)
		}
		m.eventBus.Publish(NewMinerBlockEvent{block})
		log.Printf("Miner tx block success, height: %d, hash: %s", block.Height, block.Hash.Hexstr(true))
	}
}

func (m *Miner) PreRun() error {
	addr := m.wallets.GetDefault()
	if addr == "" {
		return fmt.Errorf("not found default address")
	}
	return nil
}
func (m *Miner) Stop() error {
	if !m.running {
		return fmt.Errorf("miner is not running, unable to stop miner")
	}
	m.running = false
	return nil
}
