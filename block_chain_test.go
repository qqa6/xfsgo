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
	"xfsgo/assert"
	"xfsgo/storage/badger"
)

func TestBlockChain_GetBlockHashes(t *testing.T) {
	var stateDb = badger.New("./d0")
	var chainDb = badger.New("./d1")
	var extraDb = badger.New("./d2")
	defer func() {
		if err := stateDb.Close(); err != nil {
			t.Fatal(err)
		}
		if err := chainDb.Close(); err != nil {
			t.Fatal(err)
		}
		if err := extraDb.Close(); err != nil {
			t.Fatal(err)
		}
	}()
	event := NewEventBus()
	genesisBlock, err := WriteTestNetGenesisBlock(stateDb, chainDb)
	assert.Error(t, err)
	bc, err := NewBlockChain(stateDb, chainDb, extraDb, event)
	assert.Error(t, err)
	last := bc.CurrentBlock()
	assert.HashEqual(t, genesisBlock.Hash(), last.Hash())
	t.Logf("%s\n", last)
}

func TestBlockChain_InsertChain(t *testing.T) {
	type args struct {
		block *Block
	}
	tests := []struct {
		name    string
		bc      *BlockChain
		args    args
		want    bool
		wantErr bool
	}{
		// TODO: Add test cases.

	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.bc.InsertChain(tt.args.block)
			if (err != nil) != tt.wantErr {
				t.Errorf("BlockChain.InsertChain() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("BlockChain.InsertChain() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlockChain_removeOrphanBlock(t *testing.T) {
	type args struct {
		orphan *orphanBlock
	}
	tests := []struct {
		name string
		b    *BlockChain
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.b.removeOrphanBlock(tt.args.orphan)
		})
	}
}

func TestBlockChain_addOrphanBlock(t *testing.T) {
	type args struct {
		block *Block
	}
	tests := []struct {
		name string
		b    *BlockChain
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.b.addOrphanBlock(tt.args.block)
		})
	}
}
