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

package api

import (
	"math/big"
	"xfsgo"
	"xfsgo/common"
	"xfsgo/storage/badger"
)

type StateAPIHandler struct {
	StateDb *badger.Storage
	// State *xfsgo.StateTree
}

type GetStateObjArgs struct {
	RootHash string `json:"root_hash"`
	Address  string `json:"address"`
}

type StateObj struct {
	Address string   `json:"address"`
	Balance *big.Int `json:"balance"`
	Nonce   uint64   `json:"nonce"`
}

func (state *StateAPIHandler) GetStateObj(args GetStateObjArgs, resp *StateObj) error {

	if args.RootHash == "" {
		return xfsgo.NewRPCError(-32601, "Root hash not found")
	}
	if args.Address == "" {
		return xfsgo.NewRPCError(-32601, "Address not found")
	}
	roothash := common.Hex2Hash(args.RootHash)

	rootHashByte := roothash.Bytes()

	stateTree := xfsgo.NewStateTree(state.StateDb, rootHashByte)

	address := common.B58ToAddress([]byte(args.Address))

	data := stateTree.GetStateObj(address)

	if data == (&xfsgo.StateObj{}) || data == nil {
		result := &StateObj{}
		*resp = *result
		return nil
	}

	addressEqual := data.GetAddress()
	var addrStr string
	if !addressEqual.Equals(common.Address{}) {
		addrStr = addressEqual.B58String()
	}
	result := &StateObj{
		Balance: data.GetBalance(),
		Nonce:   data.GetNonce(),
		Address: addrStr,
	}
	*resp = *result
	return nil
}
