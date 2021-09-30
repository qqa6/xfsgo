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
	"crypto/x509"
	"math/big"
	"sync"
	"xfsgo/common"
	"xfsgo/crypto"
	"xfsgo/storage/badger"
)

// Wallet represents a software wallet that has a default address derived from private key.
type Wallet struct {
	db          *keyStoreDB
	mu          sync.RWMutex
	cacheMu     sync.RWMutex
	defaultAddr common.Address
	GasPrice    *big.Int
	GasLimit    *big.Int
	cache       map[common.Address]*ecdsa.PrivateKey
}

// NewWallet constructs and returns a new Wallet instance with badger db.
func NewWallet(storage *badger.Storage, gas, gasPrice *big.Int) *Wallet {
	w := &Wallet{
		db:       newKeyStoreDB(storage),
		cache:    make(map[common.Address]*ecdsa.PrivateKey),
		GasLimit: gas,
		GasPrice: gasPrice,
	}
	w.defaultAddr, _ = w.db.GetDefaultAddress()
	return w
}

func (w *Wallet) SetGas(gas *big.Int) {
	w.GasLimit = gas
}

func (w *Wallet) GetGas() *big.Int {
	return w.GasLimit
}

func (w *Wallet) SetGasPrice(gasPrice *big.Int) {
	w.GasPrice = gasPrice
}

func (w *Wallet) GetGasPrice() *big.Int {
	return w.GasPrice
}

// AddByRandom constructs a new Wallet with a random number and retuens the its address.
func (w *Wallet) AddByRandom() (common.Address, error) {
	key, err := crypto.GenPrvKey()
	if err != nil {
		return noneAddress, err
	}
	return w.AddWallet(key)
}

func (w *Wallet) AddWallet(key *ecdsa.PrivateKey) (common.Address, error) {
	addr := crypto.DefaultPubKey2Addr(key.PublicKey)
	if err := w.db.PutPrivateKey(addr, key); err != nil {
		return noneAddress, err
	}
	if w.defaultAddr.Equals(noneAddress) {
		if err := w.SetDefault(addr); err != nil {
			return addr, nil
		}
	}
	return addr, nil
}

func (w *Wallet) All() map[common.Address]*ecdsa.PrivateKey {
	data := make(map[common.Address]*ecdsa.PrivateKey)
	w.db.Foreach(func(address common.Address, key *ecdsa.PrivateKey) {
		data[address] = key
	})
	return data
}

func (w *Wallet) GetKeyByAddress(address common.Address) (*ecdsa.PrivateKey, error) {
	w.cacheMu.RLock()
	if pk, has := w.cache[address]; has {
		w.cacheMu.RUnlock()
		return pk, nil
	}
	w.cacheMu.RUnlock()
	key, err := w.db.GetPrivateKey(address)
	if err != nil {
		return nil, err
	}
	w.cacheMu.Lock()
	w.cache[address] = key
	w.cacheMu.Unlock()
	return key, nil
}

func (w *Wallet) SetDefault(address common.Address) error {
	if address.Equals(w.defaultAddr) {
		return nil
	}
	err := w.db.SetDefaultAddress(address)
	if err != nil {
		return err
	}
	w.mu.Lock()
	w.defaultAddr = address
	w.mu.Unlock()
	return nil
}

func (w *Wallet) GetDefault() common.Address {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.defaultAddr
}

func (w *Wallet) Remove(address common.Address) error {
	if address.Equals(w.defaultAddr) {
		w.mu.Lock()
		if err := w.db.DelDefault(); err != nil {
			w.mu.Unlock()
			return err
		}
		w.defaultAddr = noneAddress
		w.mu.Unlock()
	}
	if err := w.db.RemoveAddress(address); err != nil {
		return err
	}
	w.cacheMu.Lock()
	delete(w.cache, address)
	w.cacheMu.Unlock()
	return nil
}

func (w *Wallet) Export(address common.Address) ([]byte, error) {
	key, err := w.GetKeyByAddress(address)
	if err != nil {
		return nil, err
	}
	return x509.MarshalECPrivateKey(key)
}

func (w *Wallet) Import(der []byte) (common.Address, error) {
	pKey, err := x509.ParseECPrivateKey(der)
	if err != nil {
		return noneAddress, err
	}
	return w.AddWallet(pKey)
}
