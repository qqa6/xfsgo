package ahash

import (
	"encoding/json"
	"testing"
)

func TestSHA256(t *testing.T) {
	obj := &struct {
		To string `json:"to"`
		Value string `json:"value"`
		Gaslimit string `json:"gaslimit"`
		Gasprice string `json:"gasprice"`
		Signature *string `json:"signature"`
	}{
		To: "1",
		Value: "1",
		Gaslimit: "100",
		Gasprice: "100",
		Signature: nil,
	}
	jsondata, err := json.Marshal(obj)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("jsonstring: %s\n", string(jsondata))
	hashdata := SHA256(jsondata)
	t.Logf("hashdata: %x\n", hashdata)
}
