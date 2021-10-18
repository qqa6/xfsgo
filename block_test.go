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
	"testing"
	"xfsgo/common"
	"xfsgo/common/rawencode"
)

func TestBlock_Hash(t *testing.T) {
	block := NewBlock(&BlockHeader{
		Height:        0,
		Version:       0,
		HashPrevBlock: emptyHash,
		Timestamp:     0,
		Coinbase:      noneAddress,
		Nonce:         0,
	}, nil, nil, nil)
	t.Logf("%s\n", block)
	blockHash := block.Hash()
	want := common.Hex2Hash("0xd3706b6c55aeee64baef8b01f85a05ffa6830e77a94558813baaef3d67e5cf16")
	if bytes.Compare(blockHash.Bytes(), want.Bytes()) != common.Zero {
		t.Fatalf("not match,  got: %x, want: %x\n", blockHash, want)
	}
}

func TestBlock_Decode(t *testing.T) {
	block := NewBlock(&BlockHeader{
		Height:        1,
		Version:       2,
		HashPrevBlock: common.Bytes2Hash([]byte{0xff, 0xff}),
		Timestamp:     3,
		Coinbase:      common.Bytes2Address([]byte{0xff, 0xff}),
		Bits:          4,
		Nonce:         0,
	}, nil, nil, nil)
	wantHash := block.Hash()
	t.Logf("block: %s\n", block)
	blockData, err := rawencode.Encode(block)
	if err != nil {
		t.Fatal(err)
	}
	got := &Block{}
	err = rawencode.Decode(blockData, got)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("got: %s\n", got)
	gotHash := got.Hash()
	if bytes.Compare(gotHash.Bytes(), wantHash.Bytes()) != common.Zero {
		t.Fatalf("got: %x, want: %x\n", gotHash, wantHash)
	}
}
