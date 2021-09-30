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
	getTxpoolCommand = &cobra.Command{
		Use:   "txpool",
		Short: "transaction pool info",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	getTxpoolListCommand = &cobra.Command{
		Use:   "list",
		Short: "transaction pool transaction list",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTxPoolList()
		},
	}
	getTxpoolCountCommand = &cobra.Command{
		Use:   "count",
		Short: "transaction pool transaction number",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTxPoolCount()
		},
	}
)

func runTxPoolList() error {
	config, err := parseClientConfig(cfgFile)
	if err != nil {
		fmt.Println(err)
		return err
	}
	cli := xfsgo.NewClient(config.rpcClientApiHost)
	tran := make([]xfsgo.Transaction, 1)
	err = cli.CallMethod(1, "TxPool.GetPending", nil, &tran)
	if err != nil {
		fmt.Println(err)
		return err
	}
	str, err := json.Marshal(tran)
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println(string(str))
	return nil
}

func runTxPoolCount() error {
	config, err := parseClientConfig(cfgFile)
	if err != nil {
		fmt.Println(err)
		return err
	}
	cli := xfsgo.NewClient(config.rpcClientApiHost)
	var txPoolCount int
	err = cli.CallMethod(1, "TxPool.GetPendingSize", nil, &txPoolCount)
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println(txPoolCount)
	return nil
}

func init() {
	rootCmd.AddCommand(getTxpoolCommand)
	getTxpoolCommand.AddCommand(getTxpoolListCommand)
	getTxpoolCommand.AddCommand(getTxpoolCountCommand)
}
