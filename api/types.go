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

type GetBlockByNumberBlockHeader struct {
	Height        uint64         `json:"height"`
	Version       int32          `json:"version"`
	HashPrevBlock common.Hash    `json:"hash_prev_block"`
	Timestamp     uint64         `json:"timestamp"`
	Coinbase      common.Address `json:"coinbase"`
	// merkle tree root hash
	StateRoot        common.Hash `json:"state_root"`
	TransactionsRoot common.Hash `json:"transactions_root"`
	ReceiptsRoot     common.Hash `json:"receipts_root"`
	// pow
	Bits  uint32      `json:"bits"`
	Nonce uint64      `json:"nonce"`
	Hash  common.Hash `json:"hash"`
}

type GetBlockByNumberBlock struct {
	Header       *GetBlockByNumberBlockHeader `json:"header"`
	Transactions []*xfsgo.Transaction         `json:"transactions"`
	Receipts     []*xfsgo.Receipt             `json:"receipts"`
}

type TransferObj struct {
	To        common.Address `json:"to"`
	Nonce     uint64         `json:"nonce"`
	Value     *big.Int       `json:"value"`
	Signature []byte         `json:"signature"`
	Hash      common.Hash    `json:"hash"`
}

type GetBlocks []*GetBlockByNumberBlock
type transactions []*xfsgo.Transaction
