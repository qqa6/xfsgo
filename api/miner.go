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
	"runtime"
	"time"
	"xfsgo"
	"xfsgo/common"
	"xfsgo/miner"
)

type MinerAPIHandler struct {
	Miner *miner.Miner
}

type MinSetGasLimitArgs struct {
	GasLimit string `json:"gas_limit"`
}

type MinSetGasPriceArgs struct {
	GasPrice string `json:"gas_price"`
}

type MinCoinbaseArgs struct {
	Coinbase string `json:"coinbase"`
}

type MinWorkerArgs struct {
	WorkerNum int `json:"worker_num"`
}

type MinStatusArgs struct {
	Status        bool   `json:"status"`
	LastStartTime string `json:"last_start_time"`
	Workers       int    `json:"workers"`
	Coinbase      string `json:"coinbase"`
	GasPrice      string `json:"gas_price"`
	GasLimit      string `json:"gas_limit"`
	HashRate      int    `json:"hash_rate"`
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

func (handler *MinerAPIHandler) WorkersAdd(args MinWorkerArgs, resp *string) error {
	var num int
	if args.WorkerNum < 1 {
		num = 1
	} else {
		num = args.WorkerNum
	}
	NumWorkers := int(handler.Miner.GetWorkerNum()) + num
	maxWorkers := runtime.NumCPU() * 2
	if maxWorkers > NumWorkers {
		go func() {
			handler.Miner.SetNumWorkers(uint32(NumWorkers))
		}()
	}
	*resp = ""
	return nil
}

func (handler *MinerAPIHandler) WorkersDown(args MinWorkerArgs, resp *string) error {

	var num int
	if args.WorkerNum < 1 {
		num = 1
	} else {
		num = args.WorkerNum
	}

	NumWorkers := int(handler.Miner.GetWorkerNum()) - num
	if NumWorkers < 1 {
		NumWorkers = 1
	}
	go func() {
		handler.Miner.SetNumWorkers(uint32(NumWorkers))
	}()
	*resp = ""
	return nil
}

func (handler *MinerAPIHandler) MinSetCoinbase(args MinCoinbaseArgs, resp *string) error {
	if args.Coinbase == "" {
		return xfsgo.NewRPCError(-1006, "Parameter cannot be empty")
	}
	addr := common.StrB58ToAddress(args.Coinbase)
	handler.Miner.SetCoinbase(addr)
	return nil
}

func (handler *MinerAPIHandler) MinSetGasLimit(args MinSetGasLimitArgs, resp *string) error {
	if args.GasLimit == "" {
		return xfsgo.NewRPCError(-1006, "Parameter cannot be empty")
	}
	GasLimitBigInt := common.ParseString2BigInt(args.GasLimit)
	if GasLimitBigInt.Uint64() == uint64(0) {
		GasLimitBigInt = common.MinGasLimit
	}
	handler.Miner.SetGasLimit(GasLimitBigInt)
	return nil
}

func (handler *MinerAPIHandler) MinSetGasPrice(args MinSetGasPriceArgs, resp *string) error {
	if args.GasPrice == "" {
		return xfsgo.NewRPCError(-1006, "Parameter cannot be empty")
	}

	GasPriceBigInt := common.ParseString2BigInt(args.GasPrice)
	if GasPriceBigInt.Uint64() == uint64(0) {
		GasPriceBigInt = common.MinGasLimit
	}
	handler.Miner.SetGasPrice(GasPriceBigInt)
	return nil
}

func (handler *MinerAPIHandler) MinGetStatus(_ EmptyArgs, resp *MinStatusArgs) error {
	gasLimit := handler.Miner.GetGasLimit()
	gasPrice := handler.Miner.GetGasPrice()
	MinStartTime := handler.Miner.LastStartTime
	MinCoinbase := handler.Miner.Coinbase
	MinHashRate := handler.Miner.HashRate()
	MinWorkers := handler.Miner.GetWorkerNum()
	MinStatus := handler.Miner.GetMinStatus()

	result := &MinStatusArgs{
		Status:        MinStatus,
		GasPrice:      gasPrice.String(),
		GasLimit:      gasLimit.String(),
		LastStartTime: MinStartTime.Format(time.RFC3339),
		Coinbase:      MinCoinbase.B58String(),
		HashRate:      MinHashRate,
		Workers:       int(MinWorkers),
	}
	*resp = *result
	return nil
}
