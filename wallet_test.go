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
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/hex"
	"testing"
	"xfsgo/assert"
	"xfsgo/common"
	"xfsgo/common/urlsafeb64"
	"xfsgo/crypto"
	"xfsgo/storage/badger"
)

func randomKey(t *testing.T) *ecdsa.PrivateKey {
	pk, err := crypto.GenPrvKey()
	if err != nil {
		t.Fatal(err)
	}
	return pk
}
func TestWallets_AddByRandom(t *testing.T) {
	storage := badger.New("./d3/keys")
	defer func() {
		if err := storage.Close(); err != nil {
			t.Fatalf("Sotrage close errors: %s", err)
		}
	}()
	ws := NewWallet(storage, nil,nil)
	addr, err := ws.AddByRandom()
	if err != nil {
		t.Fatal(err)
	}
	pk, err := ws.Export(addr)
	t.Logf("addr: %s, private: %s", addr.B58String(), urlsafeb64.Encode(pk))
}

func TestWallet_AddWallet(t *testing.T) {
	storage := badger.New("./d3/keys")
	defer func() {
		if err := storage.Close(); err != nil {
			t.Fatalf("Sotrage close errors: %s", err)
		}
	}()
	key := randomKey(t)
	ws := NewWallet(storage,nil,nil)
	gotAddr, err := ws.AddWallet(key)
	if err != nil {
		t.Fatal(err)
	}
	wantAddr := crypto.DefaultPubKey2Addr(key.PublicKey)
	if !gotAddr.Equals(wantAddr) {
		t.Fatalf("got %s, want %s", gotAddr.B58String(), wantAddr.B58String())
	}
}

func TestWallet_Export(t *testing.T) {
	storage := badger.New("./d3/keys")
	defer func() {
		if err := storage.Close(); err != nil {
			t.Fatalf("Sotrage close errors: %s", err)
		}
	}()
	key := randomKey(t)
	ws := NewWallet(storage,nil,nil)
	addr, err := ws.AddWallet(key)
	if err != nil {
		t.Fatal(err)
	}
	got, err := ws.Export(addr)
	if err != nil {
		t.Fatal(err)
	}
	want, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	gotHex := hex.EncodeToString(got)
	wantHex := hex.EncodeToString(want)
	if bytes.Compare(got, want) != common.Zero {
		t.Fatalf("got %s, want %s", gotHex, wantHex)
	}
}

func TestWallet_All(t *testing.T) {
	storage := badger.New("./d3/keys")
	defer func() {
		if err := storage.Close(); err != nil {
			t.Fatalf("Sotrage close errors: %s", err)
		}
	}()
	wantAll := make(map[common.Address]*ecdsa.PrivateKey)
	key1 := randomKey(t)
	key2 := randomKey(t)
	key3 := randomKey(t)
	ws := NewWallet(storage,nil,nil)
	addr1, err := ws.AddWallet(key1)
	if err != nil {
		t.Fatal(err)
	}
	wantAll[addr1] = key1
	addr2, err := ws.AddWallet(key2)
	if err != nil {
		t.Fatal(err)
	}
	wantAll[addr2] = key2
	addr3, err := ws.AddWallet(key3)
	if err != nil {
		t.Fatal(err)
	}
	wantAll[addr3] = key3
	gotAll := ws.All()
	if len(gotAll) < len(wantAll) {
		t.Fatalf("got len: %d, want len: >= %d", len(gotAll), len(wantAll))
	}
	gotKey1, has1 := gotAll[addr1]
	if !has1 {
		t.Fatalf("not found got all by address: %s", addr1.B58String())
	}
	wantKey1 := wantAll[addr1]
	assert.PrivateKeyEqual(t, gotKey1, wantKey1)
	gotKey2, has2 := gotAll[addr2]
	if !has2 {
		t.Fatalf("not found got all by address: %s", addr2.B58String())
	}
	wantKey2 := wantAll[addr2]
	assert.PrivateKeyEqual(t, gotKey2, wantKey2)
	gotKey3, has3 := gotAll[addr3]
	if !has3 {
		t.Fatalf("not found got all by address: %s", addr3.B58String())
	}
	wantKey3 := wantAll[addr3]
	assert.PrivateKeyEqual(t, gotKey3, wantKey3)
}
