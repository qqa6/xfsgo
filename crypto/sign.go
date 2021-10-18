package crypto

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"math/big"
	"xfsgo/common"
)

func ECDSASign2Hex(hash []byte, prv *ecdsa.PrivateKey) (string, error) {
	sig, err := ECDSASign(hash, prv)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(sig), nil
}

func ECDSASign(digestHash []byte, prv *ecdsa.PrivateKey) ([]byte, error) {
	if len(digestHash) != crypto.DigestLength {
		return nil, fmt.Errorf("hash is required to be exactly %d bytes (%d)", crypto.DigestLength , len(digestHash))
	}
	seckey := math.PaddedBigBytes(prv.D, prv.Params().BitSize/8)
	defer zeroBytes(seckey)
	return secp256k1.Sign(digestHash, seckey)
}



//func VerifySignature(data []byte, sig []byte) bool {
//	totalLen := sig[0]
//	sigAll := sig[1 : totalLen+1]
//	sigBuf := bytes.NewBuffer(sigAll)
//	rBytes, err := common.ReadMixedBytes(sigBuf)
//	if err != nil {
//		return false
//	}
//	r := new(big.Int).SetBytes(rBytes)
//
//	sBytes, err := common.ReadMixedBytes(sigBuf)
//	if err != nil {
//		return false
//	}
//	s := new(big.Int).SetBytes(sBytes)
//	xBytes, err := common.ReadMixedBytes(sigBuf)
//	if err != nil {
//		return false
//	}
//	x := new(big.Int).SetBytes(xBytes)
//	yBytes, err := common.ReadMixedBytes(sigBuf)
//	if err != nil {
//		return false
//	}
//	y := new(big.Int).SetBytes(yBytes)
//	pub := &ecdsa.PublicKey{
//		Curve: elliptic.P256(),
//		X:     x,
//		Y:     y,
//	}
//	return ecdsa.Verify(pub, data, r, s)
//}

func Ecrecover(hash, sig []byte) ([]byte, error) {
	return secp256k1.RecoverPubkey(hash, sig)
}

func SigToPub(hash, sig []byte) (*ecdsa.PublicKey, error) {
	s, err := Ecrecover(hash, sig)
	if err != nil {
		return nil, err
	}
	x, y := elliptic.Unmarshal(secp256k1.S256(), s)
	return &ecdsa.PublicKey{Curve: secp256k1.S256(), X: x, Y: y}, nil
}

func ParsePubKeyFromSignature(sig []byte) (ecdsa.PublicKey, error) {
	totalLen := sig[0]
	sigAll := sig[1 : totalLen+1]
	sigBuf := bytes.NewBuffer(sigAll)
	rBytes, err := common.ReadMixedBytes(sigBuf)
	if err != nil {
		return ecdsa.PublicKey{}, err
	}
	_ = new(big.Int).SetBytes(rBytes)
	sBytes, err := common.ReadMixedBytes(sigBuf)
	if err != nil {
		return ecdsa.PublicKey{}, err
	}
	_ = new(big.Int).SetBytes(sBytes)
	xBytes, err := common.ReadMixedBytes(sigBuf)
	if err != nil {
		return ecdsa.PublicKey{}, err
	}
	x := new(big.Int).SetBytes(xBytes)
	yBytes, err := common.ReadMixedBytes(sigBuf)
	if err != nil {
		return ecdsa.PublicKey{}, err
	}
	y := new(big.Int).SetBytes(yBytes)
	return ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     x,
		Y:     y,
	}, nil
}
func zeroBytes(bytes []byte) {
	for i := range bytes {
		bytes[i] = 0
	}
}

