package xblockchain

import "testing"

func TestMem(t *testing.T) {
	tx, err:= NewCoinBaseTransaction(DefaultGenesisBlockOpts().Address, "aaa")
	if err != nil {
		t.Fatal(err)
	}
	eb := NewEventBus()
	sub := eb.Subscript(TxPreEvent{})
	txPool := NewTxPool(eb, nil)
	go func() {
		if err = txPool.AddTx(tx); err != nil {
			t.Error(err)
			return
		}
	}()
	for {
		select {
		case c := <-sub.Chan():
			a := c.(TxPreEvent)
			t.Logf("t: %s", a.Tx.ID.Hex())
		}
	}
}
