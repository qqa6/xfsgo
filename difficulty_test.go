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
	"math/big"
	"testing"
	"xfsgo/common"

	"github.com/magiconair/properties/assert"
)

func TestBigByZip(t *testing.T) {
	bs := []byte{
		0, 0, 0, 0xff, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0,
	}
	bigN := new(big.Int).SetBytes(bs)
	bits := BigByZip(bigN)
	bn := BitsUnzip(bits)
	assert.Equal(t, len(bn.Bytes()), 32-3)
}

func TestA(t *testing.T) {
	a := new(big.Int).Lsh(big0xff, 256-(8*2))
	aBs := a.Bytes()
	t.Logf("a len: %d, %x", len(aBs), aBs)
	b := new(big.Int).Lsh(big0xff, 256-(8*1))
	bBs := b.Bytes()
	t.Logf("b len: %d, %x", len(bBs), bBs)
	if bytes.Compare(aBs, bBs) > common.Zero {
		t.Fatalf("this is not the expected value")
	}
}

func TestB(t *testing.T) {
	target := BitsUnzip(1069547545)
	targetBs := target.Bytes()
	t.Logf("a len: %d, %x", len(targetBs), targetBs)

	//if bytes.Compare(aBs, targetBs) != 0 {
	//	t.Fatalf("this is not the expected value")
	//}
}
