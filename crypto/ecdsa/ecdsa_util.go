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

package ecdsa

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
)

func GenP256PrivateKey() []byte {
	key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	bs, _ := x509.MarshalECPrivateKey(key)
	return bs
}
func ParsePubKeyWithPrivateKey(bytes []byte) []byte {
	key, _ := x509.ParseECPrivateKey(bytes)
	pub := key.PublicKey
	bs, _ := x509.MarshalPKIXPublicKey(&pub)
	return bs
}
