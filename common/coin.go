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

import (
	"math/big"
	"strconv"
)

const Coin = 10 ^ 6

const Atto = 10e-18
const AttoCoin = 10e18
const Nano = 10e-9

var MaxAtto = new(big.Int).SetBytes([]byte{
	0xff,0xff,0xff,0xff,0xff,0xff,0xff,0xff,
	0xff,0xff,0xff,0xff,0xff,0xff,0xff,0xff,
	0xff,0xff,0xff,0xff,0xff,0xff,0xff,0xff,
	0xff,0xff,0xff,0xff,0xff,0xff,0xff,0xff,
})

func BaseCoin2Atto(coin float64) *big.Int {
	attocoin := coin * AttoCoin
	numberStr := strconv.FormatFloat(attocoin, 'f',-1,64 )
	result, ok := new(big.Int).SetString(numberStr,10)
	if !ok {
		return new(big.Int).SetUint64(0)
	}
	if result.Sign() < 0{
		return new(big.Int).SetUint64(0)
	}
	if result.Cmp(MaxAtto) >= 0 {
		return MaxAtto
	}
	return result
}