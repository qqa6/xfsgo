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

import "math/big"

var (
	// bigOne is 1 represented as a big.Int.  It is defined here to avoid
	// the overhead of creating it multiple times.
	bigOne  = big.NewInt(1)
	big0xff = big.NewInt(0xff)

	// oneLsh256 is 1 shifted left 256 bits.  It is defined here to avoid
	// the overhead of creating it multiple times.
	oneLsh256 = new(big.Int).Lsh(bigOne, 256)
	ffLsh256  = new(big.Int).Lsh(big0xff, 256)
)

// BigByZip zips 256 bit difficulty to uint32
func BigByZip(target *big.Int) uint32 {
	if target.Sign() == 0 {
		return 0
	}
	c := uint(3)
	e := uint(len(target.Bytes()))
	var mantissa uint
	if e <= c {
		mantissa = uint(target.Bits()[0])
		shift := 8 * (c - e)
		mantissa <<= shift
	} else {
		shift := 8 * (e - c)
		mantissaNum := target.Rsh(target, shift)
		mantissa = uint(mantissaNum.Bits()[0])
	}
	mantissa <<= 8
	mantissa = mantissa & 0xffffffff
	return uint32(mantissa | e)
}

// BitsUnzip unzips 32bit Bits in BlockHeader to 256 bit Difficulty.
func BitsUnzip(bits uint32) *big.Int {
	mantissa := bits & 0xffffff00
	mantissa >>= 8
	e := uint(bits & 0xff)
	c := uint(3)
	var bn *big.Int
	if e <= c {
		shift := 8 * (c - e)
		mantissa >>= shift
		bn = big.NewInt(int64(mantissa))
	} else {
		bn = big.NewInt(int64(mantissa))
		shift := 8 * (e - c)
		bn.Lsh(bn, shift)
	}
	return bn
}

func CalcWorkload(bits uint32) *big.Int {
	difficultyNum := BitsUnzip(bits)
	if difficultyNum.Sign() <= 0 {
		return big.NewInt(0)
	}
	// (1 << 256) / (difficultyNum + 1)
	denominator := new(big.Int).Add(difficultyNum, bigOne)
	return new(big.Int).Div(oneLsh256, denominator)
}
