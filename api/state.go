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
	"xfsgo"
	"xfsgo/common"
	"xfsgo/storage/badger"
)

type StateAPIHandler struct {
	StateDb    *badger.Storage
	BlockChain *xfsgo.BlockChain
}

type GetAccountArgs struct {
	RootHash string `json:"root_hash"`
	Address  string `json:"address"`
}

type GetBalanceArgs struct {
	RootHash string `json:"root_hash"`
	Address  string `json:"address"`
}

func (state *StateAPIHandler) GetBalance(args GetBalanceArgs, resp *string) error {
	if args.RootHash == "" {
		rootHash := state.BlockChain.CurrentBlock().StateRoot()
		args.RootHash = rootHash.Hex()
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
		*resp = "0"
		return nil
	}
	*resp = data.GetBalance().String()
	return nil

}

func (state *StateAPIHandler) GetAccount(args GetAccountArgs, resp *StateObjResp) error {

	if args.RootHash == "" {
		rootHash := state.BlockChain.CurrentBlock().StateRoot()
		args.RootHash = rootHash.Hex()
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
		result := &StateObjResp{
			Address: args.Address,
		}
		*resp = *result
		return nil
	}

	addressEqual := data.GetAddress()
	var addrStr string
	if !addressEqual.Equals(common.Address{}) {
		addrStr = addressEqual.B58String()
	}
	ist := data.GetBalance()
	// if ist == nil {
	// 	ist = new(big.Int).SetUint64(0)
	// }
	// bal2Byte := ist.Bytes()
	result := &StateObjResp{
		Balance: ist.String(),
		Nonce:   data.GetNonce(),
		Address: addrStr,
	}
	*resp = *result
	return nil
}
