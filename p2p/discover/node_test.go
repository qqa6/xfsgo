package discover

import (
	"bytes"
	"fmt"
	"math/big"
	"net"
	"testing"
	"xfsgo/crypto"
)


func TestPubKey2NodeId(t *testing.T) {
	key, err := crypto.GenPrvKey()
	if err != nil {
		t.Fatal(err)
	}
	wantBuf := crypto.PubKeyEncode(key.PublicKey)
	gotId := PubKey2NodeId(key.PublicKey)
	if !bytes.Equal(wantBuf,gotId[:]){
		t.Fatalf("got id: %s, want: %x", gotId, wantBuf)
	}
}

func TestHex2NodeId(t *testing.T) {
	ref := NodeId{
		0,0,0,0,0,0,0,0,
		0,0,0,0,0,0,0,0,
		0,0,0,0,0,0,0,0,
		0,0,0,0,0,0,0,0,
		0,0,0,0,0,0,0,0,
		0,0,0,0,0,0,0,0,
		0,0,0,0,0,0,0,0,
		0,36,22,98,66,128,214,1}
	hex := "0x0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002416624280d601"
	nid, err := Hex2NodeId(hex)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(ref[:], nid[:]) {
		t.Fatalf("got id: %s, want: %s", nid, ref)
	}
}

func TestParseNode(t *testing.T) {
	nid := NodeId{
		0,0,0,0,0,0,0,0,
		0,0,0,0,0,0,0,0,
		0,0,0,0,0,0,0,0,
		0,0,0,0,0,0,0,0,
		0,0,0,0,0,0,0,0,
		0,0,0,0,0,0,0,0,
		0,0,0,0,0,0,0,0,
		0,36,22,98,66,128,214,1}
	nidHex := "0x0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002416624280d601"
	wantNode := newNode(net.IP{127,0,0,1},52150,22334, nid)
	rawUrl := fmt.Sprintf("xfsnode://%s@127.0.0.1:52150?discport=22334", nidHex)
	t.Logf("parse raw url: %s", rawUrl)
	got,err := ParseNode(rawUrl)

	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(wantNode.IP, got.IP) {
		t.Fatalf("got ip: %s, want: %s", got.IP, wantNode.IP)
	}
	if wantNode.TCP != got.TCP {
		t.Fatalf("got tcp port: %d, want: %d", got.TCP, wantNode.TCP)
	}
	if wantNode.UDP != got.UDP {
		t.Fatalf("got udp port: %d, want: %d", got.UDP, wantNode.UDP)
	}
	if !bytes.Equal(wantNode.ID[:],got.ID[:]) {
		t.Fatalf("got id: %s, want: %s", got.ID, wantNode.ID)
	}
	if !bytes.Equal(wantNode.Hash[:], got.Hash[:]){
		t.Fatalf("got hash: %s, want: %s", got.ID, wantNode.ID)
	}
}


func TestNode_logdist(t *testing.T) {
	logdistBig:= func(a,b []byte) int {
		aBig := new(big.Int).SetBytes(a)
		bBig := new(big.Int).SetBytes(b)
		return new(big.Int).Xor(aBig,bBig).BitLen()
	}
	var (
		a = [2]byte{1,2}
		b = [2]byte{3,4}
	)
	if n := logdist(a[:],a[:]); n != 0{
		t.Fatalf("got self logdist: %d, want: 0", n)
	}
	want := logdistBig(a[:],b[:])
	got := logdist(a[:],b[:])
	t.Logf("logdist: %d", got)
	if want != got {
		t.Fatalf("got logdist: %d, want: %d", got, want)
	}
}

func TestNode_distcmp(t *testing.T) {
	distcmpBig := func(target, a, b []byte) int {
		tbig := new(big.Int).SetBytes(target)
		abig := new(big.Int).SetBytes(a)
		bbig := new(big.Int).SetBytes(b)
		return new(big.Int).Xor(tbig, abig).Cmp(new(big.Int).Xor(tbig, bbig))
	}
	var (
		a = [2]byte{1,2}
		b = [2]byte{3,4}
		c = [2]byte{3,4}
	)
	if n := distcmp(a[:],a[:],a[:]); n != 0{
		t.Fatalf("got self logdist: %d, want: 0", n)
	}
	want := distcmpBig(a[:],b[:],c[:])
	got := distcmp(a[:],b[:],c[:])
	if want != got {
		t.Fatalf("got distcmp: %d, want: %d", got, want)
	}
}

