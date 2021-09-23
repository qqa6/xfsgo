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

package miner

import (
	"path"
	"testing"
	"xfsgo"
	"xfsgo/assert"
	"xfsgo/common"
	"xfsgo/storage/badger"
)

var (
	datadir = "./d0"
	//defaultProtocolVersion = uint32(1)
	//coinbasePrivateKey = "MHcCAQEEIONs7L6Y8K552y_EAJokBTQxA3ejz162pdhZzUaO9bdKoAoGCCqGSM49AwEHoUQDQgAE-KtA6lYlnWAYShJJ2aQnScLz79Iv-vmlVjq8bvxfj9VvOKIPQop87jXyQ01QfTbgZorGEserwb2hDwQAp_xzRA"
	defaultCoinbase = "1A2QiH4FYc9c4nsNjCMxygg9HKTK9EJWX5"
)

func TestMiner_Start(t *testing.T) {
	eventBus := xfsgo.NewEventBus()
	datadir = path.Join(path.Dir(datadir), datadir)
	chainDb := badger.New(path.Join(datadir, "chain"))
	stateDb := badger.New(path.Join(datadir, "state"))
	extraDb := badger.New(path.Join(datadir, "extraDB"))
	_, err := xfsgo.WriteTestGenesisBlock(
		stateDb, chainDb)
	assert.Error(t, err)
	bc, err := xfsgo.NewBlockChain(stateDb, chainDb, extraDb, eventBus)
	assert.Error(t, err)
	b := bc.GetHead()
	txpool := xfsgo.NewTxPool(bc.CurrentStateTree, eventBus)
	miner := NewMiner(&Config{
		Coinbase: common.StrB58ToAddress(defaultCoinbase),
	}, stateDb, bc, eventBus, txpool, b.Header.GasLimit)
	_ = miner
	miner.Start()
	select {}
}
