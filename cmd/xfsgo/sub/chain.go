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
	"fmt"
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
		Short: "get the latest block information of local chain data",
		RunE: func(cmd *cobra.Command, args []string) error {
			return getHead()
		},
	}
	chainGetBlockByNumCommond = &cobra.Command{
		Use:   "getblockbynum <number>",
		Short: "query block information of specified height",
		RunE:  getBlockByNum,
	}
	chainGetBlockByHashCommond = &cobra.Command{
		Use:   "getblockbyhash <hash>",
		Short: "query the block information of the specified hash value",
		RunE:  getBlockByHash,
	}
	chainGetTxsByBlockNumCommond = &cobra.Command{
		Use:   "gettxsbyblocknum <number>",
		Short: "query transaction information of specified block height",
		RunE:  getTxsByBlockNum,
	}
	chainGetTxsByBlockHashCommond = &cobra.Command{
		Use:   "gettxsbyblockhash <hash>",
		Short: "query the transaction information of the specified block hash value",
		RunE:  getTxsByBlockHash,
	}
	chainGetTransactionCommand = &cobra.Command{
		Use:   "gettxbyhash <transaction_hash>",
		Short: "query the transaction information of the specified transaction hash value",
		RunE:  getTransaction,
	}
	chainGetReceiptByHashCommand = &cobra.Command{
		Use:   "getreceiptbytxhash <transaction_hash>",
		Short: "query receipt information of specified transaction hash value",
		RunE:  getReceiptByTxHash,
	}
	chainExportCommand = &cobra.Command{
		Use:   "export <filename> <number> [count]",
		Short: "export local chain data snapshot",
		RunE:  exportBlocks,
	}
	// chainImportCommand = &cobra.Command{
	// 	Use:   "import <filename>",
	// 	Short: "import local chain data snapshot",
	// 	RunE:  xfsLotusImport,
	// }
	chainprogressCommand = &cobra.Command{
		Use:   "getprogress",
		Short: "query block import progress",
		RunE:  exportProgressBar,
	}
)

// Gets the header of the highest block
func getHead() error {
	config, err := parseClientConfig(cfgFile)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	cli := xfsgo.NewClient(config.rpcClientApiHost)
	var block common.BlocksMap
	err = cli.CallMethod(1, "Chain.Head", nil, &block)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	result := block.MapMerge()
	sortIndex := []string{"version", "height", "hash_prev_block", "hash", "timestamp", "state_root", "transactions_root", "receipts_root", "bits", "nonce", "coinbase", "gas_limit", "gas_used"}
	bs, err := common.Marshal(result, sortIndex, true)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	fmt.Printf("%v\n", bs)
	return nil
}

// Block information according to block height
func getBlockByNum(cmd *cobra.Command, args []string) error {

	// Required parameters
	if len(args) < 1 {
		return cmd.Help()
	}

	// Service configuration
	config, err := parseClientConfig(cfgFile)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	cli := xfsgo.NewClient(config.rpcClientApiHost)
	var receipt []common.BlocksMap
	req := &getBlockByNumArgs{
		Number: args[0],
	}
	err = cli.CallMethod(1, "Chain.GetBlockByNumber", &req, &receipt)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	var jsons []common.BlocksMap
	for _, item := range receipt {
		jsons = append(jsons, item.MapMerge())
	}
	sortIndex := []string{"version", "height", "hash_prev_block", "hash", "timestamp", "state_root", "transactions_root", "receipts_root", "bits", "nonce", "coinbase", "gas_limit", "gas_used"}
	bs, err := common.Marshals(jsons, sortIndex, true)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	fmt.Println(string(bs))
	return nil
}

// Get blocks from hash
func getBlockByHash(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return cmd.Help()
	}

	config, err := parseClientConfig(cfgFile)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	cli := xfsgo.NewClient(config.rpcClientApiHost)
	var block common.BlocksMap
	req := &getBlockByHashArgs{
		Hash: args[0],
	}
	err = cli.CallMethod(1, "Chain.GetBlockByHash", &req, &block)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	sortIndex := []string{"version", "height", "hash_prev_block", "hash", "timestamp", "state_root", "transactions_root", "receipts_root", "bits", "nonce", "coinbase", "gas_limit", "gas_used"}
	result := block.MapMerge()
	bs, err := common.Marshal(result, sortIndex, true)
	if err != nil {
		fmt.Println(err)
		return nil

	}
	fmt.Println(bs)

	return nil
}

// Get all transactions according to block height
func getTxsByBlockNum(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return cmd.Help()
	}

	config, err := parseClientConfig(cfgFile)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	cli := xfsgo.NewClient(config.rpcClientApiHost)
	result := make([]map[string]interface{}, 1)
	req := &getTxsByBlockNumArgs{
		Number: args[0],
	}
	cli.CallMethod(1, "Chain.GetTxsByBlockNum", &req, &result)
	bs, err := common.MarshalIndent(result)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	fmt.Println(string(bs))
	return nil
}

// Get all transactions based on block hash
func getTxsByBlockHash(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return cmd.Help()
	}
	config, err := parseClientConfig(cfgFile)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	result := make([]map[string]interface{}, 1)
	cli := xfsgo.NewClient(config.rpcClientApiHost)
	req := getTxsByBlockHashArgs{
		Hash: args[0],
	}
	cli.CallMethod(1, "Chain.GetTxsByBlockHash", &req, &result)
	bs, err := common.MarshalIndent(result)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	fmt.Println(string(bs))
	return nil
}

// Get receipt information according to hash
func getReceiptByTxHash(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return cmd.Help()
	}
	config, err := parseClientConfig(cfgFile)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	cli := xfsgo.NewClient(config.rpcClientApiHost)
	result := make(map[string]interface{}, 1)
	req := &getReceiptByHashArgs{
		Hash: args[0],
	}
	err = cli.CallMethod(1, "Chain.GetReceiptByHash", &req, &result)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	bs, err := common.MarshalIndent(result)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	fmt.Println(string(bs))
	return nil
}

// Obtain transaction information according to transaction hash
func getTransaction(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return cmd.Help()
	}
	config, err := parseClientConfig(cfgFile)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	cli := xfsgo.NewClient(config.rpcClientApiHost)
	result := make(map[string]interface{}, 1)
	req := &getTransactionArgs{
		Hash: args[0],
	}

	err = cli.CallMethod(1, "Chain.GetTransaction", &req, &result)
	if err != nil {
		fmt.Println(err)
		return err
	}
	bs, err := common.MarshalIndent(result)
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println(string(bs))
	return nil
}

// func xfsLotusImport(cmd *cobra.Command, args []string) error {
// 	if len(args) < int(1) {
// 		cmd.Help()
// 		return nil
// 	}
// 	address := args[0]
// 	data, err := ioutil.ReadFile(address)
// 	if err != nil {
// 		fmt.Println(err.Error())
// 		return nil
// 	}
// 	config, err := parseClientConfig(cfgFile)
// 	if err != nil {
// 		fmt.Println(err)
// 		return nil
// 	}
// 	req := &GetBlocksArgs{
// 		Blocks: string(data),
// 	}
// 	cli := xfsgo.NewClient(config.rpcClientApiHost)
// 	var result string
// 	err = cli.CallMethod(1, "Chain.ImportBlock", req, &result)
// 	if err != nil {
// 		fmt.Println(err.Error())
// 		return nil
// 	}
// 	fmt.Println(result)
// 	return nil
// }

func exportBlocks(cmd *cobra.Command, args []string) error {
	if len(args) < int(1) {
		cmd.Help()
		return nil
	}

	var Form, Count string
	exportPath := args[0]
	if len(args) == 2 {
		Form = args[1]
	}
	if len(args) > 2 {
		Form = args[1]
		Count = args[2]
	}

	config, err := parseClientConfig(cfgFile)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	req := &getBlocksByRangeArgs{
		From:  Form,
		Count: Count,
	}
	cli := xfsgo.NewClient(config.rpcClientApiHost)
	var result string
	err = cli.CallMethod(1, "Chain.ExportBlocks", req, &result)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	fp, err := os.Create(exportPath)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	fp.Write([]byte(result))
	defer fp.Close()
	return nil
}

func exportProgressBar(cmd *cobra.Command, args []string) error {
	config, err := parseClientConfig(cfgFile)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	cli := xfsgo.NewClient(config.rpcClientApiHost)
	var result string
	err = cli.CallMethod(1, "Chain.ProgressBar", nil, &result)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	fmt.Println(result)
	return nil
}

func init() {

	rootCmd.AddCommand(chainCommand)
	chainCommand.AddCommand(chainGetBlockByNumCommond)
	chainCommand.AddCommand(chainGetHeadCommond)
	chainCommand.AddCommand(chainGetTransactionCommand)
	chainCommand.AddCommand(chainGetReceiptByHashCommand)
	chainCommand.AddCommand(chainExportCommand)
	// chainCommand.AddCommand(chainImportCommand)
	chainCommand.AddCommand(chainprogressCommand)
	chainCommand.AddCommand(chainGetBlockByHashCommond)
	chainCommand.AddCommand(chainGetTxsByBlockHashCommond)
	chainCommand.AddCommand(chainGetTxsByBlockNumCommond)

}
