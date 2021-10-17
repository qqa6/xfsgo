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
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"sort"
	"strconv"
	"time"
	"xfsgo/common"
	"xfsgo/common/ahash"
	"xfsgo/common/rawencode"
	"xfsgo/crypto"

	"github.com/sirupsen/logrus"
)

// var defaultGasPrice = new(big.Int).SetUint64(1)    //150000000000

// Transaction type.
type Transaction struct {
	Version   uint32         `json:"version"`
	To        common.Address `json:"to"`
	GasPrice  *big.Int       `json:"gas_price"`
	GasLimit  *big.Int       `json:"gas_limit"`
	Data      []byte         `json:"data"`
	Nonce     uint64         `json:"nonce"`
	Value     *big.Int       `json:"value"`
	Timestamp      uint64         `json:"timestamp"`
	Signature []byte         `json:"signature"`
}

type StdTransaction struct {
	Version   uint32         `json:"version"`
	To        common.Address `json:"to"`
	GasPrice  *big.Int       `json:"gas_price"`
	GasLimit  *big.Int       `json:"gas_limit"`
	Data      []byte         `json:"data"`
	Nonce     uint64         `json:"nonce"`
	Value     *big.Int       `json:"value"`
	Timestamp      uint64         `json:"timestamp"`
	Signature []byte         `json:"signature"`
}

func NewTransaction(to common.Address, gasLimit, gasPrice *big.Int, value *big.Int) *Transaction {
	now := time.Now().Unix()
	result := &Transaction{
		Version:  version0,
		To:       to,
		GasLimit: gasLimit,
		GasPrice: gasPrice,
		Value:    value,
		Timestamp:  uint64(now),
	}
	return result
}

func NewTransactionByStd(tx *StdTransaction) *Transaction {
	result := &Transaction{
		Version:  tx.Version,
		To:       common.Address{},
		GasPrice: new(big.Int),
		GasLimit: new(big.Int),
		Data: tx.Data,
		Nonce: tx.Nonce,
		Value:    new(big.Int),
		Timestamp:     tx.Timestamp,
		Signature: tx.Signature,
	}
	if !bytes.Equal(tx.To[:], common.ZeroAddr[:]) {
		result.To = tx.To
	}
	if tx.GasPrice != nil {
		result.GasPrice.Set(tx.GasPrice)
	}
	if tx.GasLimit != nil {
		result.GasLimit.Set(tx.GasLimit)
	}
	if tx.Value != nil {
		result.Value.Set(tx.Value)
	}
	return result
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

func sortAndEncodeMap(data map[string]string) string {
	mapkeys := make([]string, 0)
	for k, _ := range data {
		mapkeys = append(mapkeys, k)
	}
	sort.Strings(mapkeys)
	strbuf := ""
	for i, key := range mapkeys {
		val := data[key]
		if val == "" {
			continue
		}
		strbuf += fmt.Sprintf("%s=%s", key, val)
		if i < len(mapkeys) -1 {
			strbuf += "&"
		}
	}
	return strbuf
}

func (t *Transaction) SignHash() common.Hash {
	//nt := t.copyTrim()
	tmp := map[string]string{
		"version": strconv.FormatInt(int64(t.Version), 10),
		"to": t.To.String(),
		"gas_price": hex.EncodeToString(t.GasPrice.Bytes()),
		"gas_limit": hex.EncodeToString(t.GasLimit.Bytes()),
		"data": hex.EncodeToString(t.Data),
		"nonce": strconv.Itoa(int(t.Nonce)),
		"value": hex.EncodeToString(t.Value.Bytes()),
		"timestamp": strconv.FormatInt(int64(t.Timestamp), 10),
	}
	enc := sortAndEncodeMap(tmp)
	if enc == "" {
		return common.Hash{}
	}
	logrus.Infof("signhash, enc: %s", enc)
	return common.Bytes2Hash(ahash.SHA256([]byte(enc)))
}

func (t *Transaction) Cost() *big.Int {
	i := big.NewInt(0)
	i.Mul(t.GasLimit, t.GasPrice)
	i.Add(i, t.Value)
	return i
}

func (t *Transaction) SignWithPrivateKey(key *ecdsa.PrivateKey) error {
	hash := t.SignHash()
	sig, err := crypto.ECDSASign(hash.Bytes(), key)
	if err != nil {
		return err
	}
	t.Signature = sig
	return nil
}

func (t *Transaction) VerifySignature() bool {
	if _, err := t.publicKey(); err != nil {
		logrus.Warnf("Failed verify signature: %s", err)
		return false
	}
	return true
}

func (t *Transaction) publicKey() (*ecdsa.PublicKey, error) {
	hash := t.SignHash()
	return crypto.SigToPub(hash[:], t.Signature)
}

//FromAddr checks the validation of public key from the signature in the transaction.
//if right, returns the address calculated by this public key.
func (t *Transaction) FromAddr() (common.Address, error) {
	pub, err := t.publicKey()
	if err != nil {
		logrus.Warnf("Failed parse from addr by signature: %s", err)
		return common.Address{}, err
	}
	addr := crypto.DefaultPubKey2Addr(*pub)
	return addr, nil
}

func (t *Transaction) String() string {
	jsondata, err := json.Marshal(t)
	if err != nil {
		panic(err)
	}
	return string(jsondata)
}