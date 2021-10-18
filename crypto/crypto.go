package crypto

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"errors"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"math/big"
	"xfsgo/common"
	"xfsgo/common/ahash"
)
const defaultKeyPackType = uint8(1)
const DefaultKeyPackVersion = uint8(1)

func GenPrvKey() (*ecdsa.PrivateKey, error) {
	return ecdsa.GenerateKey(secp256k1.S256(), rand.Reader)
}

func MustGenPrvKey() *ecdsa.PrivateKey {
	key, err := GenPrvKey()
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
	return bytes.Equal(want, got)
}

func DefaultPubKey2Addr(p ecdsa.PublicKey) common.Address {
	return PubKey2Addr(common.DefaultAddressVersion, p)
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

func EncodePrivateKey(version uint8, key *ecdsa.PrivateKey) []byte {
	dbytes := key.D.Bytes()
	buf := append([]byte{version, defaultKeyPackType}, dbytes...)
	return buf
}

func DefaultEncodePrivateKey(key *ecdsa.PrivateKey) []byte {
	return EncodePrivateKey(DefaultKeyPackVersion, key)
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
		curve := secp256k1.S256()
		curveOrder := curve.Params().N
		if k.Cmp(curveOrder) >= 0 {
			return 0, nil, errors.New("invalid elliptic curve private key value")
		}
		priv.Curve = curve
		priv.D = k
		privateKey := make([]byte, (curveOrder.BitLen()+7)/8)
		for len(payload) > len(privateKey) {
			if payload[0] != 0 {
				return 0, nil, errors.New("invalid private key length")
			}
			payload = payload[1:]
		}

		// Some private keys remove all leading zeros, this is also invalid
		// according to [SEC1] but since OpenSSL used to do this, we ignore
		// this too.
		copy(privateKey[len(privateKey)-len(payload):], payload)
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

