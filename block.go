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

//
package xfsgo

import (
	"encoding/json"
	"math/big"
	"xfsgo/avlmerkle"
	"xfsgo/common"
	"xfsgo/common/ahash"
	"xfsgo/common/rawencode"
)

var (
	emptyHash   = common.Bytes2Hash([]byte{})
	noneAddress = common.Bytes2Address([]byte{})
)

const version0 = uint32(0)
// BlockHeader represents a block header in the xfs blockchain.
// It is importance to note that the BlockHeader includes StateRoot,TransactionsRoot
// and ReceiptsRoot fields which implement the state management of the xfs blockchain.
type BlockHeader struct {
	Height        uint64         `json:"height"`
	Version       int32          `json:"version"`
	HashPrevBlock common.Hash    `json:"hash_prev_block"`
	Timestamp     uint64         `json:"timestamp"`
	Coinbase      common.Address `json:"coinbase"`
	// merkle tree root hash
	StateRoot        common.Hash `json:"state_root"`
	TransactionsRoot common.Hash `json:"transactions_root"`
	ReceiptsRoot     common.Hash `json:"receipts_root"`
	GasLimit *big.Int `json:"gas_limit"`
	GasUsed *big.Int `json:"gas_used"`
	// pow consensus.
	Bits  uint32 `json:"bits"`
	Nonce uint64 `json:"nonce"`
}

func (header *BlockHeader) Encode() ([]byte, error) {
	return json.Marshal(header)
}

func (header *BlockHeader) Decode(data []byte) error {
	return json.Unmarshal(data, header)
}
func (header *BlockHeader) clone() *BlockHeader {
	p := *header
	return &p
}
func (header *BlockHeader) copyTrim() *BlockHeader {
	h := header.clone()
	h.Nonce = 0
	return h
}
func (header *BlockHeader) String() string {
	jb, err := json.Marshal(header)
	if err != nil {
		return ""
	}
	return string(jb)
}

type Block struct {
	Header       *BlockHeader   `json:"header"`
	Transactions []*Transaction `json:"transactions"`
	Receipts     []*Receipt     `json:"receipts"`
}

// NewBlock creates a new block. The input data, txs and receipts are copied,
// changes to header and to the field values will not affect the block.
//
// The values of TransactionsRoot, ReceiptsRoot in header
// are ignored and set to values derived from the given txs, and receipts.
func NewBlock(header *BlockHeader, txs []*Transaction, receipts []*Receipt) *Block {
	b := &Block{
		Header: header,
	}
	if len(txs) == 0 {
		b.Header.TransactionsRoot = emptyHash
	} else {
		b.Header.TransactionsRoot = CalcTxsRootHash(txs)
		b.Transactions = make([]*Transaction, len(txs))
		copy(b.Transactions, txs)
	}
	if len(receipts) == 0 {
		b.Header.ReceiptsRoot = emptyHash
	} else {
		b.Header.ReceiptsRoot = CalcReceiptRootHash(receipts)
		b.Receipts = make([]*Receipt, len(receipts))
		copy(b.Receipts, receipts)
	}
	return b
}

func (b *Block) GetHeader() *BlockHeader {
	return b.Header
}

// CalcTxsRootHash returns the root hash of transactions merkle tree
// by creating a avl merkle tree with transactions as nodes of the tree.
func CalcTxsRootHash(txs []*Transaction) common.Hash {
	tree := avlmerkle.NewTree(nil, nil)
	for _, tx := range txs {
		data, _ := rawencode.Encode(tx)
		txHash := ahash.SHA256(data)
		tree.Put(txHash, data)
	}
	return common.Bytes2Hash(tree.Checksum())
}

// CalcReceiptRootHash returns the root hash of receipt merkle tree
// by creating a avl merkle tree with receipts as nodes of the tree.
// This function is for contract code to check the execution result quickly.
func CalcReceiptRootHash(recs []*Receipt) common.Hash {
	tree := avlmerkle.NewTree(nil, nil)
	for _, rec := range recs {
		data, _ := rawencode.Encode(rec)
		recHash := ahash.SHA256(data)
		tree.Put(recHash, data)
	}
	return common.Bytes2Hash(tree.Checksum())
}

func (b *Block) Encode() ([]byte, error) {
	return json.Marshal(b)
}

func (b *Block) Decode(data []byte) error {
	return json.Unmarshal(data, b)
}

func (b *Block) HashPrevBlock() common.Hash {
	return b.Header.HashPrevBlock
}

func (b *Block) HashNoNonce() common.Hash {
	header := b.Header.copyTrim()
	data, _ := rawencode.Encode(header)
	hash := ahash.SHA256(data)
	return common.Bytes2Hash(hash)
}

func (b *Block) Hash() common.Hash {
	data, _ := rawencode.Encode(b.Header)
	hash := ahash.SHA256(data)
	return common.Bytes2Hash(hash)
}
func (b *Block) HashHex() string {
	hash := b.Hash()
	return hash.Hex()
}
func (b *Block) Height() uint64 {
	return b.Header.Height
}

func (b *Block) StateRoot() common.Hash {
	return b.Header.StateRoot
}

func (b *Block) Coinbase() common.Address {
	return b.Header.Coinbase
}

func (b *Block) TransactionRoot() common.Hash {
	return b.Header.TransactionsRoot
}

func (b *Block) ReceiptsRoot() common.Hash {
	return b.Header.ReceiptsRoot
}

func (b *Block) Bits() uint32 {
	return b.Header.Bits
}

func (b *Block) Nonce() uint64 {
	return b.Header.Nonce
}

func (b *Block) UpdateNonce(nonce uint64) common.Hash {
	b.Header.Nonce = nonce
	return b.Hash()
}

func (b *Block) Timestamp() uint64 {
	return b.Header.Timestamp
}

func (b *Block) String() string {
	jb, err := json.Marshal(b)
	if err != nil {
		return ""
	}
	return string(jb)
}
