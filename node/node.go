package node

import (
	"log"
	"xblockchain"
	"xblockchain/api"
	configs "xblockchain/cmd/config"
	"xblockchain/p2p"
	"xblockchain/rpc"

	"github.com/sirupsen/logrus"
)

type Node struct {
	// *Opts
	p2pServer  *p2p.Server
	RPCStarter *rpc.ServerStarter
}

// type Opts struct {
// 	P2PListenAddress string
// 	P2PBootstraps    []string
// 	RPCListenAddress string
// }

func New() (*Node, error) {
	temp := configs.GetConfig()
	n := &Node{
		p2pServer: &p2p.Server{
			ListenAddr:     temp.Network.P2PListenAddress,
			BootstrapNodes: temp.Network.P2PBootstraps,
		},
	}
	var err error = nil
	if n.RPCStarter, err = rpc.NewServerStarter(); err != nil {
		return nil, err
	}
	return n, nil
}

func (n *Node) Start() error {
	if err := n.p2pServer.Start(); err != nil {
		return err
	}
	go func() {
		if err := n.RPCStarter.Run(); err != nil {
			logrus.Warnf("Start RPC ERR: %s", err)
		}
	}()
	return nil
}

func (n *Node) RegisterBackend(
	bc *xblockchain.BlockChain,
	miner *xblockchain.Miner,
	wallets *xblockchain.Wallets,
	txPool *xblockchain.TxPool) error {
	chainApiHandler := &api.ChainAPIHandler{
		BlockChain: bc,
	}

	minerApiHandler := &api.MinerAPIHandler{
		Miner: miner,
	}

	walletApiHandler := &api.WalletsHandler{
		Wallets: wallets,
	}

	txApiHandler := &api.TXAPIHandler{
		Wallets:       wallets,
		BlockChain:    bc,
		TxPendingPool: txPool,
	}
	starter := n.RPCStarter
	if err := starter.RegisterName("Chain", chainApiHandler); err != nil {
		log.Fatalf("RPC service register error: %s", err)
		return err
	}
	if err := starter.RegisterName("Wallet", walletApiHandler); err != nil {
		log.Fatalf("RPC service register error: %s", err)
		return err
	}
	if err := starter.RegisterName("Miner", minerApiHandler); err != nil {
		log.Fatalf("RPC service register error: %s", err)
		return err
	}
	if err := starter.RegisterName("Transaction", txApiHandler); err != nil {
		log.Fatalf("RPC service register error: %s", err)
		return err
	}
	return nil
}

func (n *Node) P2PServer() *p2p.Server {
	return n.p2pServer
}
