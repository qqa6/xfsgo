package xblockchain

import (
	"crypto/sha256"
	"encoding/json"
	"github.com/btcsuite/btcutil/base58"
	"xblockchain/util/crypto/ecdsa"
	"xblockchain/util/urlsafeb64"
)

type Wallet struct {
	privateKey []byte
}

func NewRandomWallet() *Wallet {
	var privateKey []byte = nil
	privateKey = ecdsa.GenP256PrivateKey()
	return &Wallet{
		privateKey: privateKey,
	}
}

func ImportWalletWithB64(enc string) (*Wallet,error) {
	privateKey,err := urlsafeb64.Decode(enc)
	if err != nil {
		return nil, err
	}
	return ImportWalletWithKey(privateKey),err
}

func ImportWalletWithKey(privateKey []byte) *Wallet {
	return &Wallet{
		privateKey: privateKey,
	}
}

func (w *Wallet) GetPublicKey() []byte {
	pubKey := ecdsa.ParsePubKeyWithPrivateKey(w.privateKey)
	return pubKey
}

func (w *Wallet) Export() string {
	return urlsafeb64.Encode(w.privateKey)
}

func (w *Wallet) GetPrivateKey() []byte  {
	return w.privateKey
}

func (w *Wallet) GetAddress() string  {
	pubkey := w.GetPublicKey()
	pubKeyHash := PubKeyHash(pubkey)
	return PubKeyHash2Address(pubKeyHash)
}

func PubKeyHash(pubKey []byte) []byte  {
	hash := sha256.Sum256(pubKey)
	return hash[:]
}

func PubKeyHash2Address(pubKeyHash []byte) string {
	fullPayload := append([]byte{0}, pubKeyHash[:]...)
	addr := base58.Encode(fullPayload)
	return addr
}

func ParsePubKeyHashByAddress(address string) []byte {
	payload := base58.Decode(address)
	pubKeyHash := payload[1:]
	return pubKeyHash
}



func (ws *Wallets) String() string {
	jsonByte,err := json.Marshal(ws)
	if err != nil {
		return ""
	}
	return string(jsonByte)
}