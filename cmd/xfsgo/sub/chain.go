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
	"strconv"
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
		Use:   "getblock",
		Short: "get block hash or number count",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}
			return nil
		},
	}
	chainGetBlockHashCommond = &cobra.Command{
		Use:   "hash <hash>",
		Short: "get block <hash>",
		RunE:  getBlockHash,
	}
	chainNumderCountCommand = &cobra.Command{
		Use:   "number [number] <count>",
		Short: "get block <number> <count>",
		RunE:  getBlockNum,
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
	if len(args) != 2 {
		return cmd.Help()
	}

	config, err := parseClientConfig(cfgFile)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	FormStr := args[0]
	CountStr := args[1]

	fromUint, err := strconv.Atoi(FormStr)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	CountUint, err := strconv.Atoi(CountStr)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	cli := xfsgo.NewClient(config.rpcClientApiHost)
	receipt := make(map[string]interface{}, 1)
	req := &getBlockNumArgs{
		From:  uint64(fromUint),
		Count: uint64(CountUint),
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
	fmt.Print("key                                value")
	fmt.Println()
	for k, v := range block {
		if v != nil {
			for ks, vs := range block[k].(map[string]interface{}) {
				fmt.Printf("%v -> %-25v", k, ks)
				fmt.Printf("%-50v", vs)
				fmt.Println()
			}
		} else {
			fmt.Printf("%-35v", k)
			fmt.Printf("%-50v", v)
			fmt.Println()
		}
	}
	return nil
}

func getHead() error {
	config, err := parseClientConfig(cfgFile)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	cli := xfsgo.NewClient(config.rpcClientApiHost)
	blockHeader := make(map[string]interface{}, 1)
	err = cli.CallMethod(1, "Chain.Head", nil, &blockHeader)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	fmt.Print("key                                value")
	fmt.Println()
	for k, v := range blockHeader {
		if v != nil {
			for ks, vs := range blockHeader[k].(map[string]interface{}) {
				fmt.Printf("%v -> %-25v", k, ks)
				fmt.Printf("%-50v", vs)
				fmt.Println()
			}
		} else {
			fmt.Printf("%-35v", k)
			fmt.Printf("%-50v", v)
			fmt.Println()
		}
	}
	return nil
}

func init() {

	rootCmd.AddCommand(chainCommand)
	chainCommand.AddCommand(chainGetBlockCommond)
	chainCommand.AddCommand(chainGetHeadCommond)
	chainCommand.AddCommand(chainGetTransactionCommand)
	chainCommand.AddCommand(chainGetReceiptCommand)
	chainGetBlockCommond.AddCommand(chainGetBlockHashCommond)
	chainGetBlockCommond.AddCommand(chainNumderCountCommand)

}
