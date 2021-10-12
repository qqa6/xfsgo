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
	"xfsgo"
	"xfsgo/common"

	"github.com/spf13/cobra"
)

var (
	getTxpoolCommand = &cobra.Command{
		Use:   "txpool",
		Short: "transaction pool info",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	getGetPendingCommand = &cobra.Command{
		Use:   "pending",
		Short: "get transaction pool pending queue",
		RunE: func(cmd *cobra.Command, args []string) error {
			return GetPending()
		},
	}
	getGetTranCommand = &cobra.Command{
		Use:   "gettx <transaction_hash>",
		Short: "get transaction pool pending by transaction hash",
		RunE:  GetTransaction,
	}
	getTxpoolCountCommand = &cobra.Command{
		Use:   "count",
		Short: "transaction pool transaction number",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTxPoolCount()
		},
	}
	modifyTransCommand = &cobra.Command{
		Use:   "replace <gas_limit> <gas_price> <transaction_hash>",
		Short: "Replace the information of the specified transaction in the transaction pool",
		RunE:  ModifyTranGas,
	}
)

func GetPending() error {
	config, err := parseClientConfig(cfgFile)
	if err != nil {
		fmt.Println(err)
		return err
	}
	txPending := make([]*xfsgo.Transaction, 1)
	cli := xfsgo.NewClient(config.rpcClientApiHost, config.rpcClientApiTimeOut)
	err = cli.CallMethod(1, "TxPool.GetPending", nil, &txPending)
	if err != nil {
		return err
	}
	bs, err := common.MarshalIndent(txPending)
	if err != nil {
		return err
	}
	fmt.Println(string(bs))
	return nil
}

func runTxPoolCount() error {
	config, err := parseClientConfig(cfgFile)
	if err != nil {
		fmt.Println(err)
		return err
	}
	cli := xfsgo.NewClient(config.rpcClientApiHost, config.rpcClientApiTimeOut)
	var txPoolCount int
	err = cli.CallMethod(1, "TxPool.GetPendingSize", nil, &txPoolCount)
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println(txPoolCount)
	return nil
}

func GetTransaction(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return cmd.Help()
	}
	config, err := parseClientConfig(cfgFile)
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Printf(args[0])
	res := make(map[string]interface{}, 1)
	hash := &getTranByHashArgs{
		Hash: args[0],
	}

	cli := xfsgo.NewClient(config.rpcClientApiHost, config.rpcClientApiTimeOut)
	err = cli.CallMethod(1, "TxPool.GetTranByHash", &hash, &res)
	if err != nil {
		return err
	}
	if res == nil {
		return nil
	}
	bs, err := common.MarshalIndent(res)
	if err != nil {
		return err
	}
	fmt.Println(string(bs))

	return nil
}

func ModifyTranGas(cmd *cobra.Command, args []string) error {
	if len(args) < 3 {
		return cmd.Help()
	}
	config, err := parseClientConfig(cfgFile)
	if err != nil {
		fmt.Println(err)
		return err
	}
	var res string
	cli := xfsgo.NewClient(config.rpcClientApiHost, config.rpcClientApiTimeOut)
	req := &TranGasArgs{
		GasLimit: args[0],
		GasPrice: args[1],
		Hash:     args[2],
	}
	err = cli.CallMethod(1, "TxPool.ModifyTranGas", &req, &res)
	if err != nil {
		return err
	}

	return nil
}

func init() {
	rootCmd.AddCommand(getTxpoolCommand)
	getTxpoolCommand.AddCommand(getTxpoolCountCommand)
	getTxpoolCommand.AddCommand(getGetPendingCommand)
	getTxpoolCommand.AddCommand(modifyTransCommand)
	getTxpoolCommand.AddCommand(getGetTranCommand)
}
