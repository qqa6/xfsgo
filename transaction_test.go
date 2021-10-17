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
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
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
		Timestamp: 0,
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
	txpub, err := tx.publicKey()
	if err != nil {
		t.Fatal(err)
	}
	txpubdata := crypto.PubKeyEncode(*txpub)
	t.Logf("txpub: %x\n", txpubdata)
	fromaddr, err := tx.FromAddr()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("txfrom: %s\n", fromaddr.B58String())
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

type StringRawTransaction struct {
	Version string `json:"version"`
	To string `json:"to"`
	Value string `json:"value"`
	Data string `json:"data"`
	GasLimit string `json:"gas_limit"`
	GasPrice string `json:"gas_price"`
	Signature string `json:"signature"`
	Nonce     string `json:"nonce"`
	Timestamp string `json:"timestamp"`
}
func CoverTransaction(r *StringRawTransaction) (*Transaction,error) {
	version, err := strconv.ParseInt(r.Version, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("failed to parse version: %s", err)
	}
	signature := common.Hex2bytes(r.Signature)
	if signature == nil || len(signature) < 1 {
		return nil, fmt.Errorf("failed to parse signature: %s", err)
	}
	toaddr := common.ZeroAddr
	if r.To != "" {
		toaddr = common.StrB58ToAddress(r.To)
		if !crypto.VerifyAddress(toaddr) {
			return nil, fmt.Errorf("failed to verify 'to' address: %s", r.To)
		}
	}else if r.Data == "" {
		return nil, fmt.Errorf("failed to parse 'to' address")
	}
	gasprice, ok := new(big.Int).SetString(r.GasPrice, 16)
	if !ok {
		return nil, fmt.Errorf("failed to parse gasprice")
	}
	gaslimit, ok := new(big.Int).SetString(r.GasLimit, 16)
	if !ok {
		return nil, fmt.Errorf("failed to parse gasprice")
	}
	data := common.Hex2bytes(r.Data)
	nonce, err := strconv.ParseInt(r.Nonce, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse nonce: %s", err)
	}
	value, ok := new(big.Int).SetString(r.Value, 16)
	if !ok {
		return nil, fmt.Errorf("failed to parse value")
	}
	timestamp, err := strconv.ParseInt(r.Timestamp, 10, 64)
	if !ok {
		return nil, fmt.Errorf("failed to parse timestamp")
	}
	return NewTransactionByStd(&StdTransaction{
		Version: uint32(version),
		To: toaddr,
		GasPrice: gasprice,
		GasLimit: gaslimit,
		Data: data,
		Nonce: uint64(nonce),
		Value: value,
		Timestamp: uint64(timestamp),
		Signature: signature,
	}), nil
}
func Test_abc(t *testing.T) {
	privkeystr := "0x0101ecad21153a7b7b8c745fc91c74e620233ec090dae8730e04deca43ddbff53f24"
	t.Logf("private key: %s", privkeystr)
	privkeybytes := common.Hex2bytes(privkeystr)
	_,privateKey,err := crypto.DecodePrivateKey(privkeybytes)
	if err != nil {
		t.Fatal(err)
	}
	publickey := privateKey.PublicKey
	publickeybytes := crypto.PubKeyEncode(publickey)
	t.Logf("public key: %x", publickeybytes)
	fromaddress := crypto.DefaultPubKey2Addr(publickey)
	t.Logf("from address: %s", fromaddress.B58String())
	signhex := "eaba0937d18cbf74e9143e1c15a066fffd16b1ab0a89be76010360bc3238095e3e8b9a5894b81b09307ed9ba18467c9a9cce7dfe9b06bdf4d503df016090a9e901"
	wantsign := common.Hex2bytes(signhex)
	tx, err := CoverTransaction(&StringRawTransaction{
		Version: "0",
		To: "beJLCrggTVQEASawNta4QFbGkLN51r3qj",
		GasPrice: "10",
		GasLimit: "10",
		Nonce: "0",
		Value: "10",
		Timestamp: "1634438019",
		Signature: signhex,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err = tx.SignWithPrivateKey(privateKey); err != nil {
		t.Fatal(err)
	}
	realsign := tx.Signature
	if !bytes.Equal(wantsign, realsign) {
		t.Fatal(fmt.Errorf("want=%x, got: %x", wantsign, realsign))
	}
	gotpubkey, err := tx.publicKey()
	if err != nil {
		t.Fatal(err)
	}
	gotpubkeybytes := crypto.PubKeyEncode(*gotpubkey)
	if !bytes.Equal(publickeybytes, gotpubkeybytes) {
		t.Fatal(fmt.Errorf("want=%x, got: %x", publickeybytes, gotpubkeybytes))
	}
	gotaddr, err := tx.FromAddr()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(fromaddress[:], gotaddr[:]) {
		t.Fatal(fmt.Errorf("want=%x, got: %x", fromaddress, gotaddr))
	}
	//signder, err := hex.DecodeString(signhex)
	//if err != nil {
	//	t.Fatal(err)
	//}
}