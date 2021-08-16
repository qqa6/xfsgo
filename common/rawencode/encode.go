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

package rawencode

import "encoding/json"

type RawEncoder interface {
	Encode() ([]byte, error)
	Decode(data []byte) error
}

func Encode(en interface{}) ([]byte, error) {
	obj, ok := en.(RawEncoder)
	if !ok {
		return json.Marshal(en)
	}
	return obj.Encode()
}

func EncodeByLen(en interface{}) ([]byte, error) {
	obj, ok := en.(RawEncoder)
	if !ok {
		return json.Marshal(en)
	}
	return obj.Encode()
}

func Decode(bs []byte, en interface{}) error {
	obj, ok := en.(RawEncoder)
	if !ok {
		return json.Unmarshal(bs, en)
	}
	return obj.Decode(bs)
}
