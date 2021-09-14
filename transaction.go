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
	"crypto/ecdsa"
	"encoding/json"
	"math/big"
	"xfsgo/common"
	"xfsgo/common/ahash"
	"xfsgo/common/rawencode"
	"xfsgo/crypto"

	"github.com/sirupsen/logrus"
)

// Transaction type.
type Transaction struct {
	Version   uint32         `json:"version"`
	To        common.Address `json:"to"`
	GasPrice  *big.Int       `json:"gas_price"`
	GasLimit  uint64         `json:"gas_limit"`
	Nonce     uint64         `json:"nonce"`
	Value     *big.Int       `json:"value"`
	Signature []byte         `json:"signature"`
}

func NewTransaction(to common.Address, value *big.Int) *Transaction {
	return &Transaction{
		Version: version0,
		To:      to,
		Value:   value,
	}
}

func (t *Transaction) Encode() ([]byte, error) {
	return json.Marshal(t)
}

func (t *Transaction) Decode(data []byte) error {
	return json.Unmarshal(data, t)
}

func (t *Transaction) Hash() common.Hash {
	bs, err := rawencode.Encode(t)
	if err != nil {
		return common.ZeroHash
	}
	return common.Bytes2Hash(ahash.SHA256(bs))
}

func (t *Transaction) clone() *Transaction {
	p := *t
	return &p
}
func (t *Transaction) copyTrim() *Transaction {
	nt := t.clone()
	nt.Signature = nil
	return nt
}

// SigHash returns the hash to be signed by the sender.
// It does not uniquely identify the transaction.
func (t *Transaction) SignHash() common.Hash {
	nt := t.copyTrim()
	bs, err := rawencode.Encode(nt)
	if err != nil {
		return common.ZeroHash
	}
	return common.Bytes2Hash(ahash.SHA256(bs))
}

func (t *Transaction) Cost() *big.Int {
	return t.Value
}

func (t *Transaction) SignWithPrivateKey(key *ecdsa.PrivateKey) error {
	hash := t.SignHash()
	logrus.Infof("sign hash: %s", hash.Hex())
	pub := key.PublicKey
	logrus.Infof("sign pubx: %x", pub.X.Bytes())
	logrus.Infof("from puby: %x", pub.Y.Bytes())
	sig, err := crypto.ECDSASign(hash.Bytes(), key)
	if err != nil {
		return err
	}
	t.Signature = sig
	return nil
}

//
func (t *Transaction) VerifySignature() bool {
	hash := t.SignHash()
	logrus.Infof("verify sign hash: %s", hash.Hex())
	return crypto.VerifySignature(hash.Bytes(), t.Signature)
}

//FromAddr checks the validation of public key from the signature in the transaction.
//if right, returns the address calculated by this public key.
func (t *Transaction) FromAddr() (common.Address, error) {
	pub, err := crypto.ParsePubKeyFromSignature(t.Signature)
	if err != nil {
		return common.Bytes2Address([]byte{}), err
	}
	addr := crypto.DefaultPubKey2Addr(pub)
	return addr, nil
}
