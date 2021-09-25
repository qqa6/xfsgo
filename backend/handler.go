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
	"github.com/sirupsen/logrus"
	"sync"
	"time"
	"xfsgo"
	"xfsgo/common"
	"xfsgo/p2p"
	"xfsgo/p2p/discover"
)

var MaxHashFetch = uint64(512)
var timeoutTTL = 3 * 60 * time.Second

var (
	errPeerClosed = errors.New("peer closed")
	errTimeout = errors.New("timeout")
	errEmptyHashes = errors.New("empty hashes")
)

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
		hashPackCh:  make(chan hashPack, 1),
		blockPackCh: make(chan blockPack, 1),
		txPackCh:    make(chan txPack, 1),
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
	pHeight := p.Height()
	pHead := p.Head()
	logrus.Debugf("Handshake success height=%d, head=%s, id=%v", pHeight, pHead.Hex(), p.p2pPeer.ID())
	p2pPeer := p.p2p()
	id := p2pPeer.ID()
	h.peers[id] = p
	defer delete(h.peers, id)
	// Send local transaction to remote synchronization
	h.syncTransactions(p)
out:
	for {
		select {
		// Node exit channel
		case <-p2pPeer.CloseCh():
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
	peerId := p.p2p().ID()
	select {
	case <-p.p2pPeer.CloseCh():
		return nil
	case msg := <-p.p2pPeer.GetProtocolMsgCh():
		msgCode := msg.Type()
		bodyBs, err := msg.ReadAll()
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
			logrus.Debugf("Handle get block hashes request: from=%d, count=%d, peerId=%x...%x",
				data.From, data.Count, peerId[:4], peerId[len(peerId)-4:])
			hashes := h.blockchain.GetBlockHashes(data.From, data.Count)
			//jsonData,_ := json.Marshal(hashes)
			logrus.Debugf("Send block hashes: dataCout=%d, requestStart=%d, requestCount=%d, peerId=%x...%x",
				len(hashes), data.From, data.Count, peerId[:4], peerId[len(peerId)-4:])
			if err := p.SendBlockHashes(hashes); err != nil {
				logrus.Warnf("Send block hashes data err: %s", err)
				return err
			}
		case BlockHashesMsg:
			// Accept block Hash list message
			var data []common.Hash = nil
			if err := json.Unmarshal(bodyBs, &data); err != nil {
				logrus.Warnf("handle BlockHashesMsg msg err: %s", err)
				return err
			}
			//logrus.Infof("Handle Peer BlockHashesMsg: count=%d, peerId=%x...%x",
			//	len(data), peerId[:4], peerId[len(peerId)-4:])
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
			//logrus.Infof("Handle Peer GetBlocksMsg: hashCount=%d, peerId=%x...%x",
			//	len(data), peerId[:4], peerId[len(peerId)-4:])
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
		case BlocksMsg: // Accept block list message
			// Accept block list message
			var data remoteBlocks = nil
			if err := json.Unmarshal(bodyBs, &data); err != nil {
				logrus.Warnf("handle BlocksMsg msg err: %s", err)
				return err
			}
			//logrus.Infof("Handle Peer BlocksMsg: count=%d, peerId=%x...%x",
			//	len(data), peerId[:4], peerId[len(peerId)-4:])
			h.blockPackCh <- blockPack{
				peerId: p.p2p().ID(),
				blocks: data,
			}
		case NewBlockMsg: // Processing block broadcasting
			// Processing block broadcasting
			var data *xfsgo.Block = nil
			if err := json.Unmarshal(bodyBs, &data); err != nil {
				logrus.Warnf("handle NewBlockMsg err: %s", err)
				return err
			}
			blockHash := data.Hash()
			blockHeight := data.Height()
			//logrus.Infof("Handle Peer NewBlockMsg: height=%d, hash=%x...%x, peerId=%x...%x",
			//	data.Height(), blockHash[:4], blockHash[len(blockHash)-4:], peerId[:4], peerId[len(peerId)-4:])
			pHead := p.Head()
			logrus.Debugf("Successfully update peer: height=%d, hash=%x...%x, peerId=%x...%x",
				p.Height(), pHead[:4], pHead[len(pHead)-4:], peerId[:4], peerId[len(peerId)-4:])
			if blockHeight > p.Height() {
				p.SetHeight(data.Height())
				p.SetHead(blockHash)

				go h.synchronise(p)
			}
			//go h.lessPeer(p)

		case TxMsg: // Process transaction broadcast
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
	default:
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
			//logrus.Debugf("Pick base peer force sync")
			// synchronise block
			go h.synchronise(h.basePeer())
		}
	}
}

func (h *handler) basePeer() *peer {
	head := h.blockchain.CurrentBlock()
	var (
		base   *peer  = nil
		baseHeight = head.Height()
	)
	for _, v := range h.peers {
		if ph := v.Height(); ph > baseHeight {
			base = v
			baseHeight = ph
		}
	}
	//if base != nil {
	//	baseId := base.p2p().ID()
	//	logrus.Debugf("Successfully Pick base peer: id: %x...%x", baseId[:4], baseId[len(baseId)-4:])
	//}
	return base
}

//Node synchronization
func (h *handler) synchronise(p *peer) {
	if p == nil {
		return
	}
	pId := p.p2p().ID()
	logrus.Debugf("Synchronise from peer: id=%x...%x", pId[:4], pId[len(pId)-4:])
	h.syncLock.Lock()
	defer func() {
		h.eventBus.Publish(xfsgo.SyncDoneEvent{})
		h.syncLock.Unlock()
	}()
	h.eventBus.Publish(xfsgo.SyncStartEvent{})
	var number uint64
	var err error
	if number, err = h.findAncestor(p); err != nil {
		return
	}
	go func() {
		if err = h.fetchHashes(p, number+1); err != nil {
			logrus.Warnf("Fetch hashes err: %s", err)
		}
	}()
	go func() {
		if err = h.fetchBlocks(p); err != nil {
			logrus.Warnf("Fetch Blocks err: %s", err)
		}
	}()
}

// Find common block height
func (h *handler) findAncestor(p *peer) (uint64, error) {
	pid := p.p2p().ID()
	var err error = nil
	headBlock := h.blockchain.CurrentBlock()
	if headBlock == nil {
		return 0, errors.New("empty")
	}
	height := headBlock.Height()
	var from = 0
	from = int(height) - int(MaxHashFetch)
	if from < 0 {
		from = 0
	}
	logrus.Debugf("Find ancestor block hashes: chainHeight=%d, start=%d, count=%d, peerId=%x...%x",
		height, from, MaxHashFetch, pid[0:4], pid[len(pid)-4:])
	if err = p.RequestHashesFromNumber(uint64(from), MaxHashFetch); err != nil {
		return 0, err
	}
	number := uint64(0)
	haveHash := common.HashZ
	timeout := time.After(timeoutTTL)
	//finished := false
	//loop:
	for finished := false; !finished; {
		select {
		case <- p.CloseCh():
			logrus.Warnf("Fetch ancestor hashes failed peer closed: chainHeight=%d, from=%d, count: %d,  peerId=%x...%x",
				height, from, MaxHashFetch, pid[0:4], pid[len(pid)-4:])
			return 0, errPeerClosed
		// Skip loop if timeout
		case <-timeout:
			logrus.Warnf("Fetch ancestor hashes timeout: chainHeight=%d, from=%d, count: %d, peerId=%x...%x",
				height,  from, MaxHashFetch, pid[0:4], pid[len(pid)-4:])
			return 0, errTimeout
		case pack := <-h.hashPackCh:
			wanId := p.p2p().ID()
			wantPeerId := wanId[:]
			gotPeerId := pack.peerId[:]
			if !bytes.Equal(wantPeerId, gotPeerId) {
				break
			}
			hashes := pack.hashes
			if len(hashes) == 0 {
				logrus.Warnf("Fetch ancestor hashes is emtpy: chainHeight=%d, from=%d, count: %d, peerId=%x...%x",
					height, from, MaxHashFetch, pid[0:4], pid[len(pid)-4:])
				return 0, errEmptyHashes
			}
			finished = true
			logrus.Debugf("Found ancestor hashes: currentHeight=%d, fetchFrom=%d, fetchCount: %d, foundCount=%d, peerId=%x...%x",
				height, from, MaxHashFetch, len(hashes), pid[:4], pid[len(pid)-4:])
			for i := len(hashes) - 1; i >= 0; i-- {
				hash := hashes[i]
				logrus.Debugf("Check ancestor hashes: chainHeight=%d, fetchFrom=%d, fetchCount: %d, foundCount=%d, index=%d, hash=%x...%x, peerId=%x...%x",
					height, from, MaxHashFetch, len(hashes), i, hash[:4], hash[len(haveHash)-4:], pid[:4], pid[len(pid)-4:])
				if h.hashBlock(hash) {
					number, haveHash = uint64(from) + uint64(i), hashes[i]
					break
				}
			}
		}
	}
	if !bytes.Equal(haveHash[:], common.HashZ[:]) {
		logrus.Debugf("Found ancestor block: height=%d, hash=%x...%x, peerId=%x...%x",
			number, haveHash[:4], haveHash[len(haveHash)-4:], pid[:4], pid[len(pid)-4:])
		return number, nil
	}
	logrus.Warnf("Not found ancestor: currentHeight=%d, from=%d, count=%d, peerId=%x...%x",
		height, from, MaxHashFetch, pid[0:4], pid[len(pid)-4:])
	// If no fixed interval value is found, traverse all blocks and binary search
	//left := 0
	//right := int(MaxHashFetch) + 1
	//for left < right {
	//	logrus.Debugf("Traversing height range: [%d, %d]", left, right)
	//	mid := (left + right) / 2
	//	if err = p.RequestHashesFromNumber(uint64(mid), 1); err != nil {
	//		return 0, err
	//	}
	//	timeout := time.After(timeoutTTL)
	//	for arrived := false; !arrived; {
	//		select {
	//		case <- p.CloseCh():
	//			return 0, errors.New("peer closed")
	//		case <-timeout:
	//			return 0, errors.New("find hashes time out")
	//		case pack := <-h.hashPackCh:
	//			wanId := p.p2p().ID()
	//			wantPeerId := wanId[:]
	//			gotPeerId := pack.peerId[:]
	//			if !bytes.Equal(wantPeerId, gotPeerId) {
	//				break
	//			}
	//			hashes := pack.hashes
	//			if len(hashes) != 1 {
	//				return 0, nil
	//			}
	//			arrived = true
	//			if h.hashBlock(hashes[0]) {
	//				left = mid + 1
	//			} else {
	//				right = mid
	//			}
	//		}
	//	}
	//}
	return 0, nil
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
	pid := p.p2p().ID()
	logrus.Debugf("Fetching Hashes: from=%d, count=%d, peerId=%x...%x", from, MaxHashFetch, pid[0:4], pid[len(pid)-4:])
	timeout := time.NewTimer(0)
	<-timeout.C
	defer timeout.Stop()
	go func() {
		if err := p.RequestHashesFromNumber(from, MaxHashFetch); err != nil {
			logrus.Warnf("Requst fetch hashes from number err: from=%d, count=%d, err=%s, peerId=%x...%x",
				from, MaxHashFetch,err , pid[0:4], pid[len(pid)-4:])
		}
		timeout.Reset(timeoutTTL)
	}()
	//timeout := time.NewTimer(0)
	//timeout := time.After(timeoutTTL)
	for {
		select {
		case <-p.CloseCh():
			logrus.Warnf("Fetch hashes failed peer closed: from=%d, count: %d,  peerId=%x...%x",
				from, MaxHashFetch, pid[0:4], pid[len(pid)-4:])
			return errPeerClosed
		case <-timeout.C:
			logrus.Warnf("Fetch hashes timeout: from=%d, count: %d, peerId=%x...%x",
				from, MaxHashFetch, pid[0:4], pid[len(pid)-4:])
			return errTimeout
		case pack := <-h.hashPackCh:
			wanId := p.p2p().ID()
			wantPeerId := wanId[:]
			gotPeerId := pack.peerId[:]
			if !bytes.Equal(wantPeerId, gotPeerId) {
				break
			}
			timeout.Stop()
			hashes := pack.hashes
			if len(hashes) == 0 {
				logrus.Warnf("Fetch hashes empty: from=%d, count: %d, peerId=%x...%x",
					from, MaxHashFetch, pid[0:4], pid[len(pid)-4:])
				return nil
			}
			logrus.Debugf("Successfully fetched hashes: count=%d, peerId=%x...%x", len(hashes), pid[0:4], pid[len(pid)-4:])
			if err := p.RequestBlocks(hashes); err != nil {
				return err
			}
			return nil
		}
	}
}

func (h *handler) fetchBlocks(p *peer) error {
	h.fetchBlocksLock.Lock()
	defer h.fetchBlocksLock.Unlock()
	for {
		select {
		case <- p.CloseCh():
			return errPeerClosed
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
			logrus.Debugf("Successfully fetched block pack: count=%d, peerId=%x...%x", len(blocks),
				gotPeerId[0:4], gotPeerId[len(gotPeerId)-4:] )
			go h.process(blocks)
		}
	}
}

func (h *handler) process(blocks remoteBlocks) {
	h.processLock.Lock()
	defer h.processLock.Unlock()
	for _, block := range blocks {
		if err := h.blockchain.InsertChain(block); err != nil {
			continue
		}
	}
}

func (h *handler) BroadcastBlock(block *xfsgo.Block) {
	hash := block.Hash()
	logrus.Debugf("Broadcast block height: %d, hash: %x...%x", block.Height(), hash[:4], hash[len(hash)-4:])
	for k := range h.peers {
		p := h.peers[k]
		if err := p.SendNewBlock(block); err != nil {
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

func (h *handler) minedBroadcastLoop() {
	newMinerBlockEventSub := h.eventBus.Subscript(xfsgo.NewMinedBlockEvent{})
	defer newMinerBlockEventSub.Unsubscribe()
	for {
		select {
		case e := <-newMinerBlockEventSub.Chan():
			event := e.(xfsgo.NewMinedBlockEvent)
			block := event.Block
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
