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
	"fmt"
	"math/big"
	"testing"
	"xfsgo/assert"
	"xfsgo/crypto"
)

func TestTransaction_VerifySignature(t *testing.T) {
	prvKeyFrom, err := crypto.GenPrvKey()
	assert.Error(t, err)
	prvKeyTo, err := crypto.GenPrvKey()
	assert.Error(t, err)
	fromAddr := crypto.DefaultPubKey2Addr(prvKeyFrom.PublicKey)
	assert.VerifyAddress(t, fromAddr)
	toAddr := crypto.DefaultPubKey2Addr(prvKeyTo.PublicKey)
	assert.VerifyAddress(t, toAddr)
	tx := &Transaction{
		To:    toAddr,
		Value: new(big.Int).SetInt64(100),
	}
	if err := tx.SignWithPrivateKey(prvKeyFrom); err != nil {
		t.Fatal(err)
	}
	if !tx.VerifySignature() {
		t.Fatal(fmt.Errorf("tx not verify"))
	}
	gotAddr, err := tx.FromAddr()
	assert.VerifyAddress(t, gotAddr)
	assert.AddressEq(t, gotAddr, fromAddr)
}
