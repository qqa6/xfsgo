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
	"xfsgo/avlmerkle"
	"xfsgo/common"
	"xfsgo/common/ahash"
	"xfsgo/common/rawencode"
	"xfsgo/storage/badger"
	"xfsgo/uint256"

	"github.com/sirupsen/logrus"
)

//StateObj is an importment type which represents an xfs account that is being modified.
// The flow of usage is as follows:
// First, you need to obtain a StateObj object.
// Second, access and modify the balance of account through the object.
// Finally, call Commit method to write the modified merkleTree into a database.
type StateObj struct {
	merkleTree *avlmerkle.Tree
	address    common.Address //hash of address of the account
	balance    *big.Int
	nonce      uint64
	Extra      []byte
	gasPool    *big.Int
}

func (so *StateObj) Decode(data []byte) error {
	ext := struct {
		Address string `json:"address"`
		Balance string `json:"balance"`
		Nonce   uint64 `json:"nonce"`
	}{}
	if err := json.Unmarshal(data, &ext); err != nil {
		return err
	}
	balanceBs, err := hex.DecodeString(ext.Balance)
	if err != nil {
		balanceBs = []byte{}
	}
	balance := new(big.Int).SetBytes(balanceBs)
	so.address = common.StrB58ToAddress(ext.Address)
	so.balance = balance
	so.nonce = ext.Nonce
	return nil
}

func (so *StateObj) Encode() ([]byte, error) {
	balance := so.balance
	if balance == nil {
		balance = zeroBigN
	}
	balanceHex := hex.EncodeToString(balance.Bytes())
	ext := struct {
		Address string `json:"address"`
		Balance string `json:"balance"`
		Nonce   uint64 `json:"nonce"`
	}{Address: so.address.B58String(), Balance: balanceHex, Nonce: so.nonce}
	return json.Marshal(ext)
}

// NewStateObj creates an StateObj with accout address and tree
func NewStateObj(address common.Address, tree *avlmerkle.Tree) *StateObj {
	obj := &StateObj{
		address:    address,
		merkleTree: tree,
		gasPool:    new(big.Int),
	}
	return obj
}

// AddBalance adds amount to StateObj's balance.
// It is used to add funds to the destination account of a transfer.
func (so *StateObj) AddBalance(val *big.Int) {
	if val == nil || val.Sign() <= 0 {
		return
	}
	oldBalance := so.balance
	if oldBalance == nil {
		oldBalance = zeroBigN
	}
	newBalance := new(big.Int).Add(oldBalance, val)
	so.SetBalance(newBalance)
}

// SubBalance removes amount from StateObj's balance.
// It is used to remove funds from the origin account of a transfer.
func (so *StateObj) SubBalance(val *big.Int) {
	if val == nil || val.Sign() <= 0 {
		return
	}
	oldBalance := so.balance
	if oldBalance == nil {
		oldBalance = zeroBigN
	}
	newBalance := oldBalance.Sub(oldBalance, val)
	so.SetBalance(newBalance)
}

func (so *StateObj) SetBalance(val *big.Int) {
	if val == nil || val.Sign() <= 0 {
		return
	}
	so.balance = val
}

func (so *StateObj) GetBalance() *big.Int {
	return so.balance
}

func (so *StateObj) GetAddress() common.Address {
	return so.address
}

func (so *StateObj) SetGasLimit(gasLimit *big.Int) {
	so.gasPool = new(big.Int).Set(gasLimit)
}

func (so *StateObj) SubGas(gas, price *big.Int) error {
	if so.gasPool.Cmp(gas) < 0 {
		return fmt.Errorf("GasLimit error. Max %s, transaction would take it to %s", so.gasPool, gas)
	}

	so.gasPool.Sub(so.gasPool, gas)

	rGas := new(big.Int).Set(gas)
	rGas.Mul(rGas, price)

	return nil
}

func (so *StateObj) SetNonce(nonce uint64) {
	so.nonce = nonce
}
func (so *StateObj) AddNonce(nonce uint64) {
	so.nonce += nonce
}
func (so *StateObj) SubNonce(nonce uint64) {
	so.nonce -= nonce
}
func (so *StateObj) GetNonce() uint64 {
	return so.nonce
}

func (so *StateObj) Update() {
	objRaw, _ := rawencode.Encode(so)
	hash := ahash.SHA256(so.address[:])
	so.merkleTree.Put(hash, objRaw)
}

type StateTree struct {
	root       []byte
	treeDB     *badger.Storage
	merkleTree *avlmerkle.Tree
	objs       map[common.Address]*StateObj
}

func NewStateTree(db *badger.Storage, root []byte) *StateTree {
	st := &StateTree{
		root:   root,
		treeDB: db,
		objs:   make(map[common.Address]*StateObj),
	}
	st.merkleTree = avlmerkle.NewTree(st.treeDB, root)
	return st
}

func (st *StateTree) HashAccount(addr common.Address) bool {
	return st.GetStateObj(addr) != nil
}

func (st *StateTree) GetBalance(addr common.Address) *big.Int {
	obj := st.GetStateObj(addr)
	if obj != nil {
		if obj.balance == nil {
			return zeroBigN
		}
		return obj.balance
	}
	return zeroBigN
}

func (st *StateTree) AddBalance(addr common.Address, val *big.Int) {
	obj := st.GetOrNewStateObj(addr)
	if obj != nil {
		obj.AddBalance(val)
	}
}
func (st *StateTree) GetNonce(addr common.Address) uint64 {
	obj := st.GetStateObj(addr)
	if obj != nil {
		return obj.nonce
	}
	return 0
}

func (st *StateTree) AddNonce(addr common.Address, val uint64) {
	obj := st.GetOrNewStateObj(addr)
	if obj != nil {
		obj.AddNonce(val)
	}
}

func (st *StateTree) GetStateObj(addr common.Address) *StateObj {
	if st.objs[addr] != nil {
		return st.objs[addr]
	}
	hash := ahash.SHA256(addr.Bytes())
	if val, has := st.merkleTree.Get(hash); has {
		obj := &StateObj{}
		if err := rawencode.Decode(val, obj); err != nil {
			return nil
		}
		obj.merkleTree = st.merkleTree
		st.objs[addr] = obj
		return obj
	}
	return nil
}

func (st *StateTree) newStateObj(address common.Address) *StateObj {
	obj := NewStateObj(address, st.merkleTree)
	st.objs[obj.address] = obj
	return obj
}
func (st *StateTree) CreateAccount(addr common.Address) *StateObj {
	old := st.GetStateObj(addr)
	add := st.newStateObj(addr)
	if old != nil {
		add.balance = old.balance
	}
	return add
}

func (st *StateTree) GetOrNewStateObj(addr common.Address) *StateObj {
	stateObj := st.GetStateObj(addr)
	if stateObj == nil {
		stateObj = st.CreateAccount(addr)
	}
	return stateObj
}

func (st *StateTree) Root() []byte {
	return st.merkleTree.Checksum()
}

func (st *StateTree) RootHex() string {
	return st.merkleTree.ChecksumHex()
}
func (st *StateTree) RootUint256() uint256.UInt256 {
	r := st.Root()
	return *uint256.NewUInt256BS(r)
}

func (st *StateTree) UpdateAll() {
	for _, v := range st.objs {
		v.Update()
	}
}

func (st *StateTree) Commit() error {
	return st.merkleTree.Commit()
}

func (st *StateTree) Print() {
	st.merkleTree.Foreach(func(key []byte, value []byte) {
		//address := common.Bytes2Address(key)
		obj := &StateObj{}
		if err := rawencode.Decode(value, obj); err != nil {
			return
		}
		balance := obj.GetBalance()
		logrus.Infof("address: %x, balance: %d", key, balance)
	})
}
