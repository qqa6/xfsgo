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

package main

import (
	"fmt"
	"testing"
)

type obj struct {
	name    string
	val     string
	balance string
}

func TestA(t *testing.T) {
	a := make([]*obj, 0)
	a = append(a, &obj{
		name:    "abc1",
		val:     "def",
		balance: "123",
	})
	a = append(a, &obj{
		name:    "abc2",
		val:     "def",
		balance: "123",
	})
	a = append(a, &obj{
		name:    "abc3",
		val:     "def",
		balance: "123",
	})
	a = append(a, &obj{
		name:    "abc4",
		val:     "def",
		balance: "123",
	})
	fmt.Println("name   val      balance")
	for _, item := range a {
		fmt.Printf("%-6s", item.name)
		fmt.Printf("%-6s", item.val)
		fmt.Printf("%-6s", item.balance)
		fmt.Println()
	}
}
