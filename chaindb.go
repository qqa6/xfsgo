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
	"xfsgo/common"
	"xfsgo/common/rawencode"
	"xfsgo/storage/badger"
)

var (
	blockHashPre = []byte("bh:")
	blockNumPre  = []byte("bn:")
	blockNumHashPre  = []byte("bnh:")
	lastBlockKey = []byte("LastBlock")
)

type chainDB struct {
	storage *badger.Storage
	debug bool
}

func newChainDB(db *badger.Storage) *chainDB {
	return newChainDBN(db, false)
}
func newChainDBN(db *badger.Storage, debug bool) *chainDB {
	tdb := &chainDB{
		storage: db,
		debug: debug,
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
	if db.debug {
		_ = db.WriteBlockNumberHash(block)
	}
	return db.storage.SetData(key, val)
}

func (db *chainDB) WriteBlockNumberHash(block *Block) error {
	height := block.Height()
	var heightbytes = make([]byte, 8)
	binary.BigEndian.PutUint64(heightbytes, height)
	hash := block.Hash()
	// bnh:<height_64bits><hash>
	key := append(blockNumHashPre, heightbytes...)
	key = append(key, hash[:]...)
	val, err := rawencode.Encode(block)
	if err != nil {
		return err
	}
	return db.storage.SetData(key, val)
}


func (db *chainDB) GetBlocksByNumber(num uint64) []*Block {
	var heightbytes = make([]byte, 8)
	binary.BigEndian.PutUint64(heightbytes, num)
	key := append(blockHashPre, heightbytes...)
	blks := make([]*Block, 0)
	db.storage.For(func(k []byte, v []byte) {
		if len(k) < len(blockNumHashPre) + 8 {
			return
		}
		gotkeypre := k[0:len(blockNumHashPre)+8]
		if !bytes.Equal(gotkeypre, key) {
			return
		}
		block := &Block{}
		if err := rawencode.Decode(v, block); err != nil {
			return
		}
		blks = append(blks, block)
	})
	return blks
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
