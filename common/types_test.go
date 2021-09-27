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

package common

import (
	"testing"
	"time"
)

func TestStrB58ToAddress(t *testing.T) {

	//var timeout = time.NewTimer(0)
	//defer timeout.Stop()
	//<-timeout.C

	timeout := time.After(5 * time.Second)
	out:
	for {
		select {
			case <-timeout:
				t.Logf("time out")
				break out
		}
	}
}

func TestIsZero(t *testing.T) {
	timeout := time.NewTimer(0) // timer to dump a non-responsive active peer
	<-timeout.C                 // timeout channel should be initially empty
	defer timeout.Stop()
	timeout.Reset(5 * time.Second)
	out:
	for {
		select {
		case <-timeout.C:
			t.Logf("time out")
			break out
		}
	}
}
