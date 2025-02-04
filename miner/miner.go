// Copyright 2018 The xfsgo Authors
// This file is part of the xfsgo library.
//
// The xfsgo library is free software: you can redistribute it and/or modify
// it under the terms of the MIT Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The xfsgo library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// MIT Lesser General Public License for more details.
//
// You should have received a copy of the MIT Lesser General Public License
// along with the xfsgo library. If not, see <https://mit-license.org/>.

package miner

import (
	"bytes"
	"errors"
	"fmt"
	"runtime"
	"sync"
	"time"
	"xfsgo"
	"xfsgo/common"
	"xfsgo/storage/badger"

	"github.com/sirupsen/logrus"
)

const hashUpdateSecs = 1

var defaultNumWorkers = uint32(runtime.NumCPU())

//var defaultNumWorkers = uint32(1)

var maxNonce = ^uint64(0)

type Config struct {
	Coinbase common.Address
}

// Miner creates blocks with transactions in tx pool and searches for proof-of-work values.
type Miner struct {
	*Config
	mu sync.Mutex
	sync.Mutex
	started          bool
	quit             chan struct{}
	runningWorkers   []chan struct{}
	updateNumWorkers chan uint32
	numWorkers       uint32
	eventBus         *xfsgo.EventBus
	canStart         bool
	pool             *xfsgo.TxPool
	chain            *xfsgo.BlockChain
	stateDb          *badger.Storage
}

func NewMiner(config *Config, stateDb *badger.Storage, chain *xfsgo.BlockChain, eventBus *xfsgo.EventBus, pool *xfsgo.TxPool) *Miner {
	m := &Miner{
		Config:           config,
		chain:            chain,
		stateDb:          stateDb,
		quit:             make(chan struct{}),
		numWorkers:       defaultNumWorkers,
		updateNumWorkers: make(chan uint32),
		pool:             pool,
		canStart:         true,
		started:          false,
		eventBus:         eventBus,
	}
	go m.mainLoop()
	return m
}

// Start starts up xfs mining
func (m *Miner) Start() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.started || !m.canStart {
		return
	}
	go m.miningWorkerController()
	m.started = true
}

func (m *Miner) GetNumWorkers() uint32 {
	return m.numWorkers
}

func (m *Miner) SetNumWorkers(num uint32) {
	m.updateNumWorkers <- num
}

// mainLoop is the miner's main event loop, waiting for and reacting to synchronize events.
func (m *Miner) mainLoop() {
	startSub := m.eventBus.Subscript(xfsgo.SyncStartEvent{})
	doneSub := m.eventBus.Subscript(xfsgo.SyncDoneEvent{})
	defer func() {
		startSub.Unsubscribe()
		doneSub.Unsubscribe()
	}()
out:
	for {
		select {
		case <-startSub.Chan():
			m.mu.Lock()
			m.canStart = false
			m.mu.Unlock()
		case <-doneSub.Chan():
			m.mu.Lock()
			m.canStart = true
			m.mu.Unlock()
			break out
		}
	}
}
func (m *Miner) mimeBlockWithParent(
	stateTree *xfsgo.StateTree,
	parentBlock *xfsgo.Block,
	coinbase common.Address,
	txs []*xfsgo.Transaction,
	quit chan struct{},
	ticker *time.Ticker) (*xfsgo.Block, error) {
	if parentBlock == nil {
		return nil, errors.New("parentBlock is nil")
	}
	//create a Blockheader which will be the header of the new block.
	lastGenerated := time.Now().Unix()
	header := &xfsgo.BlockHeader{
		Height:        parentBlock.Height() + 1,
		HashPrevBlock: parentBlock.Hash(),
		Timestamp:     uint64(lastGenerated),
		Coinbase:      coinbase,
	}
	//calculate the next difficuty for hash value of next block.
	var err error
	header.Bits, err = m.chain.CalcNextRequiredDifficulty()
	if err != nil {
		return nil, err
	}
	logrus.Debugf("block height: %d, difficuty: %d, timestamp: %d", header.Height, header.Bits, header.Timestamp)
	//process the transations
	res, err := m.chain.ApplyTransactions(stateTree, txs)
	if err != nil {
		return nil, fmt.Errorf("apply trasactions err")
	}
	//calculate the rewards
	xfsgo.AccumulateRewards(stateTree, header)
	stateTree.UpdateAll()
	stateRootBytes := stateTree.Root()
	stateRootHash := common.Bytes2Hash(stateRootBytes)
	header.StateRoot = stateRootHash
	//create a new block and execite the consensus algorithms
	perBlock := xfsgo.NewBlock(header, txs, res)
	return m.execPow(perBlock, quit, ticker)
}

// run the consensus algorithms
func (m *Miner) execPow(perBlock *xfsgo.Block, quit chan struct{}, ticker *time.Ticker) (*xfsgo.Block, error) {
	targetDifficulty := xfsgo.BitsUnzip(perBlock.Bits())
	target := targetDifficulty.Bytes()
	targetHash := make([]byte, 32)
	copy(targetHash[32-len(target):], target)
	logrus.Debugf("running the POW consensusm，block height: %d, block difficulty: %d, target: 0x%x", perBlock.Height(), perBlock.Bits(), targetHash)

out:
	for nonce := uint64(0); nonce <= maxNonce; nonce++ {
		select {
		case <-quit:
			break out
		case <-ticker.C:
			lashBlock := m.chain.CurrentBlock()
			lastHeight := lashBlock.Height()
			currentBlockHeight := perBlock.Height()
			//exit this loop if the current height is updated and larger than the height of the blockchain.
			if lastHeight >= currentBlockHeight {
				logrus.Debugf("current height of blockchain has been updated: %d, current height: %d", lastHeight, currentBlockHeight)
				break out
			}
		default:
		}
		hash := perBlock.UpdateNonce(nonce)
		if bytes.Compare(hash.Bytes(), targetHash) <= 0 {
			lashBlock := m.chain.CurrentBlock()
			lastHeight := lashBlock.Height()
			currentBlockHeight := perBlock.Height()
			if lastHeight >= currentBlockHeight {
				break out
			}
			return perBlock, nil
		}
	}
	return nil, fmt.Errorf("not")
}

func (m *Miner) generateBlocks(num uint32, quit chan struct{}) {
	ticker := time.NewTicker(time.Second * hashUpdateSecs)
	defer ticker.Stop()
out:
	for {
		select {
		case <-quit:
			break out
		default:
		}
		txs := m.pool.GetTransactions()
		logrus.Debugf("woker-%d obtaining the transactions queue", num)
		logrus.Debugf("woker-%d has obtained transaction counts: %d", num, len(txs))
		lastBlock := m.chain.CurrentBlock()
		lastStateRoot := lastBlock.StateRoot()
		logrus.Debugf("woker-%d the newest block: %s, height: %d", num, lastBlock.HashHex(), lastBlock.Height())
		stateTree := xfsgo.NewStateTree(m.stateDb, lastStateRoot.Bytes())
		block, err := m.mimeBlockWithParent(stateTree, lastBlock, m.Coinbase, txs, quit, ticker)
		if err != nil {
			continue out
		}
		if err = stateTree.Commit(); err != nil {
			continue out
		}
		if err = m.chain.WriteBlock(block); err != nil {
			continue out
		}
		hash := block.Hash()
		sr := block.StateRoot()
		logrus.Infof("woker-%d, the block has packed successfully, height: %d, hash: %s, stateRoot: %s", num, block.Height(), hash.Hex(), sr.Hex())
		st := xfsgo.NewStateTree(m.stateDb, sr.Bytes())
		balance := st.GetBalance(m.Coinbase)
		logrus.Infof("current coinbase: %s, balance: %d", m.Coinbase.B58String(), balance)
		m.eventBus.Publish(xfsgo.NewMinedBlockEvent{Block: block})
	}
}
func closeWorkers(cs []chan struct{}) {
	for _, c := range cs {
		close(c)
	}
}
func (m *Miner) miningWorkerController() {
	var runningWorkers []chan struct{}
	launchWorkers := func(numWorkers uint32) {
		for i := uint32(0); i < numWorkers; i++ {
			quit := make(chan struct{})
			runningWorkers = append(runningWorkers, quit)
			logrus.Infof("woker-%d started", i)
			go m.generateBlocks(i, quit)
		}
	}
	runningWorkers = make([]chan struct{}, 0)
	logrus.Debugf("starting up workers, workers starting number: %d", m.numWorkers)
	launchWorkers(m.numWorkers)
	txPreEventSub := m.eventBus.Subscript(xfsgo.TxPreEvent{})
	defer txPreEventSub.Unsubscribe()
out:
	for {
		select {
		case <-m.quit:
			logrus.Info("miner quit")
			closeWorkers(runningWorkers)
			m.reset()
			break out
		case e := <-txPreEventSub.Chan():
			_ = e
		case workers := <-m.updateNumWorkers:
			if m.numWorkers != workers {
				m.numWorkers = workers
				m.Stop()
				if !m.started {
					m.Start()
				}
			}
		default:
		}
	}
}

func (m *Miner) Stop() {
	if m.started {
		m.started = false
		close(m.quit)
	}
}

func (m *Miner) reset() {
	m.started = false
	m.canStart = true
	m.quit = make(chan struct{})
}
