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

package lru

import (
	"container/list"
	"fmt"
	"sync"
)

const cacheKeySize = 32

type Cache struct {
	mu     sync.Mutex
	size   int
	items  map[[cacheKeySize]byte]*list.Element
	access *list.List
}

type cacheData struct {
	key   [cacheKeySize]byte
	extra [cacheKeySize]byte
	val   []byte
}

func NewCache(size int) *Cache {
	return &Cache{
		size:   size,
		items:  make(map[[cacheKeySize]byte]*list.Element, size),
		access: list.New(),
	}
}

func (c *Cache) Get(key [cacheKeySize]byte) ([]byte, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	elem, ok := c.items[key]
	if !ok {
		return nil, false
	}

	c.access.MoveToFront(elem)

	return elem.Value.(*cacheData).val, ok
}

func (c *Cache) GetOrPut(key [cacheKeySize]byte, val []byte) ([]byte, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	elem, ok := c.items[key]

	if ok {
		val = elem.Value.(*cacheData).val
		c.access.MoveToFront(elem)
	} else {
		c.items[key] = c.access.PushFront(&cacheData{
			key: key,
			val: val,
		})
		for len(c.items) > c.size {
			back := c.access.Back()
			info := back.Value.(*cacheData)
			delete(c.items, info.key)
			c.access.Remove(back)
		}
	}
	return val, ok
}

func (c *Cache) Put(key [cacheKeySize]byte, val []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()

	elem, ok := c.items[key]
	if ok {
		elem.Value.(*cacheData).val = val
		c.access.MoveToFront(elem)
	} else {
		c.items[key] = c.access.PushFront(&cacheData{
			key: key,
			val: val,
		})
		for len(c.items) > c.size {
			back := c.access.Back()
			info := back.Value.(*cacheData)
			delete(c.items, info.key)
			c.access.Remove(back)
		}
	}
}

func (c *Cache) Remove(key [cacheKeySize]byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	elem, ok := c.items[key]
	if ok {
		delete(c.items, key)
		c.access.Remove(elem)
	}
}

func (c *Cache) GetWithExtra(pkey [cacheKeySize]byte) ([]byte, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	fmt.Printf("pkey:%v\n", pkey)

	for _, elem := range c.items {
		if elem.Value.(*cacheData).extra == pkey {
			return elem.Value.(*cacheData).val, true
		}
	}

	return nil, false
}

func (c *Cache) GetOrPutWithExtra(key [cacheKeySize]byte, val []byte, pkey [cacheKeySize]byte) ([]byte, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	elem, ok := c.items[key]

	if ok {
		val = elem.Value.(*cacheData).val
		c.access.MoveToFront(elem)
	} else {
		c.items[key] = c.access.PushFront(&cacheData{
			key:   key,
			val:   val,
			extra: pkey,
		})
		for len(c.items) > c.size {
			back := c.access.Back()
			info := back.Value.(*cacheData)
			delete(c.items, info.key)
			c.access.Remove(back)
		}
	}
	return val, ok
}

func (c *Cache) PutWithExtra(key [cacheKeySize]byte, val []byte, prekey [cacheKeySize]byte) {
	c.mu.Lock()
	defer c.mu.Unlock()

	elem, ok := c.items[key]
	if ok {
		elem.Value.(*cacheData).val = val
		elem.Value.(*cacheData).extra = prekey
		c.access.MoveToFront(elem)
	} else {
		c.items[key] = c.access.PushFront(&cacheData{
			key:   key,
			val:   val,
			extra: prekey,
		})
		for len(c.items) > c.size {
			back := c.access.Back()
			info := back.Value.(*cacheData)
			delete(c.items, info.key)
			c.access.Remove(back)
		}
	}
}
