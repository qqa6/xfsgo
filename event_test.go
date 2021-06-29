package xblockchain

import (
	"testing"
)

type tsint int
func TestEventBus_Publish(t *testing.T) {
	eb := NewEventBus()
	sub := eb.Subscript(tsint(0))
	go func() {
		for i:=0;i<100;i++ {
			eb.Publish("tsint(i)")
		}
	}()
	for {
		select {
		case c := <-sub.Chan():
			t.Logf("ccc: %v", c)
		}
	}
}