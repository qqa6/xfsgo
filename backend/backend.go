package backend

import (
	"log"
	"xblockchain"
	configs "xblockchain/cmd/config"
	"xblockchain/node"
	"xblockchain/p2p"
	"xblockchain/storage/badger"
)

type Backend struct {
	txPool        chan string
	blockchain    *xblockchain.BlockChain
	handler       *handler
	blockDb       *badger.Storage
	keysDb        *badger.Storage
	p2pServer     *p2p.Server
	txPendingPool *xblockchain.TXPendingPool
	wallets       *xblockchain.Wallets
	miner         *xblockchain.Miner
	eventBus      *xblockchain.EventBus
	txPoll        *xblockchain.TxPool
}

// type Opts struct {
// 	BlockDbPath string
// 	KeyStoragePath string
// 	Version uint32
// 	Network uint32
// }
func NewBackend(stack *node.Node) (*Backend, error) {
	temp := configs.GetConfig()
	var err error = nil
	back := &Backend{
		txPool:    make(chan string),
		p2pServer: stack.P2PServer(),
	}
	back.eventBus = xblockchain.NewEventBus()
	back.blockDb = badger.New(temp.Blockchain.BlockDbPath)
	back.keysDb = badger.New(temp.Blockchain.KeyStoragePath)
	back.txPendingPool = xblockchain.NewTXPendingPool(100)
	genesisOpts := xblockchain.DefaultGenesisBlockOpts()
	if back.blockchain, err = xblockchain.NewBlockChain(
		genesisOpts, back.blockDb,
		back.eventBus); err != nil {
		return nil, err
	}
	back.wallets = xblockchain.NewWallets(back.keysDb)

	back.txPoll = xblockchain.NewTxPool(back.eventBus, back.blockchain)
	back.miner = xblockchain.NewMiner(back.blockchain, back.wallets, back.txPoll, back.eventBus)

	if err = stack.RegisterBackend(
		back.blockchain, back.miner, back.wallets, back.txPoll); err != nil {
		return nil, err
	}

	if back.handler, err = newHandler(back.blockchain,
		temp.Network.ProtocolVersion, temp.Network.NetworkID, back.eventBus, back.txPoll); err != nil {
		return nil, err
	}
	callFn := back.handler.handlerCallFn
	back.p2pServer.PeerHandlerFn = callFn
	return back, nil
}

func (b *Backend) Start() error {
	b.handler.Start()
	return nil
}

func (b *Backend) close() {
	if err := b.blockDb.Close(); err != nil {
		log.Fatalf("Blocks Storage close errors: %s", err)
	}
	if err := b.keysDb.Close(); err != nil {
		log.Fatalf("Blocks Storage close errors: %s", err)
	}
}
