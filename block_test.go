package xblockchain

import (
	"fmt"
	"testing"
)

func TestNewBlock(t *testing.T) {
	txs,err := NewCoinBaseTransaction("abc", "aaaa")
	if err != nil {
		t.Fatal(err)
	}
	b := NewBlock(nil,[]*Transaction{txs})
	fmt.Printf("hash: %s", b.Hash.Hexstr(false))
}

func TestBlock_Serialize(t *testing.T) {
	txs,err := NewCoinBaseTransaction("abc", "aaaa")
	if err != nil {
		t.Fatal(err)
	}
	b := NewBlock(nil,[]*Transaction{txs})
	bs,err := b.Serialize()
	if err != nil {
		t.Fatal(err)
		return
	}
	t.Logf("Serialized: %x...%x\n",bs[:2],bs[len(bs)-2:])
	block,err := ReadBlockOfBytes(bs)
	if err != nil {
		t.Fatal(err)
		return
	}
	t.Logf("block hash: %s\n",block)
}