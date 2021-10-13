package crypto

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"errors"
	"math/big"
	"xfsgo/common"
	"xfsgo/common/ahash"
	"xfsgo/common/urlsafeb64"
)
const exportTypeRaw= uint8(1)
func GenPrvKey() (*ecdsa.PrivateKey, error) {
	return ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
}

func MustGenPrvKey() *ecdsa.PrivateKey {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		print(err)
	}
	return key
}

func PubKeyEncode(p ecdsa.PublicKey) []byte {
	if p.Curve == nil || p.X == nil || p.Y == nil {
		return nil
	}
	xbs := p.X.Bytes()
	ybs := p.Y.Bytes()
	buf := make([]byte, len(xbs) + len(ybs))
	copy(buf,append(xbs,ybs...))
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

func EncodePrivateKey(version uint8, key *ecdsa.PrivateKey) []byte {
	dbytes := key.D.Bytes()
	buf := append([]byte{version, exportTypeRaw}, dbytes...)
	return buf
}

func DecodePrivateKey(bs []byte) (uint8, *ecdsa.PrivateKey, error) {
	if len(bs) <= 2{
		return 0, nil, errors.New("unknown private key version")
	}
	version := bs[0]
	keytype := bs[1]
	payload := bs[2:]
	priv := new(ecdsa.PrivateKey)
	if keytype == 1 {
		k := new(big.Int).SetBytes(payload)
		curve := elliptic.P256()
		curveOrder := curve.Params().N
		if k.Cmp(curveOrder) >= 0 {
			return 0, nil, errors.New("invalid elliptic curve private key value")
		}
		priv.Curve = curve
		priv.D = k
		privateKey := make([]byte, (curveOrder.BitLen()+7)/8)
		copy(privateKey, payload)
		priv.X, priv.Y = curve.ScalarBaseMult(privateKey)
	}else {
		return 0, nil,  errors.New("unknown private key encrypt type")
	}
	return version, priv, nil
}

func ByteHash256(raw []byte) common.Hash {
	h := ahash.SHA256(raw)
	return common.Bytes2Hash(h)
}

