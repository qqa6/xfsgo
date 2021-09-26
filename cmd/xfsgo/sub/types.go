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

package sub

import (
	"encoding/json"
)

type getBlockHashArgs struct {
	Hash string `json:"hash"`
}

type getTransactionArgs struct {
	Hash string `json:"hash"`
}

type getReceiptArgs struct {
	Hash string `json:"hash"`
}

type gettxbyBlocknumArgs struct {
	Number json.Number `json:"number"`
}

type gettxbyBlockHashArgs struct {
	Hash string `json:"hash"`
}

type getAccountArgs struct {
	RootHash string `json:"root_hash"`
	Address  string `json:"address"`
}

type getWalletByAddressArgs struct {
	Address string `json:"address"`
}

type walletImportArgs struct {
	Key string `json:"key"`
}

type setWalletAddrDefArgs struct {
	Address string `json:"address"`
}

type transferFromArgs struct {
	From  string `json:"form"`
	To    string `json:"to"`
	Value string `json:"value"`
}

type getBlockNumArgs struct {
	From  json.Number `json:"from"`
	Count json.Number `json:"count"`
}

type GetBlocksArgs struct {
	Blocks string `json:"blocks"`
}

type MinSetGasPriceArgs struct {
	GasPrice string `json:"gas_price"`
}
type MinSetGasLimitArgs struct {
	GasLimit string `json:"gas_limit"`
}

type MinWorkerArgs struct {
	WorkerNum int `json:"worker_num"`
}

type MinSetCoinbaseArgs struct {
	Coinbase string `json:"coinbase"`
}

type GasLimitArgs struct {
	Gas string `json:"gas"`
}
type SetGasPriceArgs struct {
	GasPrice string `json:"gas_price"`
}

type TranGasArgs struct {
	GasLimit string `json:"gas_limit"`
	GasPrice string `json:"gas_price"`
	Hash     string `json:"hash"`
}
