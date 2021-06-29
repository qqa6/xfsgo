package xblockchain

import (
	"encoding/json"
	"testing"
)

func TestTransaction_MarshalJSON(t *testing.T) {
	tx,err := NewCoinBaseTransaction("123","123")
	if err != nil {
		t.Fatal(err)
		return
	}
	jsonbs,err := json.Marshal(tx)
	if err != nil {
		t.Fatal(err)
		return
	}
	t.Logf("json string: %s", string(jsonbs))
}