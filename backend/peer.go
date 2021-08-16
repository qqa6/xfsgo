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
	"encoding/json"
	"errors"
	"time"
	"xfsgo"
	"xfsgo/common"
	"xfsgo/p2p"

	"github.com/sirupsen/logrus"
)

type peer struct {
	p2pPeer p2p.Peer
	version uint32
	network uint32
	head    common.Hash
	height  uint64
}

const (
	MsgCodeVersion              uint8 = 5
	GetBlockHashesFromNumberMsg uint8 = 6
	BlockHashesMsg              uint8 = 7
	GetBlocksMsg                uint8 = 8
	BlocksMsg                   uint8 = 9
	NewBlockMsg                 uint8 = 10
	TxMsg                       uint8 = 11
)

func newPeer(p p2p.Peer, version uint32, network uint32) *peer {
	pt := &peer{
		p2pPeer: p,
		version: version,
		network: network,
	}
	return pt
}

func (p *peer) p2p() p2p.Peer {
	return p.p2pPeer
}

type statusData struct {
	Version uint32      `json:"version"`
	Network uint32      `json:"network"`
	Head    common.Hash `json:"head"`
	Height  uint64      `json:"height"`
}

type getBlockHashesFromNumberData struct {
	From  uint64 `json:"from"`
	Count uint64 `json:"count"`
}

type remoteTxs []*xfsgo.Transaction
type remoteHashes []common.Hash
type remoteBlocks []*xfsgo.Block

// Handshake runs the protocol handshake using messages(hash value and height of current block).
// to verifies whether the peer matchs the prptocol that attempts to add the connection as a peer.
func (p *peer) Handshake(head common.Hash, height uint64) error {
	go func() {
		if err := p2p.SendMsgData(p.p2pPeer, MsgCodeVersion, &statusData{
			Version: p.version,
			Network: p.network,
			Head:    head,
			Height:  height,
		}); err != nil {
			return
		}
	}()
	// r := p.p2pPeer.Reader()

	for {
		select {
		case msg := <-p.p2pPeer.GetProtocolMsgCh():
			msgCode := msg.Type()
			switch msgCode {
			case MsgCodeVersion:
				data, _ := msg.ReadAll()
				logrus.Infof("handle message type: %d, data: %s", msgCode, string(data))
				status := statusData{}
				if err := json.Unmarshal(data, &status); err != nil {
					return err
				}
				if status.Version != p.version {
					return errors.New("p2p version not match")
				}
				if status.Network != p.network {
					return errors.New("network id not match")
				}
				p.head = status.Head
				p.height = status.Height
				return nil
			}
		case <-time.After(3 * 60 * time.Second):
			return errors.New("time out")
		}
		// readBufBody := p.p2pPeer.Reader()
		// msg, err := p2p.ReadMessage(readBufBody)

	}
}

// RequestHashesFromNumber fetches a batch of hashes from a peer, starting at from, getting count
func (p *peer) RequestHashesFromNumber(from uint64, count uint64) error {
	logrus.Infof("form:%v count:%v\n", from, count)
	if err := p2p.SendMsgData(p.p2pPeer, GetBlockHashesFromNumberMsg, &getBlockHashesFromNumberData{
		From:  from,
		Count: count,
	}); err != nil {
		return err
	}
	return nil
}

// SendBlockHashes sends a batch of hashes from a peer
func (p *peer) SendBlockHashes(hashes remoteHashes) error {
	if err := p2p.SendMsgData(p.p2pPeer, BlockHashesMsg, &hashes); err != nil {
		return err
	}
	return nil
}

// RequestBlocks fetches a batch of blocks based on the hash values
func (p *peer) RequestBlocks(hashes remoteHashes) error {
	if err := p2p.SendMsgData(p.p2pPeer, GetBlocksMsg, &hashes); err != nil {
		return err
	}
	return nil
}

// SendBlocks sends a batch of blocks
func (p *peer) SendBlocks(blocks remoteBlocks) error {
	if err := p2p.SendMsgData(p.p2pPeer, BlocksMsg, &blocks); err != nil {
		return err
	}
	return nil
}

// SendNewBlock sends a new block
func (p *peer) SendNewBlock(data *xfsgo.Block) error {
	if err := p2p.SendMsgData(p.p2pPeer, NewBlockMsg, data); err != nil {
		return err
	}
	return nil
}

// SendTransactions sends a batch of transactions
func (p *peer) SendTransactions(data remoteTxs) error {
	if err := p2p.SendMsgData(p.p2pPeer, TxMsg, &data); err != nil {
		return err
	}
	return nil
}
