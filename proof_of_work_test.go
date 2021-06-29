package xblockchain

import (
	"testing"
	"xblockchain/uint256"
)

func TestProofOfWork_Run(t *testing.T) {
}

func TestNewProofOfWork(t *testing.T) {
	want := uint256.NewUInt256ByHex("0x10000000000000000000000000000000000000000000000000000000000")
	target := uint256.NewUInt256One()
	target = target.Lsh(256-targetBits)
	t.Logf("target: %s\n",target.HexstrFull())
	if !target.Equals(want){
		t.Fatalf("target '0x%s' want '0x%s'", target.Hexstr(true), want.Hexstr(true))
	}
}

