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
	"math/big"
	"xfsgo/common"
	"xfsgo/common/rawencode"
	"xfsgo/storage/badger"

	"github.com/sirupsen/logrus"
)

var (
	blockHashPre   = []byte("bh:")
	blockNumPre    = []byte("bn:")
	lastBlockKey   = []byte("LastBlock")
	headerPrefix   = []byte("h") // headerPrefix + num (uint64 big endian) + hash -> header
	headerTDSuffix = []byte("t") // headerPrefix + num (uint64 big endian) + hash + headerTDSuffix -> td
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

// encodeBlockNumber encodes a block number as big endian uint64
func encodeBlockNumber(number uint64) []byte {
	enc := make([]byte, 8)
	binary.BigEndian.PutUint64(enc, number)
	return enc
}

// headerKey = headerPrefix + num (uint64 big endian) + hash
func headerKey(number uint64, hash common.Hash) []byte {
	return append(append(headerPrefix, encodeBlockNumber(number)...), hash.Bytes()...)
}

// headerTDKey = headerPrefix + num (uint64 big endian) + hash + headerTDSuffix
func headerTDKey(number uint64, hash common.Hash) []byte {
	return append(headerKey(number, hash), headerTDSuffix...)
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
		logrus.Debugf("GetBlockByNumber err height:%d", num)
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

func (db *chainDB) WriteHead(block *Block) error {
	if err := db.WriteCanonNumber(block); err != nil {
		return err
	}
	hash := block.Hash()
	err := db.storage.SetData(lastBlockKey, hash.Bytes())
	return err
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
		logrus.Debugf("WriteCanonNumber err height:%d, hash:%s", block.Height(), block.HashHex())
		return err
	}
	logrus.Debugf("WriteCanonNumber success height:%d, hash:%s", block.Height(), block.HashHex())
	return nil
}

func (db *chainDB) RemoveCanonNumber(block *Block) error {
	var numBuf [8]byte
	binary.LittleEndian.PutUint64(numBuf[:], block.Height())
	key := append(blockNumPre, numBuf[:]...)
	err := db.storage.DelData(key)
	if err != nil {
		return err
	}
	return nil
}

func (db *chainDB) DelUnCanonNumber(block *Block) error {
	var numBuf [8]byte
	binary.LittleEndian.PutUint64(numBuf[:], block.Height())
	key := append(blockNumPre, numBuf[:]...)

	err := db.storage.DelData(key)
	if err != nil {
		return err
	}
	return nil
}

func (db *chainDB) RemoveNumBlock(block *Block) error {
	if err := db.RemoveCanonNumber(block); err != nil {
		return err
	}
	return nil
}

func (db *chainDB) WriteNumBlock(block *Block) error {
	if err := db.WriteCanonNumber(block); err != nil {
		return err
	}
	return nil
}

// Loss of main chain after fork
func (db *chainDB) DelHead(block *Block) error {
	if err := db.DelUnCanonNumber(block); err != nil {
		return err
	}

	blockHash := block.HashPrevBlock()
	// warning:The design here is unreasonable
	return db.storage.SetData(lastBlockKey, blockHash.Bytes())
}

// ReadTd retrieves a block's total difficulty corresponding to the hash.
func (db *chainDB) ReadTd(hash common.Hash, number uint64) *big.Int {

	data, err := db.storage.GetData(headerTDKey(number, hash))
	if err != nil {
		return nil
	}

	td := &big.Int{}
	if err := rawencode.Decode(data, td); err != nil {
		return nil
	}

	return td
}

// WriteTd stores the total difficulty of a block into the database.
func (db *chainDB) WriteTd(hash common.Hash, number uint64, td *big.Int) error {
	data, err := rawencode.Encode(td)
	if err != nil {
		logrus.Errorf("Failed to RLP encode block total difficulty", "err", err)
		return err
	}
	if err := db.storage.SetData(headerTDKey(number, hash), data); err != nil {
		logrus.Errorf("Failed to store block total difficulty", "err", err)
		return err
	}

	return nil
}

// DeleteTd removes all block total difficulty data associated with a hash.
func (db *chainDB) DeleteTd(hash common.Hash, number uint64) error {
	if err := db.storage.DelData(headerTDKey(number, hash)); err != nil {
		logrus.Errorf("Failed to delete block total difficulty", "err", err)
		return err
	}

	return nil
}
