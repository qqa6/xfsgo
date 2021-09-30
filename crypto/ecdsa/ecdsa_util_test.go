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
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"testing"
)

func TestGenP256PrivateKey(t *testing.T) {
	key := GenP256PrivateKey()
	pub := ParsePubKeyWithPrivateKey(key)
	bas := base64.StdEncoding.EncodeToString(pub)
	t.Logf("k: %s\n", bas)
}

func TestSign(t *testing.T) {
	key := GenP256PrivateKey()
	basKey := base64.StdEncoding.EncodeToString(key)
	t.Logf("key: %s\n", basKey)

	pub := ParsePubKeyWithPrivateKey(key)
	basPub := base64.StdEncoding.EncodeToString(pub)
	t.Logf("pub: %s\n", basPub)

	privateKey, err := x509.ParseECPrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	data := []byte("def")
	signature, err := privateKey.Sign(rand.Reader, data, nil)
	if err != nil {
		t.Fatal(err)
	}
	basSignature := base64.StdEncoding.EncodeToString(signature)
	t.Logf("signature: %s\n", basSignature)

	pkGen, err := x509.ParsePKIXPublicKey(pub)
	if err != nil {
		t.Fatal(err)
	}
	pk := pkGen.(*ecdsa.PublicKey)
	bl := ecdsa.VerifyASN1(pk, data, signature)
	if !bl {
		t.Fatal("VerifyASN1 err")
	}
}
