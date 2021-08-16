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

package urlsafeb64

import (
	"encoding/base64"
	"strings"
)

func Encode(src []byte) string {
	sEnc := base64.StdEncoding.EncodeToString(src)
	sEnc = strings.ReplaceAll(sEnc, "+", "-")
	sEnc = strings.ReplaceAll(sEnc, "/", "_")
	sEnc = strings.ReplaceAll(sEnc, "=", "")
	return sEnc
}

func Decode(enc string) ([]byte, error) {
	encStr := enc
	encStr = strings.ReplaceAll(encStr, "-", "+")
	encStr = strings.ReplaceAll(encStr, "_", "/")
	return base64.RawStdEncoding.DecodeString(encStr)
}
