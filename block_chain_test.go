package xblockchain

import (
	"fmt"
	"testing"
	"time"
	"xblockchain/storage/badger"
)



func TestBlockChain_AddBlock(t *testing.T) {
	storage := badger.New("./data0/blocks")
	defer func() {
		if err := storage.Close(); err != nil {
			t.Fatalf("Sotrage close errors: %s", err)
		}
	}()
	gopt := DefaultGenesisBlockOpts()
	bc,err := NewBlockChain(gopt, storage, nil)
	if err != nil {
		t.Fatal(err)
	}
	for i :=0; i<20; i++ {
		tx, err := NewCoinBaseTransaction(gopt.Address,"")
		if err != nil {
			t.Fatal(err)
		}
		block, err := bc.AddBlock([]*Transaction{tx})
		if err != nil {
			t.Fatal(err)
		}
		_ = block
	}
}

func TestNewBlockChain(t *testing.T) {
	storage := badger.New("./data0/blocks")
	defer func() {
		if err := storage.Close(); err != nil {
			t.Fatalf("Sotrage close errors: %s", err)
		}
	}()
	gopt := DefaultGenesisBlockOpts()
	bc,err := NewBlockChain(gopt, storage, nil)
	if err != nil {
		t.Fatal(err)
	}
	//addr := "12neM97Z8eXTh2Xr2U9n3Qv1AuPC4J7ZAfJB3iiEmwf9f"
	addr := gopt.Address
	balance,err := bc.GetBalanceOfAddress(addr)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("balance of %s: %d", addr,balance)
}

func TestBlockChain_SendTransaction(t *testing.T) {
	storage := badger.New("./data0/blocks")
	defer func() {
		if err := storage.Close(); err != nil {
			t.Fatalf("Sotrage close errors: %s", err)
		}
	}()
	gopt := DefaultGenesisBlockOpts()
	bc,err := NewBlockChain(gopt, storage, nil)
	if err != nil {
		t.Fatal(err)
	}
	//from := ""
	w,err := ImportWalletWithB64("MHcCAQEEILP4MC2s1i1cLTytfrI1ijYnSX43wNP2QUyMsER2V6ImoAoGCCqGSM49AwEHoUQDQgAEb0USbtlHHQp1_R1_53coySi3WaLhMWkBcR6N0Aegg_L8SnCf4mnidbW7aPmLxGhEv6jAN0pKRykJP_BTO5MckQ")
	if err != nil {
		t.Fatal(err)
	}
	from := gopt.Address
	coinbaseTx, err := NewCoinBaseTransaction(from,"")
	if err != nil {
		t.Fatal(err)
	}
	toAddr := "12neM97Z8eXTh2Xr2U9n3Qv1AuPC4J7ZAfJB3iiEmwf9f"
	val := uint64(1)
	tx0, err := NewUTXOTransaction(gopt.Address,toAddr,uint64(1),bc,w.privateKey)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("send transaction from: %s, to: %s, val: %d", w.GetAddress(),toAddr, val)
	block, err := bc.AddBlock([]*Transaction{coinbaseTx,tx0})
	if err != nil {
		t.Fatal(err)
	}
	_ = block
	//log.Printf("")
}

func TestBlockChain_GetBlockById(t *testing.T) {
	storage := badger.New("./data0/blocks")
	defer func() {
		if err := storage.Close(); err != nil {
			t.Fatalf("Sotrage close errors: %s", err)
		}
	}()
	gopt := DefaultGenesisBlockOpts()
	bc,err := NewBlockChain(gopt, storage, nil)
	if err != nil {
		t.Fatal(err)
	}
	itr := bc.Iterator()
	if !itr.HasNext() {
		t.Logf("block chain is empty")
	}
	fmt.Printf("Height    Time                Miner     Hash\n")

	for itr.HasNext() {
		b,err := itr.Next()
		if err != nil{
			t.Fatal(err)
		}
		t := time.Unix(b.Timestamp,0)
		prevHash := b.HashPrevBlock.HexstrFull()
		fmt.Printf("%-9d ", b.Height)
		fmt.Printf("%s ", t.Format("2006-01-02 15:04:05"))
		maddr := b.GetMiner()
		fmt.Printf("%-3s...%-3s ", maddr[:3],maddr[len(maddr)-3:])
		fmt.Printf("%-3s...%-3s", prevHash[:3],prevHash[len(prevHash)-3:])
		fmt.Println()
	}
}

func TestBlockChain_GetBlockHashes(t *testing.T) {
	storage := badger.New("./data0/blocks")
	defer func() {
		if err := storage.Close(); err != nil {
			t.Fatalf("Sotrage close errors: %s", err)
		}
	}()
	gopt := DefaultGenesisBlockOpts()
	bc,err := NewBlockChain(gopt, storage, nil)
	if err != nil {
		t.Fatal(err)
	}
	hashes := bc.GetBlockHashes(0,10)
	for i, hash := range hashes {
		t.Logf("hash[%d]: %s\n", i, hash.Hex())
	}
}