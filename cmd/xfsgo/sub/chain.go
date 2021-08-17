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
	"xfsgo"

	"github.com/spf13/cobra"
)

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
		Use:   "--hash <hash>",
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
	receipt := make([]map[string]interface{}, 1)
	req := &getBlockNumArgs{
		From:  json.Number(FormStr),
		Count: json.Number(CountStr),
	}
	err = cli.CallMethod(1, "Chain.GetBlockSection", &req, &receipt)
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
	jsonStr, err := json.Marshal(tran)
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println(jsonStr)
	return nil
}

func getBlockHash(cmd *cobra.Command, args []string) error {
	config, err := parseClientConfig(cfgFile)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	cli := xfsgo.NewClient(config.rpcClientApiHost)
	block := make(map[string]interface{}, 1)
	req := &getBlockHashArgs{
		Address: args[0],
	}
	err = cli.CallMethod(1, "Chain.GetBlockByHash", &req, &block)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	fmt.Println(block)
	return nil
}

func getHead() error {
	config, err := parseClientConfig(cfgFile)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	cli := xfsgo.NewClient(config.rpcClientApiHost)
	block := make(map[string]interface{}, 1)
	err = cli.CallMethod(1, "Chain.Head", nil, &block)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	// fmt.Println(block)
	fmt.Printf("transactions %v\n", block["transactions"])
	fmt.Printf("receipts %v\n", block["receipts"])
	blockheader := block["header"].(map[string]interface{})
	fmt.Printf("hash %v\n", blockheader["hash"])
	fmt.Printf("bits %v\n", blockheader["bits"])
	fmt.Printf("transactions_root %v\n", blockheader["transactions_root"])
	fmt.Printf("timestamp %v\n", blockheader["timestamp"])
	fmt.Printf("receipts_root %v\n", blockheader["receipts_root"])
	fmt.Printf("nonce %v\n", blockheader["nonce"])
	fmt.Printf("height %v\n", blockheader["height"])
	fmt.Printf("hash_prev_block %v\n", blockheader["hash_prev_block"])
	fmt.Printf("state_root %v\n", blockheader["state_root"])
	fmt.Printf("coinbase %v\n", blockheader["coinbase"])
	fmt.Printf("version %v\n", blockheader["version"])
	return nil
}

func init() {

	rootCmd.AddCommand(chainCommand)
	chainCommand.AddCommand(chainGetBlockCommond)
	chainCommand.AddCommand(chainGetHeadCommond)
	chainCommand.AddCommand(chainGetTransactionCommand)
	chainCommand.AddCommand(chainGetReceiptCommand)
	chainGetBlockCommond.AddCommand(chainGetBlockHashCommond)

}
