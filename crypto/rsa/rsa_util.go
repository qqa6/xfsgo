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

package rsa

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
)

func GenPrivateKey(bits int) []byte {
	privateKey, _ := rsa.GenerateKey(rand.Reader, bits)
	bs := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKey.Public()
	return bs
}

func ParsePubKeyWithPrivateKey(bytes []byte) []byte {
	privateKey, _ := x509.ParsePKCS1PrivateKey(bytes)
	pubKey := privateKey.PublicKey
	bs := x509.MarshalPKCS1PublicKey(&pubKey)
	return bs
}
