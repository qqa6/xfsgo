package uint256

import (
	"fmt"
	"testing"
)

func TestUInt256Add(t *testing.T) {
	u1 := NewUInt256("0x1")
	u2 := NewUInt256("0x2")
	want := NewUInt256("0x3")
	got := u1.Add(u2)
	if !got.Equals(want) {
		t.Fatalf("got '0x%s' want '0x%s'", got.Hexstr(true), want.Hexstr(true))
	}
}

func TestUInt256Sub(t *testing.T) {
	u1 := NewUInt256("0x3")
	u2 := NewUInt256("0x2")
	want := NewUInt256("0x1")
	got := u1.Sub(u2)
	if !got.Equals(want) {
		t.Fatalf("got '0x%s' want '0x%s'", got.Hexstr(true), want.Hexstr(true))
	}
}

func TestUInt256Xor(t *testing.T) {
	u1 := NewUInt256("0x1")
	u2 := NewUInt256("0x1")
	want := NewUInt256("0x0")
	got := u1.Xor(u2)
	if !got.Equals(want) {
		t.Fatalf("got '0x%s' want '0x%s'", got.Hexstr(true), want.Hexstr(true))
	}
}

func TestUInt256_Lsh(t *testing.T) {
	u1 := NewUInt256ByHex("0xff")
	got := u1.Lsh(24)
	want := NewUInt256("0xff000000")
	if !got.Equals(want) {
		t.Fatalf("got '0x%s' want '0x%s'", got.Hexstr(true), want.Hexstr(true))
	}
}

func TestUInt256_Rsh(t *testing.T) {
	u1 := NewUInt256ByHex("0xff000000")
	got := u1.Rsh(24)
	want := NewUInt256("0xff")
	if !got.Equals(want) {
		t.Fatalf("got '0x%s' want '0x%s'", got.Hexstr(true), want.Hexstr(true))
	}
}


func TestUInt256NumberFor(t *testing.T) {
	want := NewUInt256ByHex("0x953614b4")
	got := NewUInt256ByUInt32(uint32(2503349428))
	if !got.Equals(want) {
		t.Fatalf("got '0x%s' want '0x%s'", got.Hexstr(true), want.Hexstr(true))
	}
}

func TestNewUInt256Zero(t *testing.T) {
	u := NewUInt256Zero()
	x := u[0]
	want := byte(0)
	if x != want {
		t.Fatalf("u[0] got: '%x' want: '%x'", x, want)
	}
}

func TestNewUInt256One(t *testing.T) {
	u := NewUInt256One()
	x := u[0]
	want := byte(1)
	if x != want {
		t.Fatalf("u[0] got: '%x' want: '%x'", x, want)
	}
}

func TestUInt256_Gt(t *testing.T) {
	u1 := NewUInt256ByUInt32(uint32(234567))
	u2 := NewUInt256ByUInt32(uint32(234566))
	if !u1.Gt(u2) {
		t.Fatalf("u1 got: '0x%s' u2: '0x%s'", u1.HexstrFull(), u2.HexstrFull())
	}
}

func TestUInt256_Lt(t *testing.T) {
	u1 := NewUInt256("0x0000008de0044d7282933b4a5a1573bb9f40e0891e7a8da1d01f72ee3d239810")
	u2 := NewUInt256("0x0000010000000000000000000000000000000000000000000000000000000000")
	if !u1.Lt(u2){
		t.Fatalf("u1 got: '0x%s' u2: '0x%s'", u1.HexstrFull(), u2.HexstrFull())
	}
}

func TestUInt256_Cmp(t *testing.T) {
	u1 := NewUInt256ByUInt32(5)
	u2 := NewUInt256ByUInt32(1)
	r := u1.Cmp(u2)

	fmt.Printf("r: %d\n", r)
	//fmt.Printf("hi: %s\n", hashint.HexstrFull())
}

func TestUInt256_MarshalJSON(t *testing.T) {
	uz := NewUInt256One()
	jb,err := uz.MarshalJSON()
	if err != nil {
		t.Fatal(err)
		return
	}
	t.Logf("json: %s", string(jb))
}