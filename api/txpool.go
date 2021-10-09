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
	"math/big"
	"xfsgo"
	"xfsgo/common"
)

type TxPoolHandler struct {
	TxPool *xfsgo.TxPool
}

type GetTranByHashArgs struct {
	Hash string `json:"hash"`
}

type ModTranGasArgs struct {
	GasLimit string `json:"gas_limit"`
	GasPrice string `json:"gas_price"`
	Hash     string `json:"hash"`
}

func (tx *TxPoolHandler) GetPending(_ EmptyArgs, resp *transactions) error {
	data := tx.TxPool.GetTransactions()
	*resp = data
	return nil
}

func (tx *TxPoolHandler) GetPendingSize(_ EmptyArgs, resp *int) error {
	data := tx.TxPool.GetTransactionsSize()
	*resp = data
	return nil
}

func (tx *TxPoolHandler) GetTranByHash(args GetTranByHashArgs, resp *xfsgo.Transaction) error {
	if args.Hash == "" {
		return xfsgo.NewRPCError(-1006, "Parameter cannot be empty")
	}
	tranObj := tx.TxPool.GetTransaction(args.Hash)
	*resp = *tranObj
	return nil
}

func (tx *TxPoolHandler) ModifyTranGas(args ModTranGasArgs, resp *string) error {
	if args.GasLimit == "" || args.GasPrice == "" || args.Hash == "" {
		return xfsgo.NewRPCError(-1006, "Parameter cannot be empty")
	}
	var gasLimit, gasPrice *big.Int
	gasLimit = common.ParseString2BigInt(args.GasLimit)

	gasPrice = common.ParseString2BigInt(args.GasPrice)

	if err := tx.TxPool.ModifyTranGas(gasLimit, gasPrice, args.Hash); err != nil {
		return xfsgo.NewRPCErrorCause(-32001, err)
	}
	return nil
}
