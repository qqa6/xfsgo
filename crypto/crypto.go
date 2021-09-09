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

package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/md5"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"xfsgo/common"
	"xfsgo/common/ahash"
	"xfsgo/common/urlsafeb64"
)

func GenPrvKey() (*ecdsa.PrivateKey, error) {
	return ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
}

func PubKeyEncode(p ecdsa.PublicKey) []byte {
	if p.Curve == nil || p.X == nil || p.Y == nil {
		return nil
	}
	xbs := p.X.Bytes()
	ybs := p.Y.Bytes()
	buf := make([]byte, len(xbs)+len(ybs))
	copy(buf, append(xbs, ybs...))
	return buf
}

func Checksum(payload []byte) []byte {
	first := ahash.SHA256(payload)
	second := ahash.SHA256(first)
	return second[:common.AddrCheckSumLen]
}

func VerifyAddress(addr common.Address) bool {
	want := Checksum(addr.Payload())
	got := addr.Checksum()
	return bytes.Compare(want, got) == 0
}

func DefaultPubKey2Addr(p ecdsa.PublicKey) common.Address {
	return PubKey2Addr(common.AddrDefaultVersion, p)
}

func PubKeySha256HashBs(p ecdsa.PublicKey) []byte {
	pubEnc := PubKeyEncode(p)
	pubHash256 := ahash.SHA256(pubEnc)
	return pubHash256
}
func PubKeySha256Hash(p ecdsa.PublicKey) common.Hash {
	pubHash256 := PubKeySha256HashBs(p)
	return common.Bytes2Hash(pubHash256)
}

func PubKey2Addr(version uint8, p ecdsa.PublicKey) common.Address {
	pubEnc := PubKeyEncode(p)
	pubHash256 := ahash.SHA256(pubEnc)
	pubHash := ahash.Ripemd160(pubHash256)
	payload := append([]byte{version}, pubHash...)
	cs := Checksum(payload)
	full := append(payload, cs...)
	return common.Bytes2Address(full)
}
func PrivateKeyEncodeB64String(key *ecdsa.PrivateKey) (string, error) {
	der, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return "", err
	}
	return urlsafeb64.Encode(der), nil
}

func B64StringDecodePrivateKey(enc string) (*ecdsa.PrivateKey, error) {
	der, err := urlsafeb64.Decode(enc)
	if err != nil {
		return nil, err
	}
	return x509.ParseECPrivateKey(der)
}

func ByteHash256(raw []byte) common.Hash {
	h := ahash.SHA256(raw)
	return common.Bytes2Hash(h)
}

func MD5Str(src []byte) string {
	h := md5.New()
	h.Write(src)
	return hex.EncodeToString(h.Sum(nil))
}

func AesEncrypt(encodeMsg []byte, key string) string {
	k := []byte(key)
	block, _ := aes.NewCipher(k)
	blockSize := block.BlockSize()
	encodeMsg = PKCS7Padding(encodeMsg, blockSize)
	blockMode := cipher.NewCBCEncrypter(block, k[:blockSize])
	cryted := make([]byte, len(encodeMsg))
	blockMode.CryptBlocks(cryted, encodeMsg)
	return base64.StdEncoding.EncodeToString(cryted)
}

func AesDecrypt(decodeMsg string, key string) string {
	crytedByte, _ := base64.StdEncoding.DecodeString(decodeMsg)
	k := []byte(key)
	block, _ := aes.NewCipher(k)
	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, k[:blockSize])
	orig := make([]byte, len(crytedByte))
	blockMode.CryptBlocks(orig, crytedByte)
	orig = PKCS7UnPadding(orig)
	return string(orig)
}

func PKCS7Padding(ciphertext []byte, blocksize int) []byte {
	padding := blocksize - len(ciphertext)%blocksize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func PKCS7UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}
