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
	"encoding/hex"
	"github.com/magiconair/properties/assert"
	"github.com/sirupsen/logrus"
	"testing"
	"xfsgo/storage/badger"
)

func TestBuildTree(t *testing.T) {
	var db *badger.Storage = nil
	tr := NewTree(db, nil)
	tr.Put([]byte("a"), []byte("b"))
	tr.Put([]byte("c"), []byte("d"))
	tr.Put([]byte("d"), []byte("e"))
	tr.Put([]byte("e"), []byte("f"))
	tr.Put([]byte("g"), []byte("h"))
	tr.Put([]byte("i"), []byte("j"))
	root := tr.root
	_ = root
}
func printTree(t *testing.T, root *TreeNode) {
	if root == nil {
		t.Log("empty")
		return
	}
	t.Logf("key: %s\n", string(root.key))
	t.Logf("depth: %d\n", root.depth)
	t.Logf("v: %s\n", string(root.value))
	t.Logf("----------------")
	if left := root.leftNode; left != nil {
		printTree(t, left)
	}
	if right := root.rightNode; right != nil {
		printTree(t, right)
	}
}

func TestBuildTree2(t *testing.T) {
	var db *badger.Storage = nil
	tr := NewTree(db, nil)
	tr.Put([]byte("a"), []byte("b"))
	tr.Put([]byte("c"), []byte("d"))
	tr.Put([]byte("e"), []byte("f"))
	root := tr.root
	printTree(t, root)
	_ = root
}

func TestTree_Get(t *testing.T) {
	var db *badger.Storage = nil
	tr := NewTree(db, nil)
	tr.Put([]byte("a"), []byte("b"))
	tr.Put([]byte("c"), []byte("d"))
	tr.Put([]byte("d"), []byte("e"))
	tr.Put([]byte("e"), []byte("f"))
	tr.Put([]byte("g"), []byte("h"))
	tr.Put([]byte("i"), []byte("j"))
	findKey := []byte("a")
	wantValue := []byte("b")
	gotValue, has := tr.Get(findKey)
	if !has {
		t.Fatal("not found")
	}
	t.Logf("found key: %s, value: %s\n", string(findKey), string(gotValue))
	assert.Equal(t, gotValue, wantValue)
}

func TestTree_Commit(t *testing.T) {
	var db = badger.New("./d0")
	defer func() {
		if err := db.Close(); err != nil {
			t.Fatal(err)
		}
	}()
	tr := NewTree(db, nil)
	tr.Put([]byte("a"), []byte("b"))
	tr.Put([]byte("c"), []byte("d"))
	tr.PrintTree()
	//tr.Put([]byte("d"),[]byte("e"))
	if err := tr.Commit(); err != nil {
		t.Fatal(err)
	}

	logrus.Infof("tree checksum id: %s", tr.ChecksumHex())
}

func TestTree_Foreach(t *testing.T) {
	rootHex := "8c804011a57df85c5db34ed2151c1439d8d307bfa49008792c2d6929fb53c356"
	root, err := hex.DecodeString(rootHex)
	if err != nil {
		t.Fatal(err)
	}
	var db = badger.New("./d0")
	defer func() {
		if err := db.Close(); err != nil {
			t.Fatal(err)
		}
	}()
	tr := NewTree(db, root)
	tr.Put([]byte("a"), []byte("c"))
	tr.PrintTree()
	if err = tr.Commit(); err != nil {
		t.Fatal(err)
	}
	gotHex := tr.ChecksumHex()
	t.Logf("root hash: %s\n", gotHex)
	if gotHex == rootHex {
		t.Errorf("got hex: %s, want hash: %s, is no difference\n", gotHex, rootHex)
		t.Fail()
	}
}

func TestTree_Foreach2(t *testing.T) {
	rootHex := "66103bdf1a7666338e664b99c5a3349fadeaf465060d00cf41e1d400999031b2"
	//rootHex := "381867cdc2708ce07b9b01ffe6b97cbc42febf5dab4e40583b158b7293eca401"
	root, err := hex.DecodeString(rootHex)
	if err != nil {
		t.Fatal(err)
	}
	var db = badger.New("./d0")
	tr := NewTree(db, root)
	defer func() {
		if err := db.Close(); err != nil {
			t.Fatal(err)
		}
	}()
	tr.Foreach(func(key []byte, value []byte) {
		logrus.Infof("tree foreach: %s, id: %s", string(key), string(value))
	})
}
