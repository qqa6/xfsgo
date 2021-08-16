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

package avlmerkle

import (
	"sync"
	"xfsgo/common/rawencode"
	"xfsgo/storage/badger"
)

//treeDb stores the tree to the db.
type treeDb struct {
	storage   *badger.Storage
	writeLock sync.Mutex
}

func newTreeDb(db *badger.Storage) *treeDb {
	tdb := &treeDb{
		storage: db,
	}
	return tdb
}

func (db *treeDb) newWriteBatch() *badger.StorageWriteBatch {
	return db.storage.NewWriteBatch()
}

// commitBatch creates StorageWriteBatch to flush persistent data out
func (db *treeDb) commitBatch(batch *badger.StorageWriteBatch) error {
	return db.storage.CommitWriteBatch(batch)
}

func (db *treeDb) getTreeNodeByKey(key []byte) (*TreeNode, error) {
	val, err := db.storage.GetData(key)
	if err != nil {
		return nil, err
	}
	node := &TreeNode{}
	if err = rawencode.Decode(val, node); err != nil {
		return nil, err
	}
	return node, nil
}
