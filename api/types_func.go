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
	"xfsgo"
	"xfsgo/common"
)

func NewBlockByNumBlockHeader(data *xfsgo.BlockHeader, hash common.Hash) *GetBlockByNumberBlockHeader {
	return &GetBlockByNumberBlockHeader{
		Height:           data.Height,
		Version:          data.Version,
		HashPrevBlock:    data.HashPrevBlock,
		Timestamp:        data.Timestamp,
		Coinbase:         data.Coinbase,
		StateRoot:        data.StateRoot,
		TransactionsRoot: data.TransactionsRoot,
		ReceiptsRoot:     data.ReceiptsRoot,
		Bits:             data.Bits,
		Nonce:            data.Nonce,
		Hash:             hash,
	}
}

func NewBlockByNumberBlock(data *xfsgo.Block, headers *GetBlockByNumberBlockHeader) *GetBlockByNumberBlock {
	return &GetBlockByNumberBlock{
		Header:       headers,
		Transactions: data.Transactions,
		Receipts:     data.Receipts,
	}
}

func NewTransferObj(tx *xfsgo.Transaction) *TransferObj {
	return &TransferObj{
		To:        tx.To,
		Nonce:     tx.Nonce,
		Value:     tx.Value,
		Signature: tx.Signature,
		Hash:      tx.Hash(),
	}
}
