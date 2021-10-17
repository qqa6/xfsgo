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
	"encoding/json"
	"fmt"
	"math/big"
	"testing"
	"xfsgo/assert"
	"xfsgo/common"
	"xfsgo/common/ahash"
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


func TestSign2(t *testing.T) {
	tx := &Transaction{
		Version: 0,
		To: common.StrB58ToAddress("kfye3Eh6Sxe4W1dFQGYRKzQCGCuU3JaUZ"),
		GasPrice: new(big.Int).SetInt64(120000),
		GasLimit: new(big.Int).SetInt64(120000),
		Data: []byte("1"),
		Nonce: 0,
		Value: new(big.Int),
		Time: 0,
		Signature: nil,
	}
	keyhex := "0101ecad21153a7b7b8c745fc91c74e620233ec090dae8730e04deca43ddbff53f24"
	keydata, err := hex.DecodeString(keyhex)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("key: %x\n", keydata)
	_, key, err := crypto.DecodePrivateKey(keydata)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("pubkey: %x\n", crypto.PubKeyEncode(key.PublicKey))
	txsignhash := tx.SignHash()
	t.Logf("txsignhash: %x\n", txsignhash)
	if err := tx.SignWithPrivateKey(key); err != nil {
		t.Fatal(err)
	}
	t.Logf("sign: %x\n", tx.Signature)
	pub, err := crypto.SigToPub(txsignhash[:], tx.Signature)
	if err != nil {
		t.Fatal(err)
	}
	pubdata := crypto.PubKeyEncode(*pub)
	t.Logf("sign2pub: %x\n", pubdata)
}


func TestSign(t *testing.T) {
	obj := &struct {
		To string `json:"to"`
		Value string `json:"value"`
		Gaslimit string `json:"gaslimit"`
		Gasprice string `json:"gasprice"`
		Signature *string `json:"signature"`
	}{
		To: "1",
		Value: "1",
		Gaslimit: "100",
		Gasprice: "100",
		Signature: nil,
	}
	jsondata, err := json.Marshal(obj)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("jsonstring: %s\n", string(jsondata))
	hashdata := ahash.SHA256(jsondata)
	keyhex := "0101ecad21153a7b7b8c745fc91c74e620233ec090dae8730e04deca43ddbff53f24"
	keydata, err := hex.DecodeString(keyhex)
	t.Logf("hashdata: %x\n", hashdata)
	if err != nil {
		t.Fatal(err)
	}
	_, key, err := crypto.DecodePrivateKey(keydata)
	if err != nil {
		t.Fatal(err)
	}
	sign,err := crypto.ECDSASign(hashdata, key)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("sign2: %x\n", sign)
}

func Test_abc(t *testing.T) {
	a := new(big.Int).SetInt64(100)
	var b *big.Int =  new(big.Int).SetInt64(10)
	if a.Cmp(b) > 0 {
		t.Fatal(fmt.Errorf("aaa"))
	}
}