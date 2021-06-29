package badger

import (
	"bytes"
	"fmt"
	"testing"
)

func TestStorage_Set(t *testing.T) {
	storage := New("./data0")
	defer storage.Close()
	key := fmt.Sprintf("c%d",10089779)
	data := []byte("abc")
	err := storage.Set(key, data)
	if err != nil {
		t.Fatal(err)
	}
	val,err := storage.Get(key)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Compare(val, data) != 0 {
		t.Fatalf("get val err, got: %v, want: %v", val, data)
	}
}

