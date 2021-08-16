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
	"testing"
	"xfsgo/common/ahash"

	"github.com/magiconair/properties/assert"
)

func str2key(data string) [32]byte {
	var key [32]byte
	hash := ahash.SHA256([]byte(data))
	copy(key[:], hash)
	return key
}
func TestLruCache_Put(t *testing.T) {
	cache := NewCache(2)
	cache.Put(str2key("a"), []byte("b"))
	cache.Put(str2key("c"), []byte("d"))
	cache.Put(str2key("e"), []byte("f"))
	_, ok := cache.Get(str2key("a"))
	assert.Equal(t, len(cache.items), 2)
	assert.Equal(t, cache.access.Len(), 2)
	assert.Equal(t, cache.size, 2)
	assert.Equal(t, ok, false)
}

func TestLruCache_GetOrPut(t *testing.T) {
	cache := NewCache(2)
	cache.GetOrPut(str2key("a"), []byte("b"))
	got, ok := cache.Get(str2key("a"))
	assert.Equal(t, len(cache.items), 1)
	assert.Equal(t, cache.access.Len(), 1)
	assert.Equal(t, ok, true)
	assert.Equal(t, got, []byte("b"))
}

func TestLruCache_Remove(t *testing.T) {
	cache := NewCache(2)
	cache.Put(str2key("a"), []byte("b"))
	cache.Put(str2key("c"), []byte("d"))
	cache.Remove(str2key("a"))
	_, ok := cache.Get(str2key("a"))
	assert.Equal(t, len(cache.items), 1)
	assert.Equal(t, cache.access.Len(), 1)
	assert.Equal(t, ok, false)
}
