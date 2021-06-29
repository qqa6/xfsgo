package xblockchain

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"log"
	"testing"
)

func TestScriptVM_ExecScript(t *testing.T) {

	bin,err := BuildScriptBin("")
	if err != nil {
		t.Fatal(err)
	}
	log.Printf("bin: %v\n",bin)
	vm := NewScriptVM()
	r := vm.ExecScriptBin(bin)
	log.Printf("exec result: %v\n",r)
}

func TestParseScriptBin2Str(t *testing.T) {
	scriptEnc := "ZFhBb0tUdHdhMmdvS1R0d2RYTm9LQ0kwTTBOVlYxYzRVRVJLTTIxRFlsQlplSGROZFZSRVVGUnJhR0pNU0hsRlFWWkRSVlZNVTFOVWMxRnVOQ0lwTzJWeEtDazdZMmhsWTJ0VGFXY29LUT09"
	scriptBin,err := base64.StdEncoding.DecodeString(scriptEnc)
	if err != nil {
		t.Fatal(err)
	}
	scriptDec,err := ParseScriptBin2Str(scriptBin)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf(scriptDec)
}

func TestScriptVM_ExecSing(t *testing.T) {

	script0 := `push("a");push("aSq9DsNNvGhYxYyqA9wd2eduEAZ5AXWgJTbTFxbG9Hor4qvs2FhTX1hirjDGMnAFdG5xw4d4dsrrNGTQsbHpDUeZYEY3pJLRY3qYhB65XxGMch51cNj95SXBCET4");`
	script1 := `up();pkh();push("Hz7HdfcdDk1t744YNwhyEyHoBr64skPwTaAtPq1JCTTs");eq()`
	scriptsum := fmt.Sprintf("%s%s",script0,script1)
	bin,err := BuildScriptBin(scriptsum)
	if err != nil {
		t.Fatal(err)
	}
	log.Printf("script: %s\n",scriptsum)
	log.Printf("bin: %v\n",bin)
	vm := NewScriptVM()
	r := vm.ExecScriptBin(bin)
	log.Printf("exec result: %v\n",r)
	_ = r
}

func TestRunningStack_Pop(t *testing.T) {
	s := newRunningStack()
	s.push([]byte("a"))
	s.push([]byte("b"))
	s.push([]byte("c"))
	s.push([]byte("d"))
	if bytes.Compare(s.pop(),[]byte("d")) != 0 {
		t.Fatalf("stack pop result is abnormal")
	}
}