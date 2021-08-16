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

package backend

import (
	"bytes"
	"encoding/json"
	"errors"
	"sync"
	"time"
	"xfsgo"
	"xfsgo/common"
	"xfsgo/p2p"
	"xfsgo/p2p/discover"

	"github.com/sirupsen/logrus"
)

var MaxHashFetch = uint64(512)

type hashPack struct {
	peerId discover.NodeId
	hashes remoteHashes
}
type blockPack struct {
	peerId discover.NodeId
	blocks remoteBlocks
}

type txPack struct {
	peerId discover.NodeId
	txs    remoteTxs
}

type handler struct {
	newPeerCh       chan *peer
	hashPackCh      chan hashPack
	blockPackCh     chan blockPack
	txPackCh        chan txPack
	peers           map[discover.NodeId]*peer
	blockchain      *xfsgo.BlockChain
	syncLock        sync.Mutex
	fetchHashesLock sync.Mutex
	fetchBlocksLock sync.Mutex
	processLock     sync.Mutex
	version         uint32
	network         uint32
	txPool          *xfsgo.TxPool
	eventBus        *xfsgo.EventBus
}

func newHandler(bc *xfsgo.BlockChain, pv uint32, nv uint32, eventBus *xfsgo.EventBus, txPool *xfsgo.TxPool) (*handler, error) {
	h := &handler{
		newPeerCh:   make(chan *peer, 1),
		hashPackCh:  make(chan hashPack),
		blockPackCh: make(chan blockPack),
		txPackCh:    make(chan txPack),
		peers:       make(map[discover.NodeId]*peer),
		blockchain:  bc,
		version:     pv,
		network:     nv,
		eventBus:    eventBus,
		txPool:      txPool,
	}
	return h, nil
}

func (h *handler) handleNewPeer(p2p p2p.Peer) error {
	p := newPeer(p2p, h.version, h.network)
	h.newPeerCh <- p
	return h.handle(p)
}

func (h *handler) handle(p *peer) error {
	var err error = nil
	head := h.blockchain.CurrentBlock()
	if err = p.Handshake(head.Hash(), head.Height()); err != nil {
		return err
	}
	logrus.Infof("handshake success, peer.height: %d, p.head: %s  p.id %v\n", p.height, p.head.Hex(), p.p2pPeer.ID())
	p2pPeer := p.p2p()
	id := p2pPeer.ID()
	h.peers[id] = p
	logrus.Infof("peers len: %v\n", len(h.peers))
	defer delete(h.peers, id)
	// Send local transaction to remote synchronization
	h.syncTransactions(p)
out:
	for {
		select {
		// Node exit channel
		case <-p2pPeer.QuitCh():
			break out
		default:
		}
		if err = h.handleMsg(p); err != nil {
			return err
		}
	}
	return nil
}

func (h *handler) handleMsg(p *peer) error {
	msg := <-p.p2pPeer.GetProtocolMsgCh()
	msgCode := msg.Type()
	bodyBs, err := msg.ReadAll()
	// logrus.Printf(" bodybs:%v type:%v\n", string(bodyBs), msgCode)
	if err != nil {
		logrus.Printf("handle message err %s", err)
		return err
	}

	switch msgCode {
	case GetBlockHashesFromNumberMsg:
		// Get local block Hash list
		var data *getBlockHashesFromNumberData = nil
		if err := json.Unmarshal(bodyBs, &data); err != nil {
			logrus.Warnf("handle GetBlockHashesFromNumberMsg msg err: %s", err)
			return err
		}
		hashes := h.blockchain.GetBlockHashes(data.From, data.Count)
		// Send local hash value
		logrus.Infof("berthashes %v\n", hashes)
		if err := p.SendBlockHashes(hashes); err != nil {
			logrus.Warnf("send block hashes data err: %s", err)
			return err
		}
	case BlockHashesMsg:
		// Accept block Hash list message
		var data []common.Hash = nil
		if err := json.Unmarshal(bodyBs, &data); err != nil {
			logrus.Warnf("handle BlockHashesMsg msg err: %s", err)
			return err
		}
		h.hashPackCh <- hashPack{
			peerId: p.p2p().ID(),
			hashes: data,
		}
	case GetBlocksMsg:
		// Process get block list request
		var data []common.Hash = nil
		if err := json.Unmarshal(bodyBs, &data); err != nil {
			logrus.Warnf("handle GetBlocksMsg msg err: %s", err)
			return err
		}
		blocks := make([]*xfsgo.Block, 0)
		for _, hash := range data {
			block := h.blockchain.GetBlockByHash(hash)
			if block == nil {
				break
			}
			blocks = append(blocks, block)
		}
		if err := p.SendBlocks(blocks); err != nil {
			logrus.Warnf("send blocks data err: %s", err)
			return err
		}
	case BlocksMsg: // 接受区块列表消息
		// Accept block list message
		var data remoteBlocks = nil
		if err := json.Unmarshal(bodyBs, &data); err != nil {
			logrus.Warnf("handle BlocksMsg msg err: %s", err)
			return err
		}
		h.blockPackCh <- blockPack{
			peerId: p.p2p().ID(),
			blocks: data,
		}
	case NewBlockMsg: // 处理区块广播
		// Processing block broadcasting
		logrus.Printf(" bodybs:%v type:%v\n", string(bodyBs), msgCode)
		var data *xfsgo.Block = nil
		if err := json.Unmarshal(bodyBs, &data); err != nil {
			logrus.Warnf("handle NewBlockMsg err: %s", err)
			return err
		}
		p.height = data.Height()
		p.head = data.Hash()
		go h.synchronise(p)
	case TxMsg: // 处理交易广播
		// Process transaction broadcast
		var txs remoteTxs = nil
		if err := json.Unmarshal(bodyBs, &txs); err != nil {
			logrus.Warnf("handle TxMsg msg err: %s", err)
			return err
		}
		for _, tx := range txs {
			if err := h.txPool.Add(tx); err != nil {
				logrus.Warnf("handle TxMsg msg err: %s", err)
			}
		}
	}
	return nil
}

func (h *handler) syncer() {
	forceSync := time.Tick(10 * time.Second)
	for {
		select {
		case <-h.newPeerCh:
			if len(h.peers) < 5 {
				break
			}
			go h.synchronise(h.basePeer())
		case <-forceSync:
			// synchronise block
			go h.synchronise(h.basePeer())
		}
	}
}

func (h *handler) basePeer() *peer {
	head := h.blockchain.CurrentBlock()
	var (
		bestPeer   *peer  = nil
		baseHeight uint64 = head.Height()
	)
	for _, v := range h.peers {
		if ph := v.height; ph > baseHeight {
			bestPeer = v
			baseHeight = ph
		}
	}
	return bestPeer
}

//Node synchronization
func (h *handler) synchronise(p *peer) {
	if p == nil {
		return
	}
	h.syncLock.Lock()
	h.eventBus.Publish(xfsgo.SyncStartEvent{})
	defer func() {
		h.eventBus.Publish(xfsgo.SyncDoneEvent{})
		h.syncLock.Unlock()
	}()

	logrus.Warnf("Synchronizing, peerAddress: %s", p.p2pPeer.ID())
	var number uint64
	var err error
	if number, err = h.findAncestor(p); err != nil {
		logrus.Infof("findAncestor errs %v\n", err.Error())
		return
	}
	logrus.Infof("Get public block height: %d", number)
	go func() {
		if err = h.fetchHashes(p, number+1); err != nil {
			logrus.Warn("fetch hashes err")
		}
	}()
	go func() {
		if err = h.fetchBlocks(p); err != nil {
			logrus.Warn("fetch blocks err")
		}
	}()
}

// Find common block height
func (h *handler) findAncestor(p *peer) (uint64, error) {
	var err error = nil
	headBlock := h.blockchain.CurrentBlock()
	if headBlock == nil {
		return 0, errors.New("empty")
	}

	// get header block height
	height := headBlock.Height()
	var from uint64
	froms := int(height) - int(MaxHashFetch)
	if froms < int(0) {
		from = uint64(0)
	} else {
		from = uint64(froms)
	}

	logrus.Infof("Find a fixed height range: [%d, %d]", from, MaxHashFetch)
	// Get block Hash list

	if err = p.RequestHashesFromNumber(from, MaxHashFetch); err != nil {
		return 0, err
	}
	number := uint64(0)
	haveHash := common.ZeroHash
	// Blocking receive pack messages
loop:
	for {
		select {
		// Skip loop if timeout
		case <-time.After(3 * 60 * time.Second):
			return 0, errors.New("find hashes time out err1")
		case pack := <-h.hashPackCh:
			wanId := p.p2p().ID()
			wantPeerId := wanId[:]

			gotPeerId := pack.peerId[:]
			if !bytes.Equal(wantPeerId, gotPeerId) {
				break
			}

			hashes := pack.hashes
			if len(hashes) == 0 {
				return 0, errors.New("empty hashes")
			}
			for i, hash := range hashes {
				if h.hashBlock(hash) {
					continue
				}
				// Record height and hash value
				number = from + uint64(i)
				haveHash = hash
				break loop
			}
		}
	}
	if bytes.Equal(common.ZeroHash.Bytes(), haveHash.Bytes()) {
		return number, nil
	}
	logrus.Infof("The fixed interval value is not found. Continue to traverse and find...")
	// If no fixed interval value is found, traverse all blocks and binary search
	left := 0
	right := int(MaxHashFetch) + 1
	for left < right {
		logrus.Infof("Traversing height range: [%d, %d]", left, right)
		mid := (left + right) / 2
		if err = p.RequestHashesFromNumber(uint64(mid), 1); err != nil {
			return 0, err
		}
		for {
			select {
			case <-time.After(3 * 60 * time.Second):
				return 0, errors.New("find hashes time out err2")
			case pack := <-h.hashPackCh:
				wanId := p.p2p().ID()
				wantPeerId := wanId[:]
				gotPeerId := pack.peerId[:]
				// if bytes.Compare(wantPeerId, gotPeerId) == common.Zero {
				// 	break
				// }
				if !bytes.Equal(wantPeerId, gotPeerId) {
					break
				}
				hashes := pack.hashes
				if len(hashes) != 1 {
					return 0, nil
				}
				if h.hashBlock(hashes[0]) {
					left = mid + 1
				} else {
					right = mid
				}
			}
		}
	}
	return uint64(left) - 1, nil
}

// Find out whether the hash value exists locally in the local block list
func (h *handler) hashBlock(hash common.Hash) bool {
	if has := h.blockchain.GetBlockByHash(hash); has == nil {
		return false
	}
	return true
}

func (h *handler) fetchHashes(p *peer, from uint64) error {
	h.fetchHashesLock.Lock()
	defer h.fetchHashesLock.Unlock()
	go func() {
		if err := p.RequestHashesFromNumber(from, MaxHashFetch); err != nil {
			logrus.Warn("request hashes err")
		}
	}()
	for {
		select {
		case <-time.After(3 * 60 * time.Second):
			return errors.New("fetchHashes time out err")
		case pack := <-h.hashPackCh:
			wanId := p.p2p().ID()
			wantPeerId := wanId[:]
			gotPeerId := pack.peerId[:]
			if !bytes.Equal(wantPeerId, gotPeerId) {
				break
			}
			hashes := pack.hashes
			if len(hashes) == 0 {
				return nil
			}
			for _, hash := range hashes {
				logrus.Infof("handle fetch ahash: %s", hash.Hex())
			}
			if err := p.RequestBlocks(hashes); err != nil {
				return err
			}
		}
	}
}

func (h *handler) fetchBlocks(p *peer) error {
	h.fetchBlocksLock.Lock()
	defer h.fetchBlocksLock.Unlock()
	for {
		select {
		case <-time.After(3 * 60 * time.Second):
			return errors.New("fetchHashes time out err")
		case pack := <-h.blockPackCh:
			wanId := p.p2p().ID()
			wantPeerId := wanId[:]
			gotPeerId := pack.peerId[:]
			if !bytes.Equal(wantPeerId, gotPeerId) {
				break
			}
			blocks := pack.blocks
			if len(blocks) == 0 {
				return nil
			}
			go h.process(blocks)
		}
	}
}

func (h *handler) process(blocks remoteBlocks) {
	h.processLock.Lock()
	defer h.processLock.Unlock()
	for _, block := range blocks {
		if err := h.blockchain.InsertChain(block); err != nil {
			logrus.Printf("InsertChain err %v\n", err.Error())
			continue
		}
	}
}

func (h *handler) BroadcastBlock(block *xfsgo.Block) {
	for k := range h.peers {
		p := h.peers[k]
		if err := p.SendNewBlock(block); err != nil {
			logrus.Infof("peers SendNewBlock err: %v\n", err.Error())
			continue
		}
	}
}

func (h *handler) BroadcastTx(txs remoteTxs) {
	for k := range h.peers {
		p := h.peers[k]
		if err := p.SendTransactions(txs); err != nil {
			continue
		}
	}
}

// 交易广播
func (h *handler) txBroadcastLoop() {
	txPreEventSub := h.eventBus.Subscript(xfsgo.TxPreEvent{})
	defer txPreEventSub.Unsubscribe()
	for {
		select {
		case e := <-txPreEventSub.Chan():
			event := e.(xfsgo.TxPreEvent)
			tx := event.Tx
			h.BroadcastTx([]*xfsgo.Transaction{tx})
		}
	}
}

// 区块广播
func (h *handler) minedBroadcastLoop() {
	newMinerBlockEventSub := h.eventBus.Subscript(xfsgo.NewMinedBlockEvent{})
	logrus.Println("minedBroadcastLoop 1212")
	defer newMinerBlockEventSub.Unsubscribe()
	for {
		select {
		case e := <-newMinerBlockEventSub.Chan():
			event := e.(xfsgo.NewMinedBlockEvent)
			block := event.Block
			logrus.Println("newMinerBlockEventSub")
			h.BroadcastBlock(block)
		}
	}

}

func (h *handler) syncTransactions(p *peer) {
	txs := h.txPool.GetTransactions()
	if len(txs) == 0 {
		return
	}
	h.txPackCh <- txPack{
		peerId: p.p2p().ID(),
		txs:    txs,
	}
}

func (h *handler) txSyncLoop() {
	send := func(pack txPack) {
		peerId := pack.peerId
		if p, has := h.peers[peerId]; has {
			if err := p.SendTransactions(pack.txs); err != nil {
				logrus.Warnf("send txs err: %s", err)
			}
		}
	}
	for {
		select {
		case pack := <-h.txPackCh:
			send(pack)
		}
	}
}

func (h *handler) Start() {
	// start broadcasing transaction
	go h.txBroadcastLoop()
	// start broadcasing block
	go h.minedBroadcastLoop()
	// start synchronising block
	go h.syncer()
	// start synchronising transaction
	go h.txSyncLoop()
}
