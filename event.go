package xblockchain

import (
	"reflect"
	"sync"
)

type TxPreEvent struct {
	Tx *Transaction
}

type TxPostEvent struct {
	Tx *Transaction
}

type ChainHeadEvent struct {
	Block *Block
}
type NewMinerBlockEvent struct {
	Block *Block
}

type EventBus struct {
	subs map[reflect.Type][]chan interface{}
	rw sync.RWMutex
}

type Subscription struct {
	c chan interface{}
}

func (s *Subscription) Chan() chan interface{} {
	return s.c
}

func NewEventBus() *EventBus {
	return &EventBus{
		subs: make(map[reflect.Type][]chan interface{}),
	}
}

func (e *EventBus) Subscript(t interface{}) *Subscription {
	e.rw.Lock()
	defer e.rw.Unlock()
	rtyp := reflect.TypeOf(t)
	subtion := &Subscription{
		c: make(chan interface{}),
	}
	if prev, found := e.subs[rtyp]; found {
		e.subs[rtyp] = append(prev, subtion.c)
	}else {
		e.subs[rtyp] = append([]chan interface{}{}, subtion.c)
	}
	return subtion
}

func (e *EventBus) Publish(data interface{}) {
	e.rw.RLock()
	defer e.rw.RUnlock()
	rtyp := reflect.TypeOf(data)
	if cs, found := e.subs[rtyp]; found {
		go func(d interface{}, cs []chan interface{}) {
			for _, ch := range cs {
				ch <- d
			}
		}(data, cs)
	}
}
