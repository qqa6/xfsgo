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

type GetBlockByNumArgs struct {
	Number string `json:"number"`
}

type GetBlockByHashArgs struct {
	Hash string `json:"hash"`
}

type GetTxsByBlockNumArgs struct {
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
	//Count  string `json:"count"`
}

type GetBlockHeaderByHashArgs struct {
	Hash string `json:"hash"`
}

type GetBlocksByRangeArgs struct {
	From  string `json:"from"`
	Count string `json:"count"`
}

type GetBlocksArgs struct {
	Blocks string `json:"blocks"`
}

type ProgressBarArgs struct {
	Number int `json:"number"`
}

func (handler *ChainAPIHandler) GetBlockByNumber(args GetBlockByNumArgs, resp **BlockResp) error {
	var (
		number = int64(0)
		err    error
	)
	if number, err = strconv.ParseInt(args.Number, 10, 64); err != nil {
		return err
	}
	gotBlock := handler.BlockChain.GetBlockByNumber(uint64(number))
	return coverBlock2Resp(gotBlock, resp)
}

func (handler *ChainAPIHandler) Head(_ EmptyArgs, resp **BlockHeaderResp) error {
	gotBlock := handler.BlockChain.GetHead()
	return coverBlockHeader2Resp(gotBlock, resp)
}

func (handler *ChainAPIHandler) GetBlockHeaderByNumber(args GetBlockHeaderByNumberArgs, resp **BlockHeaderResp) error {
	var (
		number = int64(0)
		err    error
	)
	if number, err = strconv.ParseInt(args.Number, 10, 64); err != nil {
		return err
	}
	gotBlock := handler.BlockChain.GetBlockByNumber(uint64(number))
	return coverBlockHeader2Resp(gotBlock, resp)
}

func (handler *ChainAPIHandler) GetBlockHeaderByHash(args GetBlockHeaderByHashArgs, resp **BlockHeaderResp) error {
	if args.Hash == "" {
		return xfsgo.NewRPCError(-1006, "Parameter cannot be empty")
	}
	goBlock := handler.BlockChain.GetBlockByHash(common.Hex2Hash(args.Hash))
	return coverBlockHeader2Resp(goBlock, resp)
}

func (handler *ChainAPIHandler) GetBlockByHash(args GetBlockByHashArgs, resp **BlockResp) error {
	if args.Hash == "" {
		return xfsgo.NewRPCError(-1006, "Parameter cannot be empty")
	}
	gotBlock := handler.BlockChain.GetBlockByHash(common.Hex2Hash(args.Hash))
	return coverBlock2Resp(gotBlock, resp)

}

func (handler *ChainAPIHandler) GetTxsByBlockNum(args GetTxsByBlockNumArgs, resp *TransactionsResp) error {
	var (
		number = int64(0)
		err    error
	)
	if number, err = strconv.ParseInt(args.Number, 10, 64); err != nil {
		return err
	}
	blk := handler.BlockChain.GetBlockByNumber(uint64(number))
	if blk == nil {
		return xfsgo.NewRPCError(-1006, "Not found block")
	}
	txs := make([]*TransactionResp, 0)
	for _, item := range blk.Transactions {
		var txres = new(TransactionResp)
		if err := coverTx2Resp(item, &txres); err != nil {
			return err
		}
		txs = append(txs, txres)
	}
	*resp = txs
	return nil
}

func (handler *ChainAPIHandler) GetTxsByBlockHash(args GetTxbyBlockHashArgs, resp *TransactionsResp) error {
	if args.Hash == "" {
		return xfsgo.NewRPCError(-1006, "Parameter cannot be empty")
	}
	blk := handler.BlockChain.GetBlockByHash(common.Hex2Hash(args.Hash))
	if blk == nil {
		return xfsgo.NewRPCError(-1006, "Not found block")
	}
	txs := make([]*TransactionResp, 0)
	for _, item := range blk.Transactions {
		var txres = new(TransactionResp)
		if err := coverTx2Resp(item, &txres); err != nil {
			return err
		}
		txs = append(txs, txres)
	}
	*resp = txs
	return nil
}

func (handler *ChainAPIHandler) GetReceiptByHash(args GetReceiptByHashArgs, resp *xfsgo.Receipt) error {
	if args.Hash == "" {
		return xfsgo.NewRPCError(-1006, "Parameter cannot be empty")
	}
	data := handler.BlockChain.GetReceiptByHash(common.Hex2Hash(args.Hash))
	if data != nil {
		*resp = *data
	}
	return nil
}

func (handler *ChainAPIHandler) GetTransaction(args GetTransactionArgs, resp **TransactionResp) error {
	if args.Hash == "" {
		return xfsgo.NewRPCError(-1006, "Parameter cannot be empty")
	}
	ID := common.Hex2Hash(args.Hash)
	data := handler.BlockChain.GetTransaction(ID)
	return coverTx2Resp(data, resp)
}

func (handler *ChainAPIHandler) ExportBlocks(args GetBlocksByRangeArgs, resp *string) error {

	var numbersForm, numbersCount *big.Int
	var ok bool

	if args.From == "" {
		return xfsgo.NewRPCError(-1006, "Parameter cannot be empty")
	} else {
		numbersForm, ok = new(big.Int).SetString(args.From, 0)
		if !ok {
			return xfsgo.NewRPCError(-1006, "string to big.Int error")
		}
	}

	if args.Count == "" {
		blockHeight := handler.BlockChain.CurrentBlock().Height()
		numbersCount = new(big.Int).SetUint64(blockHeight)
	} else {
		numbersCount, ok = new(big.Int).SetString(args.Count, 0)
		if !ok {
			return xfsgo.NewRPCError(-1006, "string to big.Int error")
		}
	}

	if numbersCount.Uint64() == uint64(0) {
		b := handler.BlockChain.GetHead()
		numbersCount.SetUint64(b.Height())
	}

	if numbersForm.Uint64() >= numbersCount.Uint64() { // Export all
		b := handler.BlockChain.GetHead()
		numbersCount.SetUint64(b.Header.Height)
	}
	data := handler.BlockChain.GetBlocks(numbersForm.Uint64(), numbersCount.Uint64())
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

func (handler *ChainAPIHandler) ImportBlock(args GetBlocksArgs, resp *string) error {
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
	handler.number = len(blockChain) - 1

	for _, item := range blockChain {
		if err := handler.BlockChain.InsertChain(item); err != nil {
			continue
		}
	}
	*resp = "Import complete"
	return nil
}

func (handler *ChainAPIHandler) ProgressBar(_ EmptyArgs, resp *string) error {
	total := strconv.Itoa(handler.number)
	*resp = total
	return nil
}
