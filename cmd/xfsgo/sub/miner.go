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
	"strconv"
	"xfsgo"

	"github.com/spf13/cobra"
)

var (
	workers  int
	coinbase string
	gasprice string
	gaslimit string

	minerCommand = &cobra.Command{
		Use:   "miner",
		Short: "miner serve info",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}
			return nil
		},
	}
	minerStartCommand = &cobra.Command{
		Use:   "start",
		Short: "Start mining service",
		RunE:  runMinerStart,
	}
	minerStopCommand = &cobra.Command{
		Use:   "stop",
		Short: "Stop mining services",
		RunE:  runMinerStop,
	}
	minerWorkersAddCommand = &cobra.Command{
		Use:   "addworker [count]",
		Short: "Add number of workers",
		RunE:  WorkersAdd,
	}
	minerWorkersDownCommand = &cobra.Command{
		Use:   "delworker [count]",
		Short: "Miners reduce threads",
		RunE:  WorkersDown,
	}
	minerSetGasPriceCommand = &cobra.Command{
		Use:   "setgasprice <price>",
		Short: "Miner set gas price",
		RunE:  MinSetGasPrice,
	}
	minerSetGasLimitCommand = &cobra.Command{
		Use:   "setgaslimit <limit>",
		Short: "Miner set gas limit",
		RunE:  MinSetGasLimit,
	}
	minerGetStatusCommand = &cobra.Command{
		Use:   "status",
		Short: "Get current miner status",
		RunE:  MinGetStatus,
	}
)

func runMinerStart(_ *cobra.Command, args []string) error {

	config, err := parseClientConfig(cfgFile)
	if err != nil {
		return err
	}
	var res *string = nil
	cli := xfsgo.NewClient(config.rpcClientApiHost)
	if err = cli.CallMethod(1, "Miner.Start", nil, &res); err != nil {
		return nil
	}

	// worker
	if workers > 0 {
		num, err := strconv.Atoi(args[0])
		if err != nil {
			return err
		}
		req := &MinWorkerArgs{
			WorkerNum: num,
		}

		err = cli.CallMethod(1, "Miner.WorkersAdd", &req, &res)
		if err != nil {
			fmt.Println(err.Error())
			return nil
		}
	}
	// coinbase
	if coinbase != "" {
		req := &MinSetCoinbaseArgs{
			Coinbase: args[1],
		}
		err = cli.CallMethod(1, "Miner.WorkersAdd", &req, &res)
		if err != nil {
			fmt.Println(err.Error())
			return nil
		}
	}
	// gasprice
	if gasprice != "" {
		req := &MinSetGasPriceArgs{
			GasPrice: args[2],
		}
		err = cli.CallMethod(1, "Miner.MinSetGasPrice", &req, &res)
		if err != nil {
			fmt.Println(err.Error())
			return nil
		}
	}
	// gaslimit
	if gaslimit != "" {
		req := &MinSetGasLimitArgs{
			GasLimit: args[3],
		}
		err = cli.CallMethod(1, "Miner.MinSetGasLimit", &req, &res)
		if err != nil {
			fmt.Println(err.Error())
			return nil
		}
	}
	fmt.Println("miner running...")
	return nil
}

func runMinerStop(_ *cobra.Command, _ []string) error {
	config, err := parseClientConfig(cfgFile)
	if err != nil {
		return err
	}
	var res *string = nil
	cli := xfsgo.NewClient(config.rpcClientApiHost)
	err = cli.CallMethod(1, "Miner.Stop", nil, &res)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	fmt.Println("miner stop...")
	return nil
}

func WorkersAdd(cmd *cobra.Command, args []string) error {

	config, err := parseClientConfig(cfgFile)
	if err != nil {
		return err
	}
	var res *string = nil
	cli := xfsgo.NewClient(config.rpcClientApiHost)

	req := &MinWorkerArgs{
		WorkerNum: 1,
	}

	var num int
	if len(args) > 0 {
		num, err = strconv.Atoi(args[0])
		if err != nil {
			return err
		}
	}
	req.WorkerNum = num

	err = cli.CallMethod(1, "Miner.WorkersAdd", &req, &res)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	fmt.Println(res)
	return nil
}

func WorkersDown(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		cmd.Help()
		return nil
	}

	config, err := parseClientConfig(cfgFile)
	if err != nil {
		return err
	}
	var res *string = nil
	cli := xfsgo.NewClient(config.rpcClientApiHost)
	num, err := strconv.Atoi(args[0])
	if err != nil {
		return err
	}
	req := &MinWorkerArgs{
		WorkerNum: num,
	}
	err = cli.CallMethod(1, "Miner.WorkersDown", &req, &res)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	return nil
}

func MinSetGasPrice(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		cmd.Help()
		return nil
	}

	config, err := parseClientConfig(cfgFile)
	if err != nil {
		return err
	}
	var res *string = nil
	cli := xfsgo.NewClient(config.rpcClientApiHost)
	req := &MinSetGasPriceArgs{
		GasPrice: args[2],
	}
	err = cli.CallMethod(1, "Miner.MinSetGasPrice", &req, &res)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	return nil
}

func MinSetGasLimit(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		cmd.Help()
		return nil
	}

	config, err := parseClientConfig(cfgFile)
	if err != nil {
		return err
	}
	var res *string = nil
	cli := xfsgo.NewClient(config.rpcClientApiHost)
	req := &MinSetGasLimitArgs{
		GasLimit: args[0],
	}
	err = cli.CallMethod(1, "Miner.MinSetGasLimit", &req, &res)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	return nil
}

func MinGetStatus(_ *cobra.Command, _ []string) error {
	config, err := parseClientConfig(cfgFile)
	if err != nil {
		return err
	}
	res := make(map[string]interface{}, 1)
	cli := xfsgo.NewClient(config.rpcClientApiHost)
	err = cli.CallMethod(1, "Miner.MinGetStatus", nil, &res)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}

	var statusStr string
	if res["status"].(bool) {
		statusStr = "Running"
	} else {
		statusStr = "Stop"
	}
	fmt.Printf("Status: %v\n", statusStr)
	fmt.Printf("LastStarTime: %v\n", res["last_start_time"])
	fmt.Printf("Workers: %v\n", res["workers"])
	fmt.Printf("Coinbase: %v\n", res["coinbase"])

	fmt.Printf("GasPrice: %v (Atto)\n", res["gas_price"].(string))
	fmt.Printf("GasLimit: %v (Atto)\n", res["gas_limit"].(string))
	fmt.Printf("HashRate: %v\n", res["hash_rate"])
	return nil
}
func init() {
	minerCommand.AddCommand(minerStartCommand)
	mFlags := minerStartCommand.PersistentFlags()
	mFlags.IntVarP(&workers, "workers", "", 0, "Set number of workers")
	mFlags.StringVarP(&coinbase, "coinbase", "", "", "Set miner income address")
	mFlags.StringVarP(&gasprice, "gasprice", "", "", "Set expected gas unit price")
	mFlags.StringVarP(&gaslimit, "gaslimit", "", "", "Set maximum gas purchase quantity")
	minerCommand.AddCommand(minerStopCommand)
	minerCommand.AddCommand(minerWorkersAddCommand)
	minerCommand.AddCommand(minerWorkersDownCommand)
	minerCommand.AddCommand(minerSetGasPriceCommand)
	minerCommand.AddCommand(minerSetGasLimitCommand)
	minerCommand.AddCommand(minerGetStatusCommand)
	rootCmd.AddCommand(minerCommand)
}
