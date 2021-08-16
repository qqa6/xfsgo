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

package uint256

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
)

const uInt256Size = 32

type UInt256 [uInt256Size]byte

func HexDigit(b byte) int {
	hexdigit := [256]int{
		-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
		-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
		-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9, -1, -1, -1, -1, -1, -1,
		0, 0xa, 0xb, 0xc, 0xd, 0xe, 0xf, -1, -1, -1, -1, -1, -1, -1, -1, -1,
		-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
		0, 0xa, 0xb, 0xc, 0xd, 0xe, 0xf, -1, -1, -1, -1, -1, -1, -1, -1, -1,
		-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
		-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
		-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
		-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
		-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
		-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
		-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
		-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
		-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1}
	return hexdigit[int(b)]
}
func NewUInt256ByHex(s string) *UInt256 {
	return NewUInt256(s)
}
func NewUInt256ByUInt32(n uint32) *UInt256 {
	u3 := byte((n & uint32(0xff<<24)) >> 24)
	u2 := byte((n & uint32(0xff<<16)) >> 16)
	u1 := byte((n & uint32(0xff<<8)) >> 8)
	u0 := byte((n & uint32(0xff<<0)) >> 0)
	bs := &UInt256{u0, u1, u2, u3}
	return bs
}
func NewUInt256Zero() *UInt256 {
	bs := &UInt256{0}
	return bs
}
func NewUInt256Max() *UInt256 {
	bs := &UInt256{
		255, 255, 255, 255, 255, 255, 255, 255,
		255, 255, 255, 255, 255, 255, 255, 255,
		255, 255, 255, 255, 255, 255, 255, 255,
		255, 255, 255, 255, 255, 255, 255, 255,
	}
	return bs
}
func NewUInt256One() *UInt256 {
	bs := &UInt256{1}
	return bs
}
func NewUInt256(str string) *UInt256 {
	bs := UInt256{}
	index := 0
	if str[0] == '0' && str[1] == 'x' {
		index += 2
	}
	pbegin := index
	for i := index; i < len(str); i++ {
		bi := HexDigit(str[i])
		if bi != -1 {
			index++
		} else {
			break
		}
	}
	p1 := 0
	pend := uInt256Size
	index -= 1
	for index >= pbegin && p1 < pend {
		h1 := HexDigit(str[index])
		bs[p1] = byte(h1 & 0xff)
		index -= 1
		if index >= pbegin {
			h2 := HexDigit(str[index])
			h2 = h2 << 4
			h1h2 := h1 | h2
			bs[p1] = byte(h1h2)
			index -= 1
			p1 += 1
		}
	}
	return &bs
}

func NewUInt256BigEndian(bs []byte) *UInt256 {
	tmp := &UInt256{}
	j := 0
	for i := uInt256Size - 1; i >= 0; i-- {
		tmp[j] = bs[i]
		j += 1
	}
	return tmp
}

func NewUInt256BS(bs []byte) *UInt256 {
	tmp := &UInt256{}
	j := 0
	for i := uInt256Size - 1; i >= 0; i-- {
		tmp[j] = bs[i]
		j += 1
	}
	return tmp
}

func (uint256 *UInt256) HexstrFull() string {
	return uint256.Hexstr(false)
}
func (uint256 *UInt256) Hexstr(skipzore bool) string {
	hexmap := [16]string{
		"0", "1", "2", "3", "4", "5", "6", "7",
		"8", "9", "a", "b", "c", "d", "e", "f"}
	s := ""
	for i := uInt256Size - 1; i >= 0; i-- {
		bi := uint(uint256[i])
		a := bi >> 4
		b := 15 & bi
		if !(skipzore && a == 0 && b == 0) {
			s += hexmap[a] + hexmap[b]
		}
	}
	return s
}

func (uint256 *UInt256) Add(u *UInt256) *UInt256 {
	carry := uint(0)
	bs := UInt256{}
	for i := 0; i < uInt256Size; i++ {
		ai := uint(uint256[i])
		bi := uint(u[i])
		ri := carry + ai + bi
		bs[i] = byte(ri & 0xff)
		carry = ri >> 8
	}
	return &bs
}

