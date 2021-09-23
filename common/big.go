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

package common

import "math/big"

var Big0 = new(big.Int).SetInt64(0)

// var TxDataZeroGas = big.NewInt(4)                                                    // Per byte of data attached to a transaction that equals zero. NOTE: Not payable on data of calls between transactions.
// var TxDataNonZeroGas = big.NewInt(68)                                                // Per byte of data attached to a transaction that is not equal to zero. NOTE: Not payable on data of calls between transactions.
var TxGas = BaseCoin2Atto(float64(1))                                                // Per transaction. NOTE: Not payable on data of calls between transactions.
var GasLimitBoundDivisor = big.NewInt(1024)                                          // The bound divisor of the gas limit, used in update calculations.
var MinGasLimit = big.NewInt(5000)                                                   // Minimum the gas limit may ever be.
var GenesisGasLimit = BaseCoin2Atto(float64(121))                                    // Gas limit of the Genesis block.
var DefaultGasPrice = new(big.Int).Mul(big.NewInt(500), Atto2BaseCoin(new(big.Int))) // miner minGasPirce
func ParseString2BigInt(str string) *big.Int {
	if str == "" {
		return Big0
	}
	num, success := new(big.Int).SetString(str, 0)
	if !success {
		return Big0
	}
	return num
}

func BigMax(x, y *big.Int) *big.Int {
	if x.Cmp(y) < 0 {
		return y
	}

	return x
}

func BigMin(x, y *big.Int) *big.Int {
	if x.Cmp(y) > 0 {
		return y
	}

	return x
}

func Gasprice(price *big.Int, pct int64) *big.Int {
	p := new(big.Int).Set(price)
	p.Div(p, big.NewInt(100))
	p.Mul(p, big.NewInt(pct))
	return p
}
