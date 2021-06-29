package xblockchain

import (
	"github.com/btcsuite/btcutil/base58"
	"testing"
)

func TestWallet_GetAddress(t *testing.T) {
	w := NewRandomWallet()
	addr := w.GetAddress()
	t.Logf("address: %s\n", addr)
}

func TestWallet_GetPublicKey(t *testing.T) {
	privEnc := "MHcCAQEEILwSeKClYoJCpSkjbjpBOucLtR9PPUuSrEIzjWXzKzzUoAoGCCqGSM49AwEHoUQDQgAENugOXVaerJn1sth19RqM2eNuWrv2GihCNxs-YkN0uKHnJ-Y8eZdvyAm9HPl-3XBtgaWhqHicFFcGz9adRCLaEw"
	w,err := ImportWalletWithB64(privEnc)
	if err != nil {
		t.Fatal(err)
	}
	pubkey := w.GetPublicKey()
	strEnc := base58.Encode(pubkey)
	t.Logf("publicKey: %s\n",string(pubkey))
	t.Logf("publicKeyEnc: %s\n",strEnc)
}

