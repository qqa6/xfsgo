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

package p2p

import (
	"testing"
	"xfsgo/crypto"
	"xfsgo/p2p/discover"
)

func TestHello(t *testing.T) {
	key, err := crypto.GenPrvKey()
	if err != nil {
		t.Fatal(err)
	}
	pubKey := key.PublicKey
	id := discover.PubKey2NodeId(pubKey)
	hello := &helloRequestMsg{
		version: 1,
		id:      id,
	}
	data := hello.marshal()
	t.Logf("data: %v\n", data)
	got := new(helloRequestMsg)
	got.unmarshal(data)
	gothash := got.id
	t.Logf("gothash: %v\n", gothash)
	t.Logf("wanthash: %v\n", id)
}

func TestHello2(t *testing.T) {
	data := []byte{1, 0, 32, 0, 0, 0}
	t.Logf("data: %v\n", data)
	n := uint32(data[2]) | uint32(data[3])<<8 | uint32(data[4])<<16 | uint32(data[5])<<24
	t.Logf("n: %v\n", n)
}
