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
	"encoding/json"
	"xfsgo"
	"xfsgo/common"
	"xfsgo/common/urlsafeb64"
)

// var (
// 	notFoundBlockErr = xfsgo.NewRPCError(-32001, "Not found block")
// )

type ChainAPIHandler struct {
	BlockChain    *xfsgo.BlockChain
	TxPendingPool *xfsgo.TxPool
}

type GetBlockByIdArgs struct {
	Number json.Number `json:"number"`
}

type GetBlockByHashArgs struct {
	Hash string `json:"hash"`
}

type GetBalanceOfAddressArgs struct {
	Address string `json:"address"`
}

type GetTransactionArgs struct {
	Hash string `json:"hash"`
}

type GetReceiptByHashArgs struct {
	Hash string `json:"hash"`
}

type GetBlockHeaderByNumberArgs struct {
	Height json.Number `json:"height"`
}

type GetBlockHeaderByHashArgs struct {
	Hash string `json:"hash"`
}

type SendRawTransactionArgs struct {
	Data string `json:"data"`
}

type GetBlockSectionArgs struct {
	From  json.Number `json:"from"`
	Count json.Number `json:"count"`
}

func (receiver *ChainAPIHandler) GetBlockByNumber(args GetBlockByIdArgs, block *GetBlockByNumberBlock) error {
	numbers, err := common.Uint64s(args.Number)
	if err != nil {
		return xfsgo.NewRPCErrorCause(-32001, err)
	}
	b := receiver.BlockChain.GetBlockByNumber(numbers)
	header := NewBlockByNumBlockHeader(b.Header, b.Hash())
	result := NewBlockByNumberBlock(b, header)
	*block = *result
	return nil
}

func (receiver *ChainAPIHandler) Head(_ EmptyArgs, block *GetBlockByNumberBlock) error {
	b := receiver.BlockChain.GetHead()
	header := NewBlockByNumBlockHeader(b.Header, b.Hash())
	result := NewBlockByNumberBlock(b, header)
	*block = *result
	return nil
}

func (receiver *ChainAPIHandler) GetBlockHeaderByNumber(args GetBlockHeaderByNumberArgs, blockHeader *GetBlockByNumberBlockHeader) error {
	numbers, err := common.Uint64s(args.Height)
	if err != nil {
		return xfsgo.NewRPCErrorCause(-32001, err)
	}
	data, Hash := receiver.BlockChain.GetBlockHeaderByNumber(numbers)
	result := NewBlockByNumBlockHeader(data, Hash)
	*blockHeader = *result
	return nil
}

func (receiver *ChainAPIHandler) GetBlockHeaderByHash(args GetBlockHeaderByHashArgs, blockHeader *GetBlockByNumberBlockHeader) error {
	data, Hash := receiver.BlockChain.GetBlockHeaderByHash(common.Hex2Hash(args.Hash))
	result := NewBlockByNumBlockHeader(data, Hash)
	*blockHeader = *result
	return nil
}

func (receiver *ChainAPIHandler) GetBlockByHash(args GetBlockByHashArgs, block *GetBlockByNumberBlock) error {
	b := receiver.BlockChain.GetBlockByHash(common.Hex2Hash(args.Hash))
	header := NewBlockByNumBlockHeader(b.Header, b.Hash())
	result := NewBlockByNumberBlock(b, header)
	*block = *result
	return nil

}

func (receiver *ChainAPIHandler) GetReceiptByHash(args GetReceiptByHashArgs, receipt *xfsgo.Receipt) error {
	data := receiver.BlockChain.GetReceiptByHash(common.Hex2Hash(args.Hash))
	*receipt = *data
	return nil
}

func (receiver *ChainAPIHandler) GetTransaction(args GetTransactionArgs, resp *TransferObj) error {
	ID := common.Hex2Hash(args.Hash)
	data := receiver.BlockChain.GetTransaction(ID)
	result := NewTransferObj(data)
	*resp = *result
	return nil
}

func (receiver *ChainAPIHandler) SendRawTransaction(args SendRawTransactionArgs, resp *TransferObj) error {
	databytes, err := urlsafeb64.Decode(args.Data)
	if err != nil {
		return xfsgo.NewRPCErrorCause(-32001, err)
	}
	var tx *xfsgo.Transaction
	err = json.Unmarshal(databytes, &tx)
	if err != nil {
		return xfsgo.NewRPCErrorCause(-32001, err)
	}

	err = receiver.TxPendingPool.Add(tx)
	if err != nil {
		return xfsgo.NewRPCErrorCause(-32001, err)
	}
	result := NewTransferObj(tx)
	*resp = *result

	return nil
}

func (receiver *ChainAPIHandler) GetBlockSection(args GetBlockSectionArgs, resp *GetBlocks) error {

	numbersForm, err := common.Uint64s(args.From)
	if err != nil {
		return xfsgo.NewRPCErrorCause(-32001, err)
	}
	numbersCount, err := common.Uint64s(args.Count)
	if err != nil {
		return xfsgo.NewRPCErrorCause(-32001, err)
	}
	data := receiver.BlockChain.GetBlockSection(numbersForm, numbersCount)
	GetBlockByNumberBlock := make([]*GetBlockByNumberBlock, 0)
	if len(data) == 0 {
		*resp = GetBlockByNumberBlock
		return nil
	}
	for _, v := range data {
		blockHeader := NewBlockByNumBlockHeader(v.Header, v.Hash())
		blocks := NewBlockByNumberBlock(v, blockHeader)
		GetBlockByNumberBlock = append(GetBlockByNumberBlock, blocks)
	}
	*resp = GetBlockByNumberBlock
	return nil
}
