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

package api

import (
	"math/big"
	"xfsgo"
	"xfsgo/common"
)

type EmptyArgs = interface{}

type StateObjResp struct {
	Address string `json:"address"`
	Balance string `json:"balance"`
	Nonce   uint64 `json:"nonce"`
}

type BlockHeaderResp struct {
	Height        uint64         `json:"height"`
	Version       uint32         `json:"version"`
	HashPrevBlock common.Hash    `json:"hash_prev_block"`
	Timestamp     uint64         `json:"timestamp"`
	Coinbase      common.Address `json:"coinbase"`
	// merkle tree root hash
	StateRoot        common.Hash `json:"state_root"`
	TransactionsRoot common.Hash `json:"transactions_root"`
	ReceiptsRoot     common.Hash `json:"receipts_root"`
	GasLimit         *big.Int    `json:"gas_limit"`
	GasUsed          *big.Int    `json:"gas_used"`
	// pow
	Bits  uint32      `json:"bits"`
	Nonce uint64      `json:"nonce"`
	Hash  common.Hash `json:"hash"`
}

type BlockResp struct {
	Height        uint64         `json:"height"`
	Version       uint32         `json:"version"`
	HashPrevBlock common.Hash    `json:"hash_prev_block"`
	Timestamp     uint64         `json:"timestamp"`
	Coinbase      common.Address `json:"coinbase"`
	// merkle tree root hash
	StateRoot        common.Hash `json:"state_root"`
	TransactionsRoot common.Hash `json:"transactions_root"`
	ReceiptsRoot     common.Hash `json:"receipts_root"`
	GasLimit         *big.Int    `json:"gas_limit"`
	GasUsed          *big.Int    `json:"gas_used"`
	// pow
	Bits         uint32               `json:"bits"`
	Nonce        uint64               `json:"nonce"`
	Hash         common.Hash          `json:"hash"`
	Transactions []*xfsgo.Transaction `json:"transactions"`
	Receipts     []*xfsgo.Receipt     `json:"receipts"`
}

type TransactionResp struct {
	Version   uint32         `json:"version"`
	To        common.Address `json:"to"`
	GasPrice  *big.Int       `json:"gas_price"`
	GasLimit  *big.Int       `json:"gas_limit"`
	Nonce     uint64         `json:"nonce"`
	Value     *big.Int       `json:"value"`
	Signature []byte         `json:"signature"`
	Hash      common.Hash    `json:"hash"`
}

type MinStatusResp struct {
	Status        bool   `json:"status"`
	LastStartTime string `json:"last_start_time"`
	Workers       int    `json:"workers"`
	Coinbase      string `json:"coinbase"`
	GasPrice      string `json:"gas_price"`
	GasLimit      string `json:"gas_limit"`
	HashRate      int    `json:"hash_rate"`
}

type GetBlockChains []*xfsgo.Block
type transactions []*xfsgo.Transaction
type TransactionsResp []*TransactionResp

func coverBlock2Resp(block *xfsgo.Block, dst **BlockResp) error {
	if block == nil {
		return nil
	}
	result := new(BlockResp)
	if err := common.Objcopy(block.Header, result); err != nil {
		return err
	}
	if err := common.Objcopy(block, result); err != nil {
		return err
	}
	result.Hash = block.Hash()
	*dst = result
	return nil
}

func coverBlockHeader2Resp(block *xfsgo.Block, dst **BlockHeaderResp) error {
	if block == nil {
		return nil
	}
	result := new(BlockHeaderResp)
	if err := common.Objcopy(block.Header, result); err != nil {
		return err
	}
	result.Hash = block.Hash()
	*dst = result
	return nil
}

func coverTx2Resp(tx *xfsgo.Transaction, dst **TransactionResp) error {
	if tx == nil {
		return nil
	}
	result := new(TransactionResp)
	if err := common.Objcopy(tx, result); err != nil {
		return err
	}
	result.Hash = tx.Hash()
	*dst = result
	return nil
}
