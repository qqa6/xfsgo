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

package api

import (
	"bytes"
	"xfsgo"
	"xfsgo/common"
	"xfsgo/common/urlsafeb64"
)

type WalletHandler struct {
	Wallet        *xfsgo.Wallet
	BlockChain    *xfsgo.BlockChain
	TxPendingPool *xfsgo.TxPool
}

type GetWalletByAddressArgs struct {
	Address string `json:"address"`
}

type WalletImportArgs struct {
	Key string `json:"key"`
}

type SetDefaultAddrArgs struct {
	Address string `json:"address"`
}

type TransferArgs struct {
	To    string `json:"to"`
	Value string `json:"value"`
}

type TransferFromArgs struct {
	From  string `json:"form"`
	To    string `json:"to"`
	Value string `json:"value"`
}

func (handler *WalletHandler) Create(_ EmptyArgs, resp *string) error {
	addr, err := handler.Wallet.AddByRandom()
	if err != nil {
		return xfsgo.NewRPCErrorCause(-6001, err)
	}
	*resp = addr.B58String()
	return nil
}

func (handler *WalletHandler) Del(args GetWalletByAddressArgs, resp *interface{}) error {
	addr := common.StrB58ToAddress(args.Address)
	err := handler.Wallet.Remove(addr)
	if err != nil {
		return xfsgo.NewRPCErrorCause(-6001, err)
	}
	return nil
}

func (handler *WalletHandler) List(_ EmptyArgs, resp *[]common.Address) error {
	data := handler.Wallet.All()
	out := make([]common.Address, 0)
	for addr, v := range data {
		_ = v
		out = append(out, addr)
	}
	*resp = out
	return nil
}

func (handler *WalletHandler) GetDefaultAddress(_ EmptyArgs, resp *string) error {
	address := handler.Wallet.GetDefault()
	zero := [25]byte{0}
	if bytes.Compare(address.Bytes(), zero[:]) == common.Zero {
		return nil
	}
	*resp = address.B58String()
	return nil
}

func (handler *WalletHandler) SetDefaultAddress(args SetDefaultAddrArgs, resp *string) error {
	addr := common.StrB58ToAddress(args.Address)
	if err := handler.Wallet.SetDefault(addr); err != nil {
		return xfsgo.NewRPCErrorCause(-6001, err)
	}
	return nil

}

func (handler *WalletHandler) ExportByAddress(args GetWalletByAddressArgs, resp *string) error {
	addr := common.StrB58ToAddress(args.Address)
	pk, err := handler.Wallet.Export(addr)
	if err != nil {
		return xfsgo.NewRPCErrorCause(-6001, err)
	}
	*resp = urlsafeb64.Encode(pk)
	return nil
}

func (handler *WalletHandler) ImportByPrivateKey(args WalletImportArgs, resp *string) error {
	keyEnc := args.Key
	keyDer, err := urlsafeb64.Decode(keyEnc)
	if err != nil {
		return xfsgo.NewRPCErrorCause(-6001, err)
	}
	addr, err := handler.Wallet.Import(keyDer)
	if err != nil {
		return xfsgo.NewRPCErrorCause(-6001, err)
	}
	*resp = addr.B58String()
	return nil
}

func (handler *WalletHandler) Transfer(args TransferArgs, resp *TransferObj) error {
	if args.To == "" {
		return xfsgo.NewRPCError(-1006, "to addr not be empty")
	}
	if args.Value == "" {
		return xfsgo.NewRPCError(-1006, "value not be empty")
	}
	formAddr, err := handler.Wallet.GetKeyByAddress(handler.Wallet.GetDefault())
	if err != nil {
		return err
	}
	toAddr := common.B58ToAddress([]byte(args.To))
	value := common.ParseString2BigInt(args.Value)
	tx := xfsgo.NewTransaction(toAddr, value)
	if err = tx.SignWithPrivateKey(formAddr); err != nil {
		return xfsgo.NewRPCErrorCause(-1006, err)
	}
	if err = handler.TxPendingPool.Add(tx); err != nil {
		return xfsgo.NewRPCErrorCause(-1006, err)
	}

	result := NewTransferObj(tx)
	*resp = *result
	return nil
}

func (handler *WalletHandler) TransferFrom(args TransferFromArgs, resp *TransferObj) error {
	if args.From == "" {
		return xfsgo.NewRPCError(-1006, "from addr not be empty")
	}
	if args.To == "" {
		return xfsgo.NewRPCError(-1006, "to addr not be empty")
	}
	if args.Value == "" {
		return xfsgo.NewRPCError(-1006, "value not be empty")
	}

	privateKey, err := handler.Wallet.GetKeyByAddress(common.B58ToAddress([]byte(args.From)))
	if err != nil {
		return xfsgo.NewRPCErrorCause(-1006, err)
	}

	toAddr := common.B58ToAddress([]byte(args.To))
	value := common.ParseString2BigInt(args.Value)
	tx := xfsgo.NewTransaction(toAddr, value)
	err = tx.SignWithPrivateKey(privateKey)
	if err != nil {
		return xfsgo.NewRPCErrorCause(-1006, err)
	}

	err = handler.TxPendingPool.Add(tx)
	if err != nil {
		return xfsgo.NewRPCErrorCause(-1006, err)
	}

	result := NewTransferObj(tx)
	*resp = *result
	return nil
}
