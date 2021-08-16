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
	"encoding/hex"
	"math/big"
	"testing"
	"xfsgo/common"
	"xfsgo/common/ahash"
	"xfsgo/crypto"
	"xfsgo/storage/badger"
)

func caddr() common.Address {
	prvKey, err := crypto.GenPrvKey()
	if err != nil {
		return common.Address{}
	}
	return crypto.DefaultPubKey2Addr(prvKey.PublicKey)
}
func TestStateTree_AddBalance(t *testing.T) {
	stateDb := badger.New("./data0/state")
	defer func() {
		if err := stateDb.Close(); err != nil {
			t.Fatal(err)
		}
	}()
	st := NewStateTree(stateDb, nil)
	a1 := caddr()
	a2 := caddr()
	a3 := caddr()
	t.Logf("addr1: %s", a1.B58String())
	t.Logf("addr2: %s", a2.B58String())
	t.Logf("addr3: %s", a3.B58String())
	st.AddBalance(a1, big.NewInt(1))

	st.AddBalance(a2, big.NewInt(2))

	st.AddBalance(a3, big.NewInt(3))

	st.UpdateAll()
	if err := st.Commit(); err != nil {
		t.Fatal(err)
	}
	st.Print()
	t.Logf("-------------\n")
	st2 := NewStateTree(stateDb, st.Root())
	st2.AddBalance(a1, new(big.Int).SetInt64(8))
	st2.UpdateAll()
	if err := st.Commit(); err != nil {
		t.Fatal(err)
	}
	st2.Print()

}

func TestEn(t *testing.T) {
	stateDb := badger.New("./data0/state")
	defer func() {
		if err := stateDb.Close(); err != nil {
			t.Fatal(err)
		}
	}()
	//var rootHex = "c772eccb80bcdaa2af0628604642cb440d833b45e1c9f8fd2e9f06053f0cca08"
	var rootHex = "2cdcc452547a6057d0357b7b83694e916372003f9e063a037baadd710304d6dd"
	rootBytes, err := hex.DecodeString(rootHex)
	if err != nil {
		t.Fatal(err)
	}
	st := NewStateTree(stateDb, rootBytes)
	st.AddBalance(common.StrB58ToAddress("14dAm9rMFzf6ecQBjwZ6S24Z3QxAz9AvLo"), big.NewInt(1))
	st.UpdateAll()
	if err := st.Commit(); err != nil {
		t.Fatal(err)
	}
	st.Print()
	hash := st.RootHex()

	t.Logf("root: %s\n", hash)
}

func TestEn2(t *testing.T) {
	stateDb := badger.New("./data0/state")
	defer func() {
		if err := stateDb.Close(); err != nil {
			t.Fatal(err)
		}
	}()
	var rootHex = "05538353386d341b4ff07493614b1f49ab54ee00518b69fce3890289e841580d"
	rootBytes, err := hex.DecodeString(rootHex)
	if err != nil {
		t.Fatal(err)
	}
	st := NewStateTree(stateDb, rootBytes)
	st.Print()
}

func TestEn3(t *testing.T) {
	hash := ahash.SHA256(append([]byte("hello"), []byte("2619202")...))
	t.Logf("balance: %x\n", hash)
}
