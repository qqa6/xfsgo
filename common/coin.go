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
	"math"
	"math/big"
)

const Coin = 10 ^ 6

var AttoCoin = uint64(math.Pow(10, 18))
var NanoCoin = uint64(math.Pow(10, 9))

func BaseCoin2Atto(coin float64) *big.Int {
	a := big.NewFloat(coin)
	attocoin := a.Mul(a, big.NewFloat(float64(AttoCoin)))
	i, _ := attocoin.Int(nil)
	return i
}

func Atto2BaseCoin(atto *big.Int) *big.Int {
	i := big.NewInt(0)
	i.Add(i,atto)
	i.Div(i, big.NewInt(int64(AttoCoin)))
	return i
}

func BaseCoin2Nano(coin float64) *big.Int {
	a := big.NewFloat(coin)
	nanocoin := a.Mul(a, big.NewFloat(float64(NanoCoin)))
	i, _ := nanocoin.Int(nil)
	return i
}

func NanoCoin2BaseCoin(nano *big.Int) *big.Int {
	i := big.NewInt(0)
	i.Add(i, nano)
	i.Div(i, big.NewInt(int64(NanoCoin)))
	return i
}
