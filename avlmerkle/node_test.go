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
	"testing"
	"xfsgo/common/ahash"
	"xfsgo/common/rawencode"
)

func TestTreeNode_Decode(t *testing.T) {
	tn := newLeafNode([]byte("hello"), []byte("world"))
	tn.depth = 1
	tn.left = ahash.SHA256([]byte("abc"))
	tn.right = ahash.SHA256([]byte("def"))
	bs, err := rawencode.Encode(tn)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("bytes: %x\n", bs)
	var nn = &TreeNode{}
	if err := rawencode.Decode(bs, nn); err != nil {
		t.Fatal(err)
	}
	t.Logf("---")
	_ = nn
}
