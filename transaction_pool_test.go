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

package xfsgo

import (
	"math/big"
	"testing"
	"xfsgo/assert"
	"xfsgo/crypto"
	"xfsgo/storage/badger"
)

var stateDbPath = "./d0/state"
var cianDbPath = "./d0/chain"
var coinbasePrivateKey = "MHcCAQEEIONs7L6Y8K552y_EAJokBTQxA3ejz162pdhZzUaO9bdKoAoGCCqGSM49AwEHoUQDQgAE-KtA6lYlnWAYShJJ2aQnScLz79Iv-vmlVjq8bvxfj9VvOKIPQop87jXyQ01QfTbgZorGEserwb2hDwQAp_xzRA"

//var defaultCoinbase = "1A2QiH4FYc9c4nsNjCMxygg9HKTK9EJWX5"
func TestTxPool_Add(t *testing.T) {
	stateDb := badger.New(stateDbPath)
	chainDb := badger.New(cianDbPath)
	defer func() {
		if err := stateDb.Close(); err != nil {
			t.Fatal(err)
		}
		if err := chainDb.Close(); err != nil {
			t.Fatal(err)
		}
	}()
	genesisBlock, err := WriteTestGenesisBlock(stateDb, chainDb)
	assert.Error(t, err)
	genesisBlockStateRoot := genesisBlock.StateRoot()
	st := NewStateTree(stateDb, genesisBlockStateRoot.Bytes())
	st.Print()

	eventBus := NewEventBus()
	go func() {
		txPreEventSub := eventBus.Subscript(TxPreEvent{})
		defer txPreEventSub.Unsubscribe()
		for {
			select {
			case e := <-txPreEventSub.Chan():
				txper := e.(TxPreEvent)
				recTxHash := txper.Tx.Hash()
				t.Logf("receTxHash: %s\n", recTxHash.Hex())
			}
		}
	}()
	txPoll := NewTxPool(func() *StateTree {
		return st
	},nil, eventBus)

	key1, err := crypto.B64StringDecodePrivateKey(coinbasePrivateKey)
	assert.Error(t, err)
	key2, err := crypto.GenPrvKey()
	assert.Error(t, err)
	toAddr := crypto.DefaultPubKey2Addr(key2.PublicKey)
	tx0 := &Transaction{
		To:    toAddr,
		Value: new(big.Int).SetInt64(100),
	}
	err = tx0.SignWithPrivateKey(key1)
	assert.Error(t, err)
	tx0Hash := tx0.Hash()
	t.Logf("tx0: %s\n", tx0Hash.Hex())
	tx1 := &Transaction{
		To:    toAddr,
		Value: new(big.Int).SetInt64(100),
	}
	err = tx1.SignWithPrivateKey(key1)
	assert.Error(t, err)
	tx1Hash := tx1.Hash()
	t.Logf("tx1: %s\n", tx1Hash.Hex())
	err = txPoll.Add(tx0)
	assert.Error(t, err)
	err = txPoll.Add(tx1)
	assert.Error(t, err)
	pending := txPoll.pending
	assert.Equal(t, len(pending), 2)
	select {}
}
