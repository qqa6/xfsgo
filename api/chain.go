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
	"bytes"
	"encoding/json"
	"math/big"
	"strconv"
	"xfsgo"
	"xfsgo/common"
	"xfsgo/crypto"
)

type ChainAPIHandler struct {
	BlockChain    *xfsgo.BlockChain
	TxPendingPool *xfsgo.TxPool
	number        int
}

type GetBlockByIdArgs struct {
	Number string `json:"number"`
	Counts string `json:"count"`
}

type GetBlockByHashArgs struct {
	Hash string `json:"hash"`
}

type GetTxbyBlockNumArgs struct {
	Number string `json:"number"`
}
type GetTxbyBlockHashArgs struct {
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
	Number string `json:"number"`
	Count  string `json:"count"`
}

type GetBlockHeaderByHashArgs struct {
	Hash string `json:"hash"`
}

type GetBlockSectionArgs struct {
	From  string `json:"from"`
	Count string `json:"count"`
}

type GetBlocksArgs struct {
	Blocks string `json:"blocks"`
}

type ProgressBarArgs struct {
	Number int `json:"number"`
}

func (receiver *ChainAPIHandler) GetBlockByNumber(args GetBlockByIdArgs, resp *GetBlocks) error {
	var number *big.Int
	var count *big.Int
	if args.Number == "" {
		number = big.NewInt(0)
	} else {
		number, _ = new(big.Int).SetString(args.Number, 0)
	}

	if args.Counts == "" {
		b := receiver.BlockChain.GetHead()
		count = new(big.Int).SetUint64(b.Height())
	} else {
		count, _ = new(big.Int).SetString(args.Counts, 0)
	}

	data := receiver.BlockChain.GetBlockSection(number.Uint64(), count.Uint64())
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

func (receiver *ChainAPIHandler) Head(_ EmptyArgs, head *GetBlockByNumberBlockHeader) error {
	b := receiver.BlockChain.GetHead()
	header := NewBlockByNumBlockHeader(b.Header, b.Hash())
	*head = *header
	return nil
}

func (receiver *ChainAPIHandler) GetBlockHeaderByNumber(args GetBlockHeaderByNumberArgs, resp *BlockHeaders) error {
	var number *big.Int
	var count *big.Int
	if args.Number == "" {
		number = big.NewInt(0)
	} else {
		number, _ = new(big.Int).SetString(args.Number, 0)
	}

	if args.Count == "" {
		b := receiver.BlockChain.GetHead()
		count = new(big.Int).SetUint64(b.Height())
	} else {
		count, _ = new(big.Int).SetString(args.Count, 0)
	}

	data := receiver.BlockChain.GetBlockSection(number.Uint64(), count.Uint64())
	blockHeaders := make([]*GetBlockByNumberBlockHeader, 0)

	if len(data) == 0 {
		*resp = blockHeaders
		return nil
	}
	for _, v := range data {
		blockHeader := NewBlockByNumBlockHeader(v.Header, v.Hash())
		blockHeaders = append(blockHeaders, blockHeader)
	}
	*resp = blockHeaders

	return nil
}

func (receiver *ChainAPIHandler) GetBlockHeaderByHash(args GetBlockHeaderByHashArgs, blockHeader *GetBlockByNumberBlockHeader) error {
	if args.Hash == "" {
		return xfsgo.NewRPCError(-1006, "Parameter cannot be empty")
	}
	data, Hash := receiver.BlockChain.GetBlockHeaderByHash(common.Hex2Hash(args.Hash))
	result := NewBlockByNumBlockHeader(data, Hash)
	*blockHeader = *result
	return nil
}

func (receiver *ChainAPIHandler) GetBlockByHash(args GetBlockByHashArgs, resp *GetBlockByNumberBlock) error {
	if args.Hash == "" {
		return xfsgo.NewRPCError(-1006, "Parameter cannot be empty")
	}
	b := receiver.BlockChain.GetBlockByHash(common.Hex2Hash(args.Hash))
	header := NewBlockByNumBlockHeader(b.Header, b.Hash())
	result := NewBlockByNumberBlock(b, header)
	*resp = *result
	return nil

}

func (receiver *ChainAPIHandler) GetTxbyBlockNum(args GetTxbyBlockNumArgs, resp *TransferObjs) error {
	if args.Number == "" {
		return xfsgo.NewRPCError(-1006, "Parameter cannot be empty")
	}

	blockHeight, ok := new(big.Int).SetString(args.Number, 0)

	if !ok {
		return xfsgo.NewRPCError(-1006, "string to big.Int error")
	}

	b := receiver.BlockChain.GetBlockByNumber(blockHeight.Uint64())
	result := make([]*TransferObj, 1)
	for _, item := range b.Transactions {
		TranObj := NewTransferObj(item)
		result = append(result, TranObj)
	}
	*resp = result
	return nil
}

func (receiver *ChainAPIHandler) GetTxbyBlockHash(args GetTxbyBlockHashArgs, resp *TransferObjs) error {
	if args.Hash == "" {
		return xfsgo.NewRPCError(-1006, "Parameter cannot be empty")
	}
	b := receiver.BlockChain.GetBlockByHash(common.Hex2Hash(args.Hash))
	result := make([]*TransferObj, 1)
	for _, item := range b.Transactions {
		TranObj := NewTransferObj(item)
		result = append(result, TranObj)
	}
	*resp = result
	return nil
}

func (receiver *ChainAPIHandler) GetReceiptByHash(args GetReceiptByHashArgs, resp *xfsgo.Receipt) error {
	if args.Hash == "" {
		return xfsgo.NewRPCError(-1006, "Parameter cannot be empty")
	}
	data := receiver.BlockChain.GetReceiptByHash(common.Hex2Hash(args.Hash))
	*resp = *data
	return nil
}

func (receiver *ChainAPIHandler) GetTransaction(args GetTransactionArgs, resp *TransferObj) error {
	if args.Hash == "" {
		return xfsgo.NewRPCError(-1006, "Parameter cannot be empty")
	}
	ID := common.Hex2Hash(args.Hash)
	data := receiver.BlockChain.GetTransaction(ID)
	result := NewTransferObj(data)
	*resp = *result
	return nil
}

// func (receiver *ChainAPIHandler) GetBlockSection(args GetBlockSectionArgs, resp *GetBlocks) error {
// 	if args.Count == "" && args.From == "" {
// 		return xfsgo.NewRPCError(-1006, "Parameter cannot be empty")
// 	}

// 	numbersForm, ok := new(big.Int).SetString(args.From, 0)
// 	if !ok {
// 		return xfsgo.NewRPCError(-1006, "string to big.Int error")
// 	}
// 	numbersCount, ok := new(big.Int).SetString(args.Count, 0)
// 	if !ok {
// 		return xfsgo.NewRPCError(-1006, "string to big.Int error")
// 	}
// 	if numbersCount.Uint64() == uint64(0) {
// 		b := receiver.BlockChain.GetHead()
// 		numbersCount.SetUint64(b.Height())
// 	}
// 	data := receiver.BlockChain.GetBlockSection(numbersForm.Uint64(), numbersCount.Uint64())
// 	GetBlockByNumberBlock := make([]*GetBlockByNumberBlock, 0)

// 	if len(data) == 0 {
// 		*resp = GetBlockByNumberBlock
// 		return nil
// 	}
// 	for _, v := range data {
// 		blockHeader := NewBlockByNumBlockHeader(v.Header, v.Hash())
// 		blocks := NewBlockByNumberBlock(v, blockHeader)
// 		GetBlockByNumberBlock = append(GetBlockByNumberBlock, blocks)
// 	}
// 	*resp = GetBlockByNumberBlock
// 	return nil
// }

func (receiver *ChainAPIHandler) ExportBlock(args GetBlockSectionArgs, resp *string) error {
	numbersForm, ok := new(big.Int).SetString(args.From, 0)
	if !ok {
		return xfsgo.NewRPCError(-1006, "string to big.Int error")
	}
	numbersCount, ok := new(big.Int).SetString(args.Count, 0)
	if !ok {
		return xfsgo.NewRPCError(-1006, "string to big.Int error")
	}

	if numbersCount.Uint64() == uint64(0) {
		b := receiver.BlockChain.GetHead()
		numbersCount.SetUint64(b.Height())
	}

	if numbersForm.Uint64() >= numbersCount.Uint64() { // Export all
		b := receiver.BlockChain.GetHead()
		header := NewBlockByNumBlockHeader(b.Header, b.Hash())
		result := NewBlockByNumberBlock(b, header)
		numbersCount.SetUint64(result.Header.Height) // Get the maximum height of the current block
	}

	data := receiver.BlockChain.GetBlockSection(numbersForm.Uint64(), numbersCount.Uint64())

	encodeByte, err := json.Marshal(data)
	if err != nil {
		return xfsgo.NewRPCErrorCause(-32001, err)
	}
	key := crypto.MD5Str(encodeByte)
	encryption := crypto.AesEncrypt(encodeByte, key)
	var bt bytes.Buffer
	bt.WriteString(key)
	bt.WriteString(encryption)
	respStr := bt.String()
	*resp = respStr
	return nil

}

func (receiver *ChainAPIHandler) ImportBlock(args GetBlocksArgs, resp *string) error {
	if args.Blocks == "" {
		return xfsgo.NewRPCError(-1006, "to Blocks file path not be empty")
	}
	key := args.Blocks[:32]
	decodeBuf := args.Blocks[32:]
	decryption := crypto.AesDecrypt(decodeBuf, key)
	blockChain := make([]*xfsgo.Block, 0)
	err := json.Unmarshal([]byte(decryption), &blockChain)
	if err != nil {
		return xfsgo.NewRPCErrorCause(-32001, err)
	}
	receiver.number = len(blockChain) - 1

	for _, item := range blockChain {
		if err := receiver.BlockChain.InsertChain(item); err != nil {
			continue
		}
	}
	*resp = "Import complete"
	return nil
}

func (receiver *ChainAPIHandler) ProgressBar(_ EmptyArgs, resp *string) error {
	b := receiver.BlockChain.GetHead()
	height := strconv.FormatInt(int64(b.Header.Height), 10)
	total := strconv.Itoa(receiver.number)
	result := height + "/" + total + "(total)"
	*resp = result
	return nil
}
