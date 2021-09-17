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
	"encoding/json"
	"runtime"
	"xfsgo/miner"
)

type MinerAPIHandler struct {
	Miner *miner.Miner
}

type SetGasLimitArgs struct {
	Gas json.Number `json:"gas"`
}

type SetGasPriceArgs struct {
	GasPrice json.Number `json:"gas_price"`
}

func (handler *MinerAPIHandler) Start(_ EmptyArgs, resp *string) error {
	handler.Miner.Start()
	*resp = ""
	return nil
}

func (handler *MinerAPIHandler) Stop(_ EmptyArgs, resp *string) error {
	handler.Miner.Stop()
	*resp = ""
	return nil
}

func (handler *MinerAPIHandler) WorkersAdd(_ EmptyArgs, resp *string) error {
	NumWorkers := int(handler.Miner.GetNumWorkers()) + 1
	maxWorkers := runtime.NumCPU() * 2
	if maxWorkers > NumWorkers {
		handler.Miner.SetNumWorkers(uint32(NumWorkers))
	}
	*resp = ""
	return nil
}

func (handler *MinerAPIHandler) WorkersDown(_ EmptyArgs, resp *string) error {
	NumWorkers := int(handler.Miner.GetNumWorkers()) - 1
	if NumWorkers < 1 {
		NumWorkers = 1
	}
	handler.Miner.SetNumWorkers(uint32(NumWorkers))
	*resp = ""
	return nil
}

func (handler *MinerAPIHandler) SetGasLimit(args SetGasLimitArgs, resp *string) error {

	return nil
}

func (handler *MinerAPIHandler) SetGasPrices(args SetGasPriceArgs, resp *string) error {
	// if args.GasPrice == "" {
	// 	return xfsgo.NewRPCError(-1006, "value not be empty")
	// }
	// gasPrice, err := args.GasPrice.Int64()
	// if err != nil {
	// 	return xfsgo.NewRPCErrorCause(-32001, err)
	// }
	// priceNewBigInt := new(big.Int).SetInt64(gasPrice)
	// err = handler.Miner.SetGasPrice(priceNewBigInt)
	// if err != nil {

	// }
	// 	var defaultGasPrice = new(big.Int).SetUint64(1)    //150000000000
	// var defaultGas = common.BaseCoin2Atto(float64(10)) //500000
	return nil
}
