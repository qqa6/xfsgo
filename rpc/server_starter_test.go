package rpc

import (
	"testing"
	"xblockchain"
	"xblockchain/api"
	"xblockchain/storage/badger"
)

func TestRPCServerStarter_Run(t *testing.T) {
	keyStorage := badger.New("./data0/keys")
	defer func() {
		if err := keyStorage.Close(); err != nil {
			t.Fatalf("Sotrage close errors: %s", err)
		}
	}()
	blocksStorage := badger.New("./data0/blocks")
	defer func() {
		if err := blocksStorage.Close(); err != nil {
			t.Fatalf("Sotrage close errors: %s", err)
		}
	}()
	ws := xblockchain.NewWallets(keyStorage)
	gopt := xblockchain.DefaultGenesisBlockOpts()
	bc, err := xblockchain.NewBlockChain(gopt, blocksStorage, nil)
	if err != nil {
		t.Fatal(err)
	}

	txPendingPool := xblockchain.NewTXPendingPool(10)
	miner := xblockchain.NewMiner(bc, ws, txPendingPool)
	// xblockchain.NewMiner(back.blockchain, back.wallets, back.txPoll, back.eventBus)
	starter, err := NewServerStarter()
	if err != nil {
		t.Fatal(err)
	}
	chainApiHandler := &api.ChainAPIHandler{
		BlockChain: bc,
	}

	minerApiHandler := &api.MinerAPIHandler{
		Miner: miner,
	}

	walletApiHandler := &api.WalletsHandler{
		Wallets: ws,
	}
	err = starter.RegisterName("Chain", chainApiHandler)
	if err != nil {
		t.Fatal(err)
	}

	err = starter.RegisterName("Wallet", walletApiHandler)
	if err != nil {
		t.Fatal(err)
	}

	err = starter.RegisterName("Miner", minerApiHandler)
	if err != nil {
		t.Fatal(err)
	}
	err = starter.Run(":9005")
	if err != nil {
		t.Fatal(err)
	}
}
