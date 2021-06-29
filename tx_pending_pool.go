package xblockchain

import (
	"sync"
	"time"
)

type TXPendingPool struct {
	txs []*Transaction
	max int
	ch chan []byte
	readMutex sync.Mutex
	writeMutex sync.Mutex
}

func NewTXPendingPool(max int) *TXPendingPool {
	s := &TXPendingPool{}
	s.txs = make([]*Transaction,0)
	s.max = max
	return s
}

func (s *TXPendingPool) Push(data *Transaction) {
	s.writeMutex.Lock()
	defer s.writeMutex.Unlock()
	for len(s.txs) > s.max {
		time.Sleep(1000 * time.Millisecond)
	}
	s.txs = append(s.txs, data)
}

func (s *TXPendingPool) Pop() *Transaction {
	s.readMutex.Lock()
	defer s.readMutex.Unlock()
	for len(s.txs) == 0 {
		time.Sleep(1000 * time.Millisecond)
	}
	elms := s.txs[0]
	s.txs = s.txs[1:]
	return elms
}

func (s *TXPendingPool) PopAll() []*Transaction {
	s.readMutex.Lock()
	defer s.readMutex.Unlock()
	for len(s.txs) == 0 {
		time.Sleep(1000 * time.Millisecond)
	}
	elms := s.txs
	s.txs = s.txs[0:0]
	return elms
}

func (s *TXPendingPool) PopMin(min int) []*Transaction {
	s.readMutex.Lock()
	defer s.readMutex.Unlock()
	for len(s.txs) < min {
		time.Sleep(1000 * time.Millisecond)
	}
	elms := s.txs
	s.txs = s.txs[0:0]
	return elms
}

func (s *TXPendingPool) PopCount(count int) []*Transaction {
	s.readMutex.Lock()
	defer s.readMutex.Unlock()
	for len(s.txs) < count {
		time.Sleep(1000 * time.Millisecond)
	}
	elms := s.txs[0:count]
	s.txs = s.txs[count:]
	return elms
}


func (s *TXPendingPool) Count() int {
	s.readMutex.Lock()
	defer s.readMutex.Unlock()
	return len(s.txs)
}

func (s *TXPendingPool) List() []*Transaction {
	s.readMutex.Lock()
	defer s.readMutex.Unlock()
	return s.txs
}

func (s *TXPendingPool) Foreach(fn func(tx *Transaction)) {
	s.readMutex.Lock()
	defer s.readMutex.Unlock()
	for i := 0; i < len(s.txs); i++ {
		tx := s.txs[i]
		fn(tx)
	}
}

func (s *TXPendingPool) Empty() bool {
	s.readMutex.Lock()
	defer s.readMutex.Unlock()
	return len(s.txs) == 0
}
func (s *TXPendingPool) Clear() {
	s.writeMutex.Lock()
	defer s.writeMutex.Unlock()
	s.txs = make([]*Transaction,0)
}