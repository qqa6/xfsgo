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
	"encoding/binary"
	"xfsgo/common"
	"xfsgo/common/rawencode"
	"xfsgo/storage/badger"
)

var (
	blockHashPre = []byte("bh:")
	blockNumPre  = []byte("bn:")
	lastBlockKey = []byte("LastBlock")
)

type chainDB struct {
	storage *badger.Storage
}

func newChainDB(db *badger.Storage) *chainDB {
	tdb := &chainDB{
		storage: db,
	}
	return tdb
}

func (db *chainDB) newWriteBatch() *badger.StorageWriteBatch {
	return db.storage.NewWriteBatch()
}

func (db *chainDB) commitBatch(batch *badger.StorageWriteBatch) error {
	return db.storage.CommitWriteBatch(batch)
}

func (db *chainDB) GetBlockByHash(hash common.Hash) *Block {
	key := append(blockHashPre, hash.Bytes()...)
	val, err := db.storage.GetData(key)
	if err != nil {
		return nil
	}
	block := &Block{}
	if err := rawencode.Decode(val, block); err != nil {
		return nil
	}
	return block
}

func (db *chainDB) GetBlockByNumber(num uint64) *Block {
	var numBuf [8]byte
	binary.LittleEndian.PutUint64(numBuf[:], num)
	key := append(blockNumPre, numBuf[:]...)
	val, err := db.storage.GetData(key)
	if err != nil {
		return nil
	}
	hash := common.Bytes2Hash(val)
	return db.GetBlockByHash(hash)
}

func (db *chainDB) GetHeadBlock() *Block {

	val, err := db.storage.GetData(lastBlockKey)
	if err != nil {
		return nil
	}
	hash := common.Bytes2Hash(val)
	return db.GetBlockByHash(hash)
}

func (db *chainDB) WriteBlock(block *Block) error {
	hash := block.Hash()
	key := append(blockHashPre, hash.Bytes()...)
	val, err := rawencode.Encode(block)
	if err != nil {
		return err
	}
	return db.storage.SetData(key, val)
}

func (db *chainDB) WriteCanonNumber(block *Block) error {
	var numBuf [8]byte
	binary.LittleEndian.PutUint64(numBuf[:], block.Height())
	key := append(blockNumPre, numBuf[:]...)
	blockHash := block.Hash()
	err := db.storage.SetData(key, blockHash.Bytes())
	if err != nil {
		return err
	}
	return nil
}

func (db *chainDB) WriteHead(block *Block) error {
	if err := db.WriteCanonNumber(block); err != nil {
		return err
	}
	blockHash := block.Hash()
	return db.storage.SetData(lastBlockKey, blockHash.Bytes())
}
