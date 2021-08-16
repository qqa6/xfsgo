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
	"bytes"
	"encoding/hex"
	"errors"
	"github.com/sirupsen/logrus"
	"strings"
	"xfsgo/common"
	"xfsgo/common/rawencode"
	"xfsgo/lru"
	"xfsgo/storage/badger"
)

type Tree struct {
	db    *treeDb
	root  *TreeNode
	cache *lru.Cache
}

// NewTree creates a trie with an existing db and a root node.
// If the root exists and its format is correct, you need load the root node from the db
// and store the datas in a cache.
func NewTree(db *badger.Storage, root []byte) *Tree {
	t := &Tree{
		db: newTreeDb(db),
	}
	t.cache = lru.NewCache(2048)
	var zero [32]byte
	if root != nil && len(root) == int(32) && bytes.Compare(root, zero[:]) > common.Zero {
		t.root = t.mustLoadNode(root)
	}

	if err := FileExist(); err == nil {
		t.BackCommit()
	}

	return t
}

func (t *Tree) Put(k, v []byte) {
	if t.root == nil {
		t.root = newLeafNode(k, v)
		return
	}
	t.root = t.root.insert(t, k, v)
}

func (t *Tree) Checksum() []byte {
	if t.root == nil {
		return nil
	}
	return t.root.id
}

func (t *Tree) ChecksumHex() string {
	bs := t.Checksum()
	if bs == nil {
		return ""
	}
	return hex.EncodeToString(bs)
}

func (t *Tree) mustLoadLeft(node *TreeNode) *TreeNode {
	if node.leftNode != nil {
		return node.leftNode
	}
	ret := t.mustLoadNode(node.left)
	node.leftNode = ret
	return ret
}

func (t *Tree) mustLoadRight(node *TreeNode) *TreeNode {
	if node.rightNode != nil {
		return node.rightNode
	}
	ret := t.mustLoadNode(node.right)
	node.rightNode = ret
	return ret
}

func (t *Tree) mustLoadNode(id []byte) *TreeNode {
	n, err := t.loadNode(id)
	if err != nil {
		panic(err)
	}
	return n
}

func (t *Tree) loadLeft(n *TreeNode) (*TreeNode, error) {
	if n.leftNode != nil {
		return n.leftNode, nil
	}

	ret, err := t.loadNode(n.left)
	if err != nil {
		return nil, err
	}

	n.leftNode = ret

	return ret, nil
}
func (t *Tree) loadRight(n *TreeNode) (*TreeNode, error) {
	if n.rightNode != nil {
		return n.rightNode, nil
	}

	ret, err := t.loadNode(n.right)
	if err != nil {
		return nil, err
	}

	n.rightNode = ret

	return ret, nil
}

func (t *Tree) loadNode(id []byte) (*TreeNode, error) {
	var zero [32]byte
	if bytes.Compare(zero[:], id) == common.Zero {
		return nil, nil
	}
	var mId [32]byte
	copy(mId[:], id)
	if data, has := t.cache.Get(mId); has {
		tn := &TreeNode{}
		if err := rawencode.Decode(data, tn); err != nil {
			return nil, err
		}
		return tn, nil
	}
	tn, err := t.db.getTreeNodeByKey(append([]byte("tree:"), id...))
	if err != nil {
		return nil, err
	}
	// push cache
	buf, err := rawencode.Encode(tn)
	if err != nil {
		return nil, err
	}
	t.cache.Put(mId, buf)
	return tn, nil
}

// Get returns the value for key stored in the trie.
// The value bytes must not be modified by the caller.
func (t *Tree) Get(k []byte) ([]byte, bool) {
	if t.root == nil {
		return nil, false
	}
	return t.root.lookup(t, k)
}

func (t *Tree) Foreach(fn func(key []byte, value []byte)) {
	if t.root == nil {
		return
	}
	t.foreach(t.root, fn)
}

func (t *Tree) foreach(n *TreeNode, fn func(key []byte, value []byte)) {
	if n == nil {
		return
	}
	if n.isLeaf() {
		fn(n.key, n.value)
		return
	}
	t.foreach(t.mustLoadLeft(n), fn)
	t.foreach(t.mustLoadRight(n), fn)
}

func (t *Tree) BackCommit() error {
	if t.root == nil {
		return nil
	}
	journalObj, err := NewJournal()
	if err != nil {
		return err
	}
	scanner, err := journalObj.Replay()
	if err != nil {
		return err
	}
	root := common.Encode16Byte(t.Checksum())
	batch := t.db.newWriteBatch()
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, root) {
			result := journalObj.StrToMap(line, root)
			if result != nil {
				rootJournal := result["root"].(string)
				if rootJournal == root {
					keys := common.Decode16Byte(result["key"].(string))
					bs := common.Decode16Byte(result["bs"].(string))
					batch.Put(keys, bs)
				}
			} else {
				continue
			}

		}

	}
	if err = t.db.commitBatch(batch); err != nil {
		return err
	}
	return nil
}

func (t *Tree) Commit() error {
	if t.root == nil {
		return nil
	}
	journalObj, err := NewJournal()
	if err != nil {
		return err
	}
	batch := t.db.newWriteBatch()
	err = t.root.dfsCall(t, func(node *TreeNode) error {
		root := t.Checksum()
		_ = root
		key := append([]byte("tree:"), node.id...)

		bs, err := rawencode.Encode(node)
		if err != nil {
			return err
		}
		// return batch.Put(append([]byte("tree:"), node.id...), bs)
		if len(root) == 0 || len(key) == 0 || len(bs) == 0 {
			return errors.New("root or key or bs not null")
		}
		writeJournal := "root: " + common.Encode16Byte(root) + " key: " + common.Encode16Byte(key) + " bs: " + common.Encode16Byte(bs)
		journalObj.JouWrite(writeJournal)
		return batch.Put(key, bs)
	})
	if err != nil {
		return err
	}
	if err = t.db.commitBatch(batch); err != nil {
		return err
	}
	// delete journal Log
	if err = journalObj.DelWellDate(); err != nil {
		return err
	}
	return nil
}

func (t *Tree) Print(key []byte) error {
	tn, err := t.db.getTreeNodeByKey(key)
	if err != nil {
		return err
	}
	tn.rehash()
	logrus.Infof("tn Key: %s", string(tn.key))
	logrus.Infof("tn Key: %x", tn.id)
	logrus.Infof("tn val: %s", string(tn.value))
	return nil
}

func (t *Tree) PrintTree() {
	t.Foreach(func(key []byte, value []byte) {
		logrus.Infof("tree foreach: %s, id: %s", string(key), string(value))
	})
}
