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

package node

import (
	"crypto/ecdsa"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"xfsgo"
	"xfsgo/api"
	"xfsgo/crypto"
	"xfsgo/miner"
	"xfsgo/p2p"
	"xfsgo/p2p/discover"
	"xfsgo/storage/badger"

	"github.com/sirupsen/logrus"
)

// Node is a container on which services can be registered.
type Node struct {
	// *Opts
	config    *Config
	p2pServer p2p.Server
	rpcServer *xfsgo.RPCServer
}

type Config struct {
	P2PListenAddress string
	ProtocolVersion  uint8
	P2PBootstraps    []string
	P2PStaticNodes   []string
	NodeDBPath       string
	RPCConfig        *xfsgo.RPCConfig
}

const nodedbKeyName = "/dbkey"

// New creates a new P2P node, ready for protocol registration.
func New(config *Config) (*Node, error) {
	var key *ecdsa.PrivateKey
	nodedbKeyDir := config.NodeDBPath + nodedbKeyName
	NewNodeDBFile(config.NodeDBPath, nodedbKeyName)
	EncodeKey, err := GetNodeDB(nodedbKeyDir)
	if err != nil {
		return nil, err
	}
	if EncodeKey != "" {
		key, _ = crypto.B64StringDecodePrivateKey(EncodeKey)
	} else {
		key, _ = crypto.GenPrvKey()
		str, err := crypto.PrivateKeyEncodeB64String(key)
		if err != nil {
			return nil, err
		}
		err = SetNodeDB(nodedbKeyDir, str)
		if err != nil {
			return nil, err
		}
	}

	bootstraps := make([]*discover.Node, 0)
	for _, nodeUri := range config.P2PBootstraps {
		node, err := discover.ParseNode(nodeUri)
		if err != nil {
			logrus.Warnf("parse node uri err: %s", err)
		}
		bootstraps = append(bootstraps, node)
	}
	staticNodes := make([]*discover.Node, 0)
	for _, nodeUri := range config.P2PStaticNodes {
		node, err := discover.ParseNode(nodeUri)
		if err != nil {
			logrus.Warnf("parse node uri err: %s", err)
		}
		staticNodes = append(staticNodes, node)
	}
	p2pServer := p2p.NewServer(p2p.Config{
		ListenAddr:      config.P2PListenAddress,
		ProtocolVersion: config.ProtocolVersion,
		Key:             key,
		BootstrapNodes:  bootstraps,
		StaticNodes:     staticNodes,
		Discover:        true,
		MaxPeers:        10,
		NodeDBPath:      config.NodeDBPath,
	})
	n := &Node{
		config:    config,
		p2pServer: p2pServer,
	}
	n.rpcServer = xfsgo.NewRPCServer(config.RPCConfig)
	return n, nil
}

func NewNodeDBFile(filemkdir string, filename string) error {
	nodedb := filemkdir + filename
	if FileExist(nodedb) {
		return errors.New("file already exists")
	}
	os.MkdirAll(filemkdir, os.ModePerm)
	// if err != nil {
	_, err := os.Create(nodedb)
	return err
	// }
}

func GetNodeDB(filename string) (string, error) {
	f, err := os.OpenFile(filename, os.O_RDONLY, 0600)
	if err != nil {
		return "", err
	} else {
		contentByte, err := ioutil.ReadAll(f)
		return string(contentByte), err
	}
}

func SetNodeDB(filename string, key string) error {
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return err
	} else {
		_, err = f.Write([]byte(key))
		return err
	}
}

func FileExist(path string) bool {
	_, err := os.Lstat(path)
	return !os.IsNotExist(err)
}

// Start starts p2p networking and RPC services runs in a goroutine.
// Node can only be started once.
func (n *Node) Start() error {
	if err := n.p2pServer.Start(); err != nil {
		return err
	}
	go func() {
		if err := n.rpcServer.Start(); err != nil {
			logrus.Errorf("start rpc err: %s", err)
		}
	}()
	return nil
}

//RegisterBackend registers built-in APIs.
func (n *Node) RegisterBackend(
	stateDb *badger.Storage,
	bc *xfsgo.BlockChain,
	miner *miner.Miner,
	wallet *xfsgo.Wallet,
	txPool *xfsgo.TxPool) error {
	chainApiHandler := &api.ChainAPIHandler{
		BlockChain:    bc,
		TxPendingPool: txPool,
	}
	minerApiHandler := &api.MinerAPIHandler{
		Miner: miner,
	}

	walletApiHandler := &api.WalletHandler{
		Wallet:        wallet,
		BlockChain:    bc,
		TxPendingPool: txPool,
	}

	txPoolHandler := &api.TxPoolHandler{
		TxPool: txPool,
	}
	stateHandler := &api.StateAPIHandler{
		StateDb: stateDb,
	}
	if err := n.rpcServer.RegisterName("Chain", chainApiHandler); err != nil {
		log.Fatalf("RPC service register error: %s", err)
		return err
	}

	if err := n.rpcServer.RegisterName("Wallet", walletApiHandler); err != nil {
		log.Fatalf("RPC service register error: %s", err)
		return err
	}
	if err := n.rpcServer.RegisterName("Miner", minerApiHandler); err != nil {
		log.Fatalf("RPC service register error: %s", err)
		return err
	}
	if err := n.rpcServer.RegisterName("TxPool", txPoolHandler); err != nil {
		log.Fatalf("RPC service register error: %s", err)
		return err
	}
	if err := n.rpcServer.RegisterName("State", stateHandler); err != nil {
		log.Fatalf("RPC service register error: %s", err)
		return err
	}
	return nil
}

func (n *Node) P2PServer() p2p.Server {
	return n.p2pServer
}
