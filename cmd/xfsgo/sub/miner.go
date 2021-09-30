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
	"math/big"
	"xfsgo"
	"xfsgo/common"

	"github.com/spf13/cobra"
)

var (
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
		Short: "start miner",
		RunE:  runMinerStart,
	}
	minerStopCommand = &cobra.Command{
		Use:   "stop",
		Short: "stop miner",
		RunE:  runMinerStop,
	}
	minerWorkersAddCommand = &cobra.Command{
		Use:   "workeradd",
		Short: "Miner supplemental thread",
		RunE:  WorkersAdd,
	}
	minerWorkersDownCommand = &cobra.Command{
		Use:   "workerdown",
		Short: "Miners reduce threads",
		RunE:  WorkersDown,
	}
	minerSetGasPriceCommand = &cobra.Command{
		Use:   "setgasprice <price>",
		Short: "Miner set gas price",
		RunE:  MinSetGasPrice,
	}
	minerGetGasInfoCommand = &cobra.Command{
		Use:   "getinfo",
		Short: "Miner get gas info",
		RunE:  GetGasInfo,
	}
)

func runMinerStart(_ *cobra.Command, _ []string) error {
	config, err := parseClientConfig(cfgFile)
	if err != nil {
		return err
	}
	var res *string = nil
	cli := xfsgo.NewClient(config.rpcClientApiHost)
	if err = cli.CallMethod(1, "Miner.Start", nil, &res); err != nil {
		return nil
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

func WorkersAdd(_ *cobra.Command, _ []string) error {
	config, err := parseClientConfig(cfgFile)
	if err != nil {
		return err
	}
	var res *string = nil
	cli := xfsgo.NewClient(config.rpcClientApiHost)
	err = cli.CallMethod(1, "Miner.WorkersAdd", nil, &res)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	return nil
}

func WorkersDown(_ *cobra.Command, _ []string) error {
	config, err := parseClientConfig(cfgFile)
	if err != nil {
		return err
	}
	var res *string = nil
	cli := xfsgo.NewClient(config.rpcClientApiHost)
	err = cli.CallMethod(1, "Miner.WorkersDown", nil, &res)
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
	gasPrice, _ := new(big.Int).SetString(args[0], 0)
	r := common.NanoCoin2BaseCoin(gasPrice)
	req := &MinSetGasPriceArgs{
		GasPrice: r.String(),
	}
	err = cli.CallMethod(1, "Miner.MinSetGasPrice", &req, &res)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	return nil
}

func GetGasInfo(_ *cobra.Command, _ []string) error {
	config, err := parseClientConfig(cfgFile)
	if err != nil {
		return err
	}
	res := make(map[string]interface{}, 1)
	cli := xfsgo.NewClient(config.rpcClientApiHost)
	err = cli.CallMethod(1, "Miner.GetGasInfo", nil, &res)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	bs, err := common.MarshalIndent(res)
	if err != nil {
		return err
	}
	fmt.Println(string(bs))

	return nil
}
func init() {
	minerCommand.AddCommand(minerStartCommand)
	minerCommand.AddCommand(minerStopCommand)
	minerCommand.AddCommand(minerWorkersAddCommand)
	minerCommand.AddCommand(minerWorkersDownCommand)
	minerCommand.AddCommand(minerSetGasPriceCommand)
	minerCommand.AddCommand(minerGetGasInfoCommand)
	rootCmd.AddCommand(minerCommand)
}
