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
	"reflect"
	"sync"
)

// EventBus dispatches events to registered receivers. Receivers can be
// registered to handle events of certain type.
type EventBus struct {
	subs map[reflect.Type][]chan interface{}
	rw   sync.RWMutex
}

type Subscription struct {
	eb    *EventBus
	typ   reflect.Type
	index int
	c     chan interface{}
}

func (s *Subscription) Chan() chan interface{} {
	return s.c
}

func (s *Subscription) Unsubscribe() {
	s.eb.unsubscribe(s.typ, s.index)
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
		typ: rtyp,
		c:   make(chan interface{}),
		eb:  e,
	}
	if prev, found := e.subs[rtyp]; found {
		nextIndex := len(prev)
		subtion.index = nextIndex
		e.subs[rtyp] = append(prev, subtion.c)
	} else {
		subtion.index = 0
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

func (e *EventBus) unsubscribe(data interface{}, index int) {
	e.rw.Lock()
	defer e.rw.Unlock()
	rtyp := reflect.TypeOf(data)
	if old, found := e.subs[rtyp]; found {
		e.subs[rtyp] = append(old[:index], old[index+1:]...)
	}
}
