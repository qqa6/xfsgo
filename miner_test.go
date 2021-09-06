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
	"fmt"
	"testing"
)

func TestMiner_Run(t *testing.T) {
	a := make([]int, 0)
	a = append(a, []int{}...)
	tmp := make([]int, 0)
	max := 5
	tmp = append(tmp, 0)
	if len(a) > 0 && len(a) < max-1 {
		tmp = append(tmp, a...)
	} else if len(a) > max-1 {
		tmp = append(tmp, a[:max-1]...)
		a = a[max-1:]
	}
	t.Logf("")
	//var max = 5
	//txs := make([]int, 0)
	//q := make([]int, 0)
	//
	//q = append(q, 1)
	//q = append(q, 2)
	//q = append(q, 3)
	//q = append(q, 4)
	//q = append(q, 5)
	//q = append(q, 6)
	//q = append(q, 7)
	//q = append(q, 8)
	//q = append(q, 9)
	//txs = append(txs, q...)
	//tmp := make([]int, 0)
	//tmp = append(tmp, 0)
	//if len(txs) > 0 && len(txs) < max {
	//	tmp = append(tmp, txs...)
	//}else if len(txs) > max {
	//	tmp = append(tmp, txs[:max]...)
	//	txs = txs[max:]
	//}

}

func TestRunning(t *testing.T) {
	//var mu0 sync.Mutex
	//var mu1 sync.Mutex
	//pool := NewTXPendingPool(100)
	//mpool := make([]string,0)
	//mu.Lock()
	//mu0.Lock()
	//go func() {
	//	for i:=0;;i++{
	//		pool.Push([]byte(fmt.Sprintf("c%d",i)))
	//		time.Sleep(5 * time.Second)
	//	}
	//}()
	//go func() {
	//	for {
	//		msg := pool.Pop()
	//		fmt.Printf("i: %s\n", string(msg))
	//		time.Sleep(5 * time.Second)
	//	}
	//}()
	//wait := make(chan struct{})
	//<- wait
	//mu0.Lock()
	//var num *int
	//for i:=0;i<10;i++{
	//	in := i
	//	mu.Lock()
	//
	//}
	//mu.Lock()
	//
	//go func() {
	//	run1("c1")
	//	mu.Unlock()
	//}()
	//mu.Lock()
	//go func() {
	//	run1("c2")
	//	mu.Unlock()
	//}()
	//mu.Lock()

}

func TestName(t *testing.T) {
	//s := []byte("abcdef")
	//fmt.Printf("len(s): %d\n",len(s))
	//s = s[0:1]
	//fmt.Printf("len(s): %d\n",len(s))
	fmt.Printf("len(s): %v\n", 10 < 10)
}

func run0(c string, i int) {
	fmt.Printf("run[%s]: %d\n", c, i)
}

func run1(c string) {
	for i := 0; i < 10; i++ {
		fmt.Printf("run[%s]: %d\n", c, i)
	}
}
