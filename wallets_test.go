package xblockchain

import (
	"testing"
	"xblockchain/storage/badger"
)

func TestWallets_NewWallet(t *testing.T) {
	storage := badger.New("./data0/keys")
	defer func() {
		if err := storage.Close(); err != nil {
			t.Fatalf("Sotrage close errors: %s", err)
		}
	}()
	ws := NewWallets(storage)
	addr,err := ws.NewWallet()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("addr: %s", addr)
}

func TestWallets_List(t *testing.T) {
	storage := badger.New("./data0/keys")
	defer func() {
		if err := storage.Close(); err != nil {
			t.Fatalf("Sotrage close errors: %s", err)
		}
	}()
	ws := NewWallets(storage)
	warr := ws.List()
	for _, wallet := range warr {
		t.Logf("addr: %s", wallet.GetAddress())
	}
}

func TestWallets_GetDefault(t *testing.T) {
	storage := badger.New("./data0/keys")
	defer func() {
		if err := storage.Close(); err != nil {
			t.Fatalf("Sotrage close errors: %s", err)
		}
	}()
	ws := NewWallets(storage)
	addr := ws.GetDefault()
	t.Logf("addr: %s", addr)
}