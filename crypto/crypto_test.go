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

package crypto

import (
	"errors"
	"fmt"
	"testing"
	"xfsgo/common"
)

func TestPubKey2Addr(t *testing.T) {
	prv, err := GenPrvKey()
	if err != nil {
		t.Fatal(err)
	}
	addr := DefaultPubKey2Addr(prv.PublicKey)
	t.Logf("addr: %s\n", addr.String())
	keyEnc, err := PrivateKeyEncodeB64String(prv)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("privateKey: %s\n", keyEnc)
	got := common.StrB58ToAddress(addr.String())

	if !addr.Equals(got) {
		t.Fatal(fmt.Errorf("not equals"))
	}
	if !VerifyAddress(addr) {
		t.Fatal(fmt.Errorf("not verify"))
	}
}

func TestPubKeyEncode(t *testing.T) {
	key,err := GenPrvKey()
	if err != nil {
		t.Fatal(err)
	}
	enc := PubKeyEncode(key.PublicKey)
	t.Logf("len: %d, %x", len(enc), enc)
}

func TestVerifyAddress(t *testing.T) {
	key,err := GenPrvKey()
	if err != nil {
		t.Fatal(err)
	}
	p := key.PublicKey
	if p.Curve == nil || p.X == nil || p.Y == nil {
		t.Fatal(errors.New("nil err"))
	}
	xbs := p.X.Bytes()
	ybs := p.Y.Bytes()
	buf := make([]byte, len(xbs) + len(ybs))
	copy(buf,append(xbs,ybs...))
	t.Logf("x len: %d, %x", len(xbs), xbs)
	t.Logf("y len: %d, %x", len(ybs), ybs)
	t.Logf("buf len: %d, %x", len(buf), buf)
}
