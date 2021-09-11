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
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/big"
	"strings"
	"xfsgo/common"
	"xfsgo/storage/badger"

	"github.com/sirupsen/logrus"
)

var (
	maxTarget = new(big.Int).Lsh(big0xff, ((32-4)*8)+6)
)

// WriteGenesisBlock constructs the genesis blcok for the blockchain and stores it in the hd.
func WriteGenesisBlock(stateDB, chainDB *badger.Storage, reader io.Reader) (*Block, error) {
	contents, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	// Genesis specifies the header fields, state of a genesis block. It also defines accounts
	var genesis struct {
		Version       int32  `json:"version"`
		HashPrevBlock string `json:"hash_prev_block"`
		Timestamp     string `json:"timestamp"`
		Coinbase      string `json:"coinbase"`
		Bits          uint32 `json:"bits"`
		Nonce         uint64 `json:"nonce"`
		Accounts      map[string]struct {
			Balance string `json:"balance"`
		} `json:"accounts"`
	}
	if err = json.Unmarshal(contents, &genesis); err != nil {
		return nil, err
	}

	stateTree := NewStateTree(stateDB, nil)
	//logrus.Debugf("initialize genesis account count: %d", len(genesis.Accounts))
	for addr, a := range genesis.Accounts {
		address := common.B58ToAddress([]byte(addr))
		balance := common.ParseString2BigInt(a.Balance)
		stateTree.AddBalance(address, balance)
		//logrus.Debugf("initialize genesis account: %s, balance: %d", address, balance)
	}
	stateTree.UpdateAll()
	timestamp := common.ParseString2BigInt(genesis.Timestamp)
	var coinbase common.Address
	if genesis.Coinbase != "" {
		coinbase = common.B58ToAddress([]byte(genesis.Coinbase))
	}
	rootHash := common.Bytes2Hash(stateTree.Root())
	HashPrevBlock := common.Hex2Hash(genesis.HashPrevBlock)
	block := NewBlock(&BlockHeader{
		Nonce:         genesis.Nonce,
		HashPrevBlock: HashPrevBlock,
		Timestamp:     timestamp.Uint64(),
		Coinbase:      coinbase,
		Bits:          genesis.Bits,
		StateRoot:     rootHash,
	}, nil, nil)
	chain := newChainDB(chainDB)
	if old := chain.GetBlockByHash(block.Hash()); old != nil {
		logrus.Infof("get genesis block hash: %s", old.HashHex())
		return old, nil
	}
	logrus.Infof("write genesis block hash: %s", block.HashHex())
	if err = stateTree.Commit(); err != nil {
		return nil, err
	}
	if err = chain.WriteBlock(block); err != nil {
		return nil, err
	}
	if err = chain.WriteHead(block); err != nil {
		return nil, err
	}
	return block, nil
}

// WriteMainNetGenesisBlock constructs and stores a genesis blcok with default for xfs blockchain in mainnet model.
//
func WriteMainNetGenesisBlock(stateDB, blockDB *badger.Storage) (*Block, error) {
	bits := BigByZip(maxTarget)
	jsonStr := fmt.Sprintf(`{
	"nonce": 0,
	"bits": %d,
	"coinbase": "1Eux8FG5RuaEiqRKEZgbSqHZwhskDjmS2p",
	"accounts": {
		"1Eux8FG5RuaEiqRKEZgbSqHZwhskDjmS2p": {"balance": "10000"}
	}
}`, bits)
	return WriteGenesisBlock(stateDB, blockDB, strings.NewReader(jsonStr))
}

// WriteMainNetGenesisBlock constructs and stores a genesis blcok with default for xfs blockchain in testnet model.
func WriteTestNetGenesisBlock(stateDB, blockDB *badger.Storage) (*Block, error) {
	bits := BigByZip(maxTarget)
	jsonStr := fmt.Sprintf(`{
	"nonce": 0,
	"bits": %d,
	"coinbase": "18eAREmwhthQZeW3LASBXWi6ibPjjpYC6j",
	"accounts": {
		"18eAREmwhthQZeW3LASBXWi6ibPjjpYC6j": {"balance": "10000"}
	}
}`, bits)
	return WriteGenesisBlock(stateDB, blockDB, strings.NewReader(jsonStr))
}

func WriteTestGenesisBlock(stateDB, blockDB *badger.Storage) (*Block, error) {
	bits := BigByZip(maxTarget)
	jsonStr := fmt.Sprintf(`{
	"nonce": 0,
	"bits": %d,
	"coinbase": "1A2QiH4FYc9c4nsNjCMxygg9HKTK9EJWX5",
	"accounts": {
		"1A2QiH4FYc9c4nsNjCMxygg9HKTK9EJWX5": {"balance": "10000000"}
	}
}`, bits)
	return WriteGenesisBlock(stateDB, blockDB, strings.NewReader(jsonStr))
}
