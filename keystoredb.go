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
	"xfsgo/common"
	"xfsgo/storage/badger"
)

var (
	addrKeyPre        = []byte("addr:")
	defaultAddressKey = []byte("default")
)

type keyStoreDB struct {
	storage *badger.Storage
}

func newKeyStoreDB(storage *badger.Storage) *keyStoreDB {
	return &keyStoreDB{
		storage: storage,
	}
}

func (db *keyStoreDB) GetDefaultAddress() (common.Address, error) {
	data, err := db.storage.GetData(defaultAddressKey)
	if err != nil {
		return noneAddress, err
	}
	return common.Bytes2Address(data), nil
}

func (db *keyStoreDB) Foreach(fn func(address common.Address, key *ecdsa.PrivateKey)) {
	_ = db.storage.PrefixForeachData(addrKeyPre, func(k []byte, v []byte) error {
		privateKey, err := x509.ParseECPrivateKey(v)
		if err != nil {
			return err
		}
		addr := common.Bytes2Address(k)
		fn(addr, privateKey)
		return nil
	})
}

func (db *keyStoreDB) GetPrivateKey(address common.Address) (*ecdsa.PrivateKey, error) {
	key := append(addrKeyPre, address.Bytes()...)
	keyDer, err := db.storage.GetData(key)
	if err != nil {
		return nil, err
	}
	return x509.ParseECPrivateKey(keyDer)
}

func (db *keyStoreDB) PutPrivateKey(addr common.Address, key *ecdsa.PrivateKey) error {
	bs, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return err
	}
	sKey := append(addrKeyPre, addr.Bytes()...)
	return db.storage.SetData(sKey, bs)
}

func (db *keyStoreDB) SetDefaultAddress(address common.Address) error {
	return db.storage.SetData(defaultAddressKey, address.Bytes())
}

func (db *keyStoreDB) RemoveAddress(address common.Address) error {
	key := append(addrKeyPre, address.Bytes()...)
	_, err := db.storage.GetData(key)
	if err != nil {
		return err
	}
	return db.storage.DelData(key)
}

func (db *keyStoreDB) DelDefault() error {
	_, err := db.storage.GetData(defaultAddressKey)
	if err != nil {
		return err
	}
	return db.storage.DelData(defaultAddressKey)
}
