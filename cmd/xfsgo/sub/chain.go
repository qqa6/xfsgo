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

package sub

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"xfsgo"
	"xfsgo/common"

	"github.com/spf13/cobra"
)

// var fileType string = ".car"
var (
	chainCommand = &cobra.Command{
		Use:   "chain",
		Short: "show chain info",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}
			return nil
		},
	}
	chainGetHeadCommond = &cobra.Command{
		Use:   "gethead",
		Short: "get the current height block",
		RunE: func(cmd *cobra.Command, args []string) error {
			return getHead()
		},
	}
	chainGetBlockCommond = &cobra.Command{
		Use:   "getblock <form> <count>",
		Short: "get block hash or number count",
		RunE:  getBlockNum,
	}
	chainGetBlockHashCommond = &cobra.Command{
		Use:   "hash <hash>",
		Short: "get block <hash>",
		RunE:  getBlockHash,
	}
	chainGetTransactionCommand = &cobra.Command{
		Use:   "gettransaction <hash>",
		Short: "get transaction <hash>",
		RunE:  getTransaction,
	}
	chainGetReceiptCommand = &cobra.Command{
		Use:   "getreceipt <hash>",
		Short: "get receipt <hash>",
		RunE:  getReceipt,
	}
	chainExportCommand = &cobra.Command{
		Use:   "export <address> <form> <count>",
		Short: "import chain state from a given chain export file or <address> <form> <count>",
		RunE:  xfsLotusExport,
	}
	chainImportCommand = &cobra.Command{
		Use:   "import <address>",
		Short: "import chain state from a given chain export file or <address>",
		RunE:  xfsLotusImport,
	}
	chainprogressCommand = &cobra.Command{
		Use:   "getprogress",
		Short: "Block pouring progress bar",
		RunE:  ExportProgressBar,
	}
)

func getBlockNum(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return cmd.Help()
	}

	config, err := parseClientConfig(cfgFile)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	var FormStr string
	var CountStr string
	if len(args) == 1 {
		FormStr = args[0]
	} else {
		FormStr = args[0]
		CountStr = args[1]
	}

	cli := xfsgo.NewClient(config.rpcClientApiHost)
	var receipt []common.BlockMap
	req := &getBlockNumArgs{
		From:  json.Number(FormStr),
		Count: json.Number(CountStr),
	}
	err = cli.CallMethod(1, "Chain.GetBlockSection", &req, &receipt)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	var jsons []common.BlockMap
	// common.Marshals
	for _, item := range receipt {
		jsons = append(jsons, item.MapMerge())
	}
	sortKey := []string{"version", "height", "hash_prev_block", "hash", "timestamp", "state_root", "transactions_root", "receipts_root", "bits", "nonce", "coinbase", "transactions", "receipts"}

	bs, err := common.Marshals(jsons, sortKey, true)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	fmt.Println(string(bs))
	return nil
}

func getReceipt(cmd *cobra.Command, args []string) error {
	config, err := parseClientConfig(cfgFile)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	cli := xfsgo.NewClient(config.rpcClientApiHost)
	receipt := make(map[string]interface{}, 1)
	req := &getReceiptArgs{
		Hash: args[0],
	}
	err = cli.CallMethod(1, "Chain.GetReceiptByHash", &req, &receipt)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	jsonStr, err := json.Marshal(receipt)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	fmt.Println(jsonStr)
	return nil
}

func getTransaction(cmd *cobra.Command, args []string) error {
	config, err := parseClientConfig(cfgFile)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	cli := xfsgo.NewClient(config.rpcClientApiHost)
	tran := make(map[string]interface{}, 1)
	req := &getTransactionArgs{
		Hash: args[0],
	}

	err = cli.CallMethod(1, "Chain.GetTransaction", &req, &tran)
	if err != nil {
		fmt.Println(err)
		return err
	}
	bs, err := json.MarshalIndent(&tran, "", "    ")
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println(string(bs))
	return nil
}

func getBlockHash(cmd *cobra.Command, args []string) error {
	config, err := parseClientConfig(cfgFile)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	cli := xfsgo.NewClient(config.rpcClientApiHost)
	var block common.BlockMap
	req := &getBlockHashArgs{
		Hash: args[0],
	}
	err = cli.CallMethod(1, "Chain.GetBlockByHash", &req, &block)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	sortKey := []string{"version", "height", "hash_prev_block", "hash", "timestamp", "state_root", "transactions_root", "receipts_root", "bits", "nonce", "coinbase", "transactions", "receipts"}
	result := block.MapMerge()
	bs, err := common.Marshal(result, sortKey, true)
	if err != nil {
		fmt.Println(err)
		return nil

	}
	fmt.Println(bs)

	return nil
}

func getHead() error {
	config, err := parseClientConfig(cfgFile)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	cli := xfsgo.NewClient(config.rpcClientApiHost)
	var block common.BlockMap
	err = cli.CallMethod(1, "Chain.Head", nil, &block)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	result := block.MapMerge()
	sortKey := []string{"version", "height", "hash_prev_block", "hash", "timestamp", "state_root", "transactions_root", "receipts_root", "bits", "nonce", "coinbase", "transactions", "receipts"}
	bs, err := common.Marshal(result, sortKey, true)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	fmt.Printf("%v\n", bs)
	return nil
}

func xfsLotusImport(cmd *cobra.Command, args []string) error {
	if len(args) < int(1) {
		cmd.Help()
		return nil
	}
	address := args[0]
	data, err := ioutil.ReadFile(address)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	config, err := parseClientConfig(cfgFile)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	req := &GetBlocksArgs{
		Blocks: string(data),
	}
	cli := xfsgo.NewClient(config.rpcClientApiHost)
	var isInsertBlock string
	err = cli.CallMethod(1, "Chain.ImportBlock", req, &isInsertBlock)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}

	fmt.Println(isInsertBlock)
	return nil
}

func xfsLotusExport(cmd *cobra.Command, args []string) error {
	if len(args) < int(1) {
		cmd.Help()
		return nil
	}

	var FormStr string
	var CountStr string
	exportPath := args[0]
	if len(args) == 2 {
		FormStr = args[1]
	}
	if len(args) > 2 {
		FormStr = args[1]
		CountStr = args[2]
	}

	config, err := parseClientConfig(cfgFile)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	req := &getBlockNumArgs{
		From:  json.Number(FormStr),
		Count: json.Number(CountStr),
	}
	cli := xfsgo.NewClient(config.rpcClientApiHost)
	var blockEncodeData string
	err = cli.CallMethod(1, "Chain.ExportBlock", req, &blockEncodeData)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	fp, err := os.Create(exportPath)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	fp.Write([]byte(blockEncodeData))
	defer fp.Close()
	return nil
}

func ExportProgressBar(cmd *cobra.Command, args []string) error {
	config, err := parseClientConfig(cfgFile)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	cli := xfsgo.NewClient(config.rpcClientApiHost)
	var resp string
	err = cli.CallMethod(1, "Chain.ProgressBar", nil, &resp)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	fmt.Println(resp)
	return nil
}

func init() {

	rootCmd.AddCommand(chainCommand)
	chainCommand.AddCommand(chainGetBlockCommond)
	chainCommand.AddCommand(chainGetHeadCommond)
	chainCommand.AddCommand(chainGetTransactionCommand)
	chainCommand.AddCommand(chainGetReceiptCommand)
	chainCommand.AddCommand(chainExportCommand)
	chainCommand.AddCommand(chainImportCommand)
	chainCommand.AddCommand(chainprogressCommand)
	chainCommand.AddCommand(chainGetBlockHashCommond)

}
