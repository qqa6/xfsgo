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
	"encoding/hex"
	"math/big"
	"sort"
	"xfsgo"
	"xfsgo/common"

	"github.com/sirupsen/logrus"
)

type WalletHandler struct {
	Wallet        *xfsgo.Wallet
	BlockChain    *xfsgo.BlockChain
	TxPendingPool *xfsgo.TxPool
}

type WalletByAddressArgs struct {
	Address string `json:"address"`
}

type WalletImportArgs struct {
	Key string `json:"key"`
}

type SetDefaultAddrArgs struct {
	Address string `json:"address"`
}

// type TransferArgs struct {
// 	To       string `json:"to"`
// 	GasLimit string `json:"gas_limit"`
// 	GasPrice string `json:"gas_price"`
// 	Value    string `json:"value"`
// }

type SendTransactionArgs struct {
	From     string `json:"from"`
	To       string `json:"to"`
	GasLimit string `json:"gas_limit"`
	GasPrice string `json:"gas_price"`
	Value    string `json:"value"`
}

type SetGasLimitArgs struct {
	Gas string `json:"gas"`
}

type SetGasPriceArgs struct {
	GasPrice string `json:"gas_price"`
}

func (handler *WalletHandler) Create(_ EmptyArgs, resp *string) error {
	addr, err := handler.Wallet.AddByRandom()
	if err != nil {
		return xfsgo.NewRPCErrorCause(-6001, err)
	}
	*resp = addr.B58String()
	return nil
}

func (handler *WalletHandler) Del(args WalletByAddressArgs, resp *interface{}) error {
	addr := common.StrB58ToAddress(args.Address)
	err := handler.Wallet.Remove(addr)
	if err != nil {
		return xfsgo.NewRPCErrorCause(-6001, err)
	}
	return nil
}

func (handler *WalletHandler) List(_ EmptyArgs, resp *[]common.Address) error {
	data := handler.Wallet.All()
	var out Wallets
	for addr, v := range data {
		_ = v
		r, _ := handler.Wallet.GetWalletNewTime(addr)
		req := &Wallet{
			addr:    addr,
			newTime: int64(common.Byte2Int(r)),
		}
		out = append(out, req)
	}

	sort.Sort(Wallets(out))

	result := make([]common.Address, 0)
	for i := 0; i < len(out); i++ {
		result = append(result, out[i].addr)
	}

	*resp = result
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

	if args.Address == "" {
		return xfsgo.NewRPCError(-1006, "parameter cannot be empty")
	}
	addr := common.StrB58ToAddress(args.Address)
	if err := handler.Wallet.SetDefault(addr); err != nil {
		return xfsgo.NewRPCErrorCause(-6001, err)
	}
	return nil

}

func (handler *WalletHandler) ExportByAddress(args WalletByAddressArgs, resp *string) error {
	if args.Address == "" {
		return xfsgo.NewRPCError(-1006, "parameter cannot be empty")
	}
	addr := common.StrB58ToAddress(args.Address)
	pk, err := handler.Wallet.Export(addr)
	if err != nil {
		return xfsgo.NewRPCErrorCause(-6001, err)
	}
	*resp = "0x" + hex.EncodeToString(pk)
	return nil
}

func (handler *WalletHandler) ImportByPrivateKey(args WalletImportArgs, resp *string) error {
	if args.Key == "" {
		return xfsgo.NewRPCError(-1006, "parameter cannot be empty")
	}
	keyEnc := args.Key
	if len(keyEnc) > 1 {
		if keyEnc[0] == '0' && keyEnc[1] == 'x' {
			keyEnc = keyEnc[2:]
		}
	}
	keyDer, err := hex.DecodeString(keyEnc)
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

func (handler *WalletHandler) SendTransaction(args SendTransactionArgs, resp *string) error {

	// Judgment target address cannot be empty
	if args.To == "" {
		return xfsgo.NewRPCError(-1006, "to addr not be empty")
	}
	// Judge that the transfer amount cannot be blank
	if args.Value == "" {
		return xfsgo.NewRPCError(-1006, "value not be empty")
	}

	// Get the wallet address of the initiating transaction
	var fromAddr common.Address
	if args.From != "" {
		fromAddr = common.B58ToAddress([]byte(args.From))
	} else {
		fromAddr = handler.Wallet.GetDefault()
	}
	// Take out the private key according to the wallet address of the initiating transaction
	privateKey, err := handler.Wallet.GetKeyByAddress(fromAddr)
	if err != nil {
		return xfsgo.NewRPCErrorCause(-1006, err)
	}

	toAddr := common.B58ToAddress([]byte(args.To))
	value := common.ParseString2BigInt(args.Value)

	// Set the gas price and gas limit of this transaction
	var GasLimit, GasPrice *big.Int
	if args.GasLimit != "" {
		GasLimit = common.ParseString2BigInt(args.GasLimit)
	} else {
		GasLimit = handler.Wallet.GetGas()
	}

	if args.GasPrice != "" {
		GasPrice = common.ParseString2BigInt(args.GasPrice)
	} else {
		GasPrice = handler.Wallet.GetGasPrice()
	}

	logrus.Debugf("transaction obj: gasPrice=%v, gasLimit=%v, From=%v, to=%v val:%v", GasPrice, GasLimit, fromAddr.B58String(), args.To, args.Value)

	tx := xfsgo.NewTransaction(toAddr, GasLimit, GasPrice, common.BaseCoin2Atto(float64(value.Uint64())))
	tx.Nonce = handler.BlockChain.GetNonce(fromAddr)
	err = tx.SignWithPrivateKey(privateKey)
	if err != nil {
		return xfsgo.NewRPCErrorCause(-1006, err)
	}

	err = handler.TxPendingPool.Add(tx)
	if err != nil {
		return xfsgo.NewRPCErrorCause(-1006, err)
	}

	result := tx.Hash()
	*resp = result.Hex()
	return nil
}

func (handler *WalletHandler) SetGasLimit(args SetGasLimitArgs, resp *string) error {
	if args.Gas == "" {
		return xfsgo.NewRPCError(-1006, "Parameter cannot be empty")
	}

	GasLimitBigInt := common.ParseString2BigInt(args.Gas)
	if GasLimitBigInt.Uint64() == uint64(0) {
		GasLimitBigInt = common.DefaultGasPrice
	}
	handler.Wallet.SetGas(GasLimitBigInt)
	return nil
}

func (handler *WalletHandler) SetGasPrice(args SetGasPriceArgs, resp *string) error {
	if args.GasPrice == "" {
		return xfsgo.NewRPCError(-1006, "Parameter cannot be empty")
	}

	GasPriceBigInt := common.ParseString2BigInt(args.GasPrice)
	if GasPriceBigInt.Uint64() == uint64(0) {
		GasPriceBigInt = common.DefaultGasPrice
	}
	handler.Wallet.SetGasPrice(GasPriceBigInt)
	return nil
}
