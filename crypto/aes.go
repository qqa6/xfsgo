package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
)
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