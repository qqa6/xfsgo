// Copyright 2018 The xfsgo Authors
// This file is part of the xfsgo library.
//
// The xfsgo library is free software: you can redistribute it and/or modify
// it under the terms of the MIT Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The xfsgo library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// MIT Lesser General Public License for more details.
//
// You should have received a copy of the MIT Lesser General Public License
// along with the xfsgo library. If not, see <https://mit-license.org/>.

package xfsgo

import (
	"testing"
)

type tsint int

func TestEventBus_Publish(t *testing.T) {
	eb := NewEventBus()
	sub := eb.Subscript(tsint(0))
	go func() {
		for i := 0; i < 100; i++ {
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
