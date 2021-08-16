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
	"bytes"
	"encoding/binary"
	"encoding/json"
	"xfsgo/common"
	"xfsgo/common/rawencode"
	"xfsgo/storage/badger"
)

var (
	txPre            = []byte("tx:")
	txIndexPre       = []byte("txIndex:")
	receiptPre       = []byte("receipt:")
	blockReceiptsPre = []byte("bh:receipt:")
)

type extraDB struct {
	storage *badger.Storage
}

func newExtraDB(db *badger.Storage) *extraDB {
	tdb := &extraDB{
		storage: db,
	}
	return tdb
}

func (db *extraDB) newWriteBatch() *badger.StorageWriteBatch {
	return db.storage.NewWriteBatch()
}

func (db *extraDB) commitBatch(batch *badger.StorageWriteBatch) error {
	return db.storage.CommitWriteBatch(batch)
}

type txIndex struct {
	BlockHash  common.Hash `json:"block_hash"`
	BlockIndex uint64      `json:"block_index"`
	Index      uint64      `json:"index"`
}

func (t *txIndex) Encode() ([]byte, error) {
	return json.Marshal(t)
}

func (t *txIndex) Decode(data []byte) error {
	return json.Unmarshal(data, t)
}

func (db *extraDB) GetTransactionByHash(txHash common.Hash) *Transaction {
	key := append(txPre, txHash.Bytes()...)
	txData, err := db.storage.GetData(key)
	if err != nil {
		return nil
	}
	tx := &Transaction{}
	if err = rawencode.Decode(txData, tx); err != nil {
		return nil
	}
	return tx
}
func (db *extraDB) WriteBlockTransaction(block *Block) error {
	for i, tx := range block.Transactions {
		txData, txHash, err := common.ObjSHA256(tx)
		if err != nil {
			return err
		}
		txKey := append(txPre, txHash...)
		if err = db.storage.SetData(txKey, txData); err != nil {
			return err
		}
		index := &txIndex{
			BlockHash:  block.Hash(),
			BlockIndex: block.Height(),
			Index:      uint64(i),
		}
		indexData, err := rawencode.Encode(index)
		if err != nil {
			return err
		}
		indexKey := append(txIndexPre, txHash...)
		if err = db.storage.SetData(indexKey, indexData); err != nil {
			return err
		}
	}
	return nil
}

func (db *extraDB) WriteReceipts(receipts []*Receipt) error {
	for _, receipt := range receipts {
		data, err := rawencode.Encode(receipt)
		if err != nil {
			return err
		}
		key := append(receiptPre, receipt.TxHash.Bytes()...)
		if err = db.storage.SetData(key, data); err != nil {
			return err
		}
	}
	return nil
}

func (db *extraDB) GetReceipt(txHash common.Hash) *Receipt {
	key := append(receiptPre, txHash.Bytes()...)
	data, err := db.storage.GetData(key)
	if err != nil {
		return nil
	}
	r := &Receipt{}
	if err = rawencode.Decode(data, r); err != nil {
		return nil
	}
	return r
}

func (db *extraDB) WriteBlockReceipts(block *Block) error {
	buf := bytes.NewBuffer(nil)
	for _, receipt := range block.Receipts {
		data, err := rawencode.Encode(receipt)
		if err != nil {
			return err
		}
		var dataLenBuf [4]byte
		dataLen := uint32(len(data))
		binary.LittleEndian.PutUint32(dataLenBuf[:], dataLen)
		buf.Write(dataLenBuf[:])
		buf.Write(data)
	}
	blockHash := block.Hash()
	key := append(blockReceiptsPre, blockHash.Bytes()...)
	if err := db.storage.SetData(key, buf.Bytes()); err != nil {
		return err
	}
	return nil
}

func (db *extraDB) GetBlockReceipts(hash common.Hash) []*Receipt {
	key := append(blockReceiptsPre, hash.Bytes()...)
	data, err := db.storage.GetData(key)
	if err != nil {
		return nil
	}
	tmp := make([]*Receipt, 0)
	buf := bytes.NewBuffer(data)
	for {
		var dataLenBuf [4]byte
		_, err = buf.Read(dataLenBuf[:])
		if err != nil {
			break
		}
		dataLen := binary.LittleEndian.Uint32(dataLenBuf[:])
		var dataBuf = make([]byte, dataLen)
		r := &Receipt{}
		if err = rawencode.Decode(dataBuf, r); err != nil {
			return nil
		}
		tmp = append(tmp, r)
	}
	return tmp
}
