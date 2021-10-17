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
	"fmt"
	"math/big"
	"strings"
	"testing"
	"xfsgo/assert"
	"xfsgo/crypto"
	"xfsgo/storage/badger"
)

func TestWriteGenesisBlock(t *testing.T) {
	var stateDb = badger.New("./tmpdir/state")
	var chainDb = badger.New("./tmpdir/chain")
	defer func() {
		if err := stateDb.Close(); err != nil {
			t.Fatal(err)
		}
		if err := chainDb.Close(); err != nil {
			t.Fatal(err)
		}
	}()
	coinbaseKey, err := crypto.GenPrvKey()
	assert.Error(t, err)
	coinbaseAddr := crypto.DefaultPubKey2Addr(coinbaseKey.PublicKey)
	bits := BigByZip(maxTarget)
	initCoinbaseBalance := int64(1000)
	jsonStr := fmt.Sprintf(`{
	"bits": %d,
	"nonce": 0,
	"coinbase": "%s",
	"accounts": {
		"%s": {"balance": "%d"}
	}
}`, bits, coinbaseAddr.B58String(), coinbaseAddr.B58String(), initCoinbaseBalance)
	block, err := WriteGenesisBlock(stateDb, chainDb, strings.NewReader(jsonStr))
	assert.Error(t, err)
	assert.HashEqual(t, block.HashPrevBlock(), emptyHash)
	cdb := newChainDB(chainDb)
	gotBlock := cdb.GetBlockByHash(block.Hash())
	assert.HashEqual(t, block.Hash(), gotBlock.Hash())
	stateRootBs := block.StateRoot()
	stateTree := NewStateTree(stateDb, stateRootBs.Bytes())
	gotBalance := stateTree.GetBalance(coinbaseAddr)
	wantBalance := new(big.Int).SetInt64(initCoinbaseBalance)
	assert.BigIntEqual(t, gotBalance, wantBalance)
	assert.Error(t, err)
	t.Logf("init coinbase address: %s\n", coinbaseAddr.B58String())
}

func TestWriteTestNetGenesisBlock(t *testing.T) {
	var stateDb = badger.New("./d0")
	var chainDb = badger.New("./d1")
	defer func() {
		if err := stateDb.Close(); err != nil {
			t.Fatal(err)
		}
		if err := chainDb.Close(); err != nil {
			t.Fatal(err)
		}
	}()
	block, err := WriteTestNetGenesisBlock(stateDb, chainDb)
	assert.Error(t, err)
	cdb := newChainDB(chainDb)
	gotBlock := cdb.GetBlockByHash(block.Hash())
	assert.HashEqual(t, block.Hash(), gotBlock.Hash())
	stateRootBs := gotBlock.StateRoot()
	stateTree := NewStateTree(stateDb, stateRootBs.Bytes())
	gotBalance := stateTree.GetBalance(gotBlock.Coinbase())
	wantBalance := new(big.Int).SetInt64(10000)
	assert.BigIntEqual(t, gotBalance, wantBalance)
}
