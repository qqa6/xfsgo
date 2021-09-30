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
	"log"
	"os"
	"xfsgo"
	"xfsgo/common"
	"xfsgo/miner"
	"xfsgo/node"
	"xfsgo/p2p"
	"xfsgo/storage/badger"
)

// Backend represents the backend server of the xfs and implements the xfs full node service.
type Backend struct {
	config     *Config
	blockchain *xfsgo.BlockChain
	handler    *handler
	p2pServer  p2p.Server
	wallet     *xfsgo.Wallet
	miner      *miner.Miner
	eventBus   *xfsgo.EventBus
	txPool     *xfsgo.TxPool
}

type Params struct {
	NetworkID       uint32
	GenesisFile     string
	Coinbase        common.Address
	ProtocolVersion uint32
}

// Config contains the configuration options of the Backend.
type Config struct {
	*Params
	ChainDB *badger.Storage
	KeysDB  *badger.Storage
	StateDB *badger.Storage
	ExtraDB *badger.Storage
}

// NewBackend constructs and returns a Backend instance by a note in network and config.
// This method is for daemon whick should be started firstly when xfs blockchain runs.
//
func NewBackend(stack *node.Node, config *Config) (*Backend, error) {
	var err error = nil
	back := &Backend{
		config:    config,
		p2pServer: stack.P2PServer(),
	}

	back.eventBus = xfsgo.NewEventBus()
	if config.NetworkID == uint32(1) {
		if _, err = xfsgo.WriteMainNetGenesisBlock(
			back.config.StateDB, back.config.ChainDB); err != nil {
			return nil, err
		}
	} else if config.NetworkID == uint32(2) {
		if _, err = xfsgo.WriteTestNetGenesisBlock(
			back.config.StateDB, back.config.ChainDB); err != nil {
			return nil, err
		}
	} else if len(config.GenesisFile) > 0 {
		var fr *os.File
		if fr, err = os.Open(config.GenesisFile); err != nil {
			return nil, err
		}
		if _, err = xfsgo.WriteGenesisBlock(
			back.config.StateDB, back.config.ChainDB, fr); err != nil {
			return nil, err
		}
	}
	if back.blockchain, err = xfsgo.NewBlockChain(
		back.config.StateDB, back.config.ChainDB, back.config.ExtraDB, back.eventBus); err != nil {
		return nil, err
	}

	back.wallet = xfsgo.NewWallet(back.config.KeysDB)
	back.txPool = xfsgo.NewTxPool(back.blockchain.CurrentStateTree, back.eventBus)

	coinbase := config.Coinbase
	addrdef := back.wallet.GetDefault()

	if !coinbase.Equals(common.Address{}) || addrdef.Equals(common.Address{}) {
		coinbase, err = back.wallet.AddByRandom()
		if err != nil {
			return nil, err
		}
		if err = back.wallet.SetDefault(coinbase); err != nil {
			return nil, err
		}
	}
	//constructs Miner instance.
	back.miner = miner.NewMiner(&miner.Config{
		Coinbase: back.wallet.GetDefault(),
	}, back.config.StateDB, back.blockchain, back.eventBus, back.txPool)
	//Node resgisters apis of baclend on the node  for RPC service.
	if err = stack.RegisterBackend(
		back.config.StateDB, back.blockchain, back.miner, back.wallet, back.txPool); err != nil {
		return nil, err
	}

	if back.handler, err = newHandler(back.blockchain,
		back.config.ProtocolVersion, back.config.NetworkID, back.eventBus, back.txPool); err != nil {
		return nil, err
	}
	back.p2pServer.Bind(&p2p.SimpleProtocol{
		Func: func(p p2p.Peer) error {
			return back.handler.handleNewPeer(p)
		},
	})
	return back, nil
}

func (b *Backend) Start() error {
	b.handler.Start()
	return nil
}

func (b *Backend) BlockChain() *xfsgo.BlockChain {
	return b.blockchain
}
func (b *Backend) EventBus() *xfsgo.EventBus {
	return b.eventBus
}

func (b *Backend) StateDB() *badger.Storage {
	return b.config.StateDB
}

func (b *Backend) close() {
	if err := b.config.ChainDB.Close(); err != nil {
		log.Fatalf("Blocks Storage close errors: %s", err)
	}
	if err := b.config.KeysDB.Close(); err != nil {
		log.Fatalf("Blocks Storage close errors: %s", err)
	}
}
