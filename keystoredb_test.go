package xfsgo

import (
	"fmt"
	"log"
	"testing"

	"github.com/dgraph-io/badger/v3"
)

func TestKeyDBAll(t *testing.T) {

	opts := badger.DefaultOptions("./cmd/xfsgo/d0/keys")
	db, err := badger.Open(opts)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	err = db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			k := item.Key()
			fmt.Printf("key=%v\n", string(k))
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

}