func (uint256 *UInt256) Sub(u *UInt256) *UInt256 {
	tmp := UInt256{}
	for i := 0; i < uInt256Size; i++ {
		bi := uint(u[i])
		tmp[i] = byte(^bi)
	}
	one := NewUInt256One()
	c := tmp.Add(one)
	return uint256.Add(c)
}

func (uint256 *UInt256) Equals(u *UInt256) bool {
	for i := 0; i < uInt256Size; i++ {
		ai := uint(uint256[i])
		bi := uint(u[i])
		if ai != bi {
			return false
		}
	}
	return true
}

func (uint256 *UInt256) Xor(u *UInt256) *UInt256 {
	tmp := UInt256{}
	for i := 0; i < uInt256Size; i++ {
		ai := uint(uint256[i])
		bi := uint(u[i])
		tmp[i] = byte(ai ^ bi)
	}
	return &tmp
}

func (uint256 *UInt256) IsZero() bool {
	for i := 0; i < uInt256Size; i++ {
		b := uint256[i]
		if b != 0 {
			return false
		}
	}
	return true
}

func (uint256 *UInt256) ToBytes() []byte {
	tmp := [uInt256Size]byte{}
	j := 0
	for i := uInt256Size - 1; i >= 0; i-- {
		tmp[j] = uint256[i]
		j += 1
	}
	return tmp[:]
}

func (uint256 *UInt256) ToBigEndianBytesArr() [32]byte {
	tmp := [uInt256Size]byte{}
	j := 0
	for i := uInt256Size - 1; i >= 0; i-- {
		tmp[j] = uint256[i]
		j += 1
	}
	return tmp
}

func (uint256 *UInt256) Hex() string {
	return hex.EncodeToString(uint256.ToBytes())
}

func (uint256 *UInt256) Lsh(shift int) *UInt256 {
	k := shift / 8
	s := shift % 8
	tmp := NewUInt256Zero()
	for i := 0; i < uInt256Size; i++ {
		if i+k+1 < uInt256Size && s != 0 {
			tmp[i+k+1] |= uint256[i] >> (8 - s)
		}
		if i+k < uInt256Size {
			tmp[i+k] |= uint256[i] << s
		}
	}
	return tmp
}

func (uint256 *UInt256) Rsh(shift int) *UInt256 {
	k := shift / 8
	s := shift % 8
	big.NewInt(23)
	tmp := NewUInt256Zero()
	for i := 0; i < uInt256Size; i++ {
		if i-k-1 >= 0 && s != 0 {
			tmp[i-k-1] |= uint256[i] << (8 - s)
		}
		if i-k >= 0 {
			tmp[i-k] |= uint256[i] >> s
		}
	}
	return tmp
}
func (uint256 UInt256) Len() int {
	for i := uInt256Size - 1; i >= 0; i-- {
		a := uint256[i]
		if a > 0 {
			return i + 1
		}
	}
	return 0
}
func (uint256 *UInt256) Gt(target *UInt256) bool {
	return uint256.Cmp(target) == 1
}

func (uint256 *UInt256) Lt(target *UInt256) bool {
	return uint256.Cmp(target) == -1
}

func (uint256 *UInt256) Cmp(y *UInt256) int {
	m := uint256.Len()
	n := y.Len()
	if m != n || m == 0 {
		switch {
		case m < n:
			return -1
		case m > n:
			return 1
		}
	}
	i := m - 1
	for i > 0 && uint256[i] == y[i] {
		i--
	}
	switch {
	case uint256[i] < y[i]:
		return -1
	case uint256[i] > y[i]:
		return 1
	}
	return 0
}

func (uint256 *UInt256) MarshalJSON() ([]byte, error) {
	s := fmt.Sprintf("\"%s\"", uint256.Hex())
	return []byte(s), nil
}

func (uint256 *UInt256) UnmarshalJSON(data []byte) (err error) {
	hexs := strings.ReplaceAll(string(data), "\"", "")
	u256 := NewUInt256(hexs)
	*uint256 = *u256
	return
}

func (uint256 *UInt256) ToBigInt() *big.Int {
	return new(big.Int).SetBytes(uint256.ToBytes())
}

func (uint256 *UInt256) ToUint64() uint64 {
	return uint256.ToBigInt().Uint64()
}
