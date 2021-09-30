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
	"encoding/hex"
	"fmt"
	"math/big"
	"xfsgo"
	"xfsgo/common"

	"github.com/spf13/cobra"
)

var (
	fromAddr      string
	gasLimit      string
	gasPrice      string
	walletCommand = &cobra.Command{
		Use:   "wallet",
		Short: "get wallet info",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	walletListCommand = &cobra.Command{
		Use:   "list",
		Short: "get wallet address list",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWalletList()
		},
	}
	walletNewCommand = &cobra.Command{
		Use:   "new",
		Short: "Create wallet address",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWalletNew()
		},
	}
	walletDelCommand = &cobra.Command{
		Use:   "del <address>",
		Short: "Delete wallet <address>",
		RunE:  runWalletDel,
	}
	walletSetAddrDefCommand = &cobra.Command{
		Use:   "setdef <address>",
		Short: "set default wallet <address>",
		RunE:  setWalletAddrDef,
	}
	walletGetAddrDefCommand = &cobra.Command{
		Use:   "getdef",
		Short: "get default wallet address",
		RunE: func(cmd *cobra.Command, args []string) error {
			return getWalletAddrDef()
		},
	}
	walletExportCommand = &cobra.Command{
		Use:   "export <address>",
		Short: "export wallet <address>",
		RunE:  runWalletExport,
	}
	walletImportCommand = &cobra.Command{
		Use:   "import <key>",
		Short: "import wallet <key>",
		RunE:  runWalletImport,
	}
	walletTransferCommand = &cobra.Command{
		Use:   "transfer <address> <value>",
		Short: "Send the transaction to the specified destination address",
		RunE:  runWalletTransfer,
	}
	walletSetGasCommand = &cobra.Command{
		Use:   "setgasprice <price>",
		Short: "set gas price",
		RunE:  SetGasPrice,
	}
	walletSetGasLimitCommand = &cobra.Command{
		Use:   "setgaslimit <limit>",
		Short: "set gas limit",
		RunE:  SetGas,
	}
)

func runWalletTransfer(cmd *cobra.Command, args []string) error {
	if len(args) < 2 {
		return cmd.Help()
	}
	config, err := parseClientConfig(cfgFile)
	if err != nil {
		return err
	}

	cli := xfsgo.NewClient(config.rpcClientApiHost)
	result := make(map[string]interface{}, 1)
	req := &transferFromArgs{
		To:    args[0],
		Value: args[1],
	}
	if fromAddr != "" {
		req.From = fromAddr
	}
	if gasLimit != "" {
		req.GasLimit = gasLimit
	}
	if gasPrice != "" {
		req.GasPrice = gasPrice
	}

	err = cli.CallMethod(1, "Wallet.TransferFrom", &req, &result)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	bs, err := common.MarshalIndent(result)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	fmt.Printf("%v\n", string(bs))

	return nil
}

func runWalletNew() error {
	config, err := parseClientConfig(cfgFile)
	if err != nil {
		return err
	}
	cli := xfsgo.NewClient(config.rpcClientApiHost)
	var addr *string = nil
	err = cli.CallMethod(1, "Wallet.Create", nil, &addr)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	fmt.Println(*addr)
	return nil
}

func runWalletDel(cmd *cobra.Command, args []string) error {
	config, err := parseClientConfig(cfgFile)
	if err != nil {
		return err
	}
	addr := args[0]
	addrq := &getWalletByAddressArgs{
		Address: addr,
	}
	cli := xfsgo.NewClient(config.rpcClientApiHost)
	var r *interface{} = nil
	err = cli.CallMethod(1, "Wallet.Del", addrq, &r)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	fmt.Println("delete Wallet success")
	return nil
}

func runWalletExport(cmd *cobra.Command, args []string) error {
	config, err := parseClientConfig(cfgFile)
	if err != nil {
		return err
	}
	addr := args[0]
	addrq := &getWalletByAddressArgs{
		Address: addr,
	}
	cli := xfsgo.NewClient(config.rpcClientApiHost)
	var r *string = nil
	err = cli.CallMethod(1, "Wallet.ExportByAddress", addrq, &r)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	fmt.Printf("%s\n", *r)
	return nil
}

func runWalletImport(cmd *cobra.Command, args []string) error {
	config, err := parseClientConfig(cfgFile)
	if err != nil {
		fmt.Println(err)
		return err
	}
	addr := args[0]
	importrq := &walletImportArgs{
		Key: addr,
	}
	cli := xfsgo.NewClient(config.rpcClientApiHost)
	var r *string = nil
	err = cli.CallMethod(1, "Wallet.ImportByPrivateKey", importrq, &r)
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Printf("%s\n", *r)
	return nil
}

func setWalletAddrDef(cmd *cobra.Command, args []string) error {
	config, err := parseClientConfig(cfgFile)
	if err != nil {
		return err
	}
	cli := xfsgo.NewClient(config.rpcClientApiHost)
	addr := args[0]

	walletAddress := make([]common.Address, 0)
	err = cli.CallMethod(1, "Wallet.List", nil, &walletAddress)
	if err != nil {
		fmt.Println(err)
		return err
	}
	for _, item := range walletAddress {
		if item.String() == addr {
			req := &setWalletAddrDefArgs{
				Address: addr,
			}
			var r *string = nil
			err = cli.CallMethod(1, "Wallet.SetDefaultAddress", req, &r)
			if err != nil {
				fmt.Println(err)
				return err
			}
			fmt.Println(*r)
			return nil
		}
	}
	fmt.Println("Wallet address does not exist")
	return nil
}

func getWalletAddrDef() error {
	config, err := parseClientConfig(cfgFile)
	if err != nil {
		return err
	}
	cli := xfsgo.NewClient(config.rpcClientApiHost)
	var defStr *string = nil
	err = cli.CallMethod(1, "Wallet.GetDefaultAddress", nil, &defStr)
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println(*defStr)
	return nil
}

func runWalletList() error {
	config, err := parseClientConfig(cfgFile)
	if err != nil {
		return err
	}
	//Get wallet default address
	var defAddr common.Address
	cli := xfsgo.NewClient(config.rpcClientApiHost)
	err = cli.CallMethod(1, "Wallet.GetDefaultAddress", nil, &defAddr)
	if err != nil {
		fmt.Println(err)
		return err
	}
	// get height and hash
	block := make(map[string]interface{}, 1)
	err = cli.CallMethod(1, "Chain.Head", nil, &block)
	if err != nil {
		fmt.Println(err)
		return err
	}
	hash := block["header"].(map[string]interface{})["state_root"].(string)

	// wallet list
	walletAddress := make([]common.Address, 0)
	err = cli.CallMethod(1, "Wallet.List", nil, &walletAddress)
	if err != nil {
		fmt.Println(err)
		return err
	}

	// Wallet balance
	// getBalance
	balance := make(map[string]interface{}, 1)
	fmt.Print("Address                            Balance                       Default")
	fmt.Println()
	for _, w := range walletAddress {

		req := &getAccountArgs{
			RootHash: hash,
			Address:  w.B58String(),
		}
		err = cli.CallMethod(1, "State.GetStateObj", &req, &balance)
		if err != nil {
			fmt.Println(err)
			return err
		}

		var bal *big.Int = nil
		balanceStr := balance["balance"].(string)
		if balanceStr == "" {
			bal = new(big.Int).SetUint64(0)
		}

		bs, err := hex.DecodeString(balanceStr)
		if err != nil {
			return err
		}

		bal = new(big.Int).SetBytes(bs)
		fto := common.Atto2BaseCoin(bal)
		toFloat := new(big.Float).SetUint64(fto.Uint64())
		fmt.Printf("%-35v", w.B58String())
		fmt.Printf("%-30.9f", toFloat)

		if w == defAddr {
			fmt.Printf("%-10v", "x")
		}
		fmt.Println()
	}
	return nil
}

func SetGasPrice(cmd *cobra.Command, args []string) error {
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
	req := &SetGasPriceArgs{
		GasPrice: r.String(),
	}
	err = cli.CallMethod(1, "Wallet.SetGasPrice", &req, &res)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	return nil
}

func SetGas(cmd *cobra.Command, args []string) error {
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
	gasLimit, _ := new(big.Int).SetString(args[0], 0)
	r := common.NanoCoin2BaseCoin(gasLimit)
	req := &GasLimitArgs{
		Gas: r.String(),
	}
	err = cli.CallMethod(1, "Wallet.SetGasLimit", &req, &res)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	return nil
}

func init() {
	walletCommand.AddCommand(walletListCommand)
	walletCommand.AddCommand(walletNewCommand)
	walletCommand.AddCommand(walletDelCommand)
	walletCommand.AddCommand(walletImportCommand)
	walletCommand.AddCommand(walletExportCommand)
	walletCommand.AddCommand(walletGetAddrDefCommand)
	walletCommand.AddCommand(walletTransferCommand)
	mFlags := walletTransferCommand.PersistentFlags()
	mFlags.StringVarP(&fromAddr, "address", "f", "", "Set miner income address")
	mFlags.StringVarP(&gasPrice, "gasprice", "", "", "Set expected gas unit price")
	mFlags.StringVarP(&gasLimit, "gaslimit", "", "", "Set maximum gas purchase quantity")
	walletCommand.AddCommand(walletSetAddrDefCommand)
	walletCommand.AddCommand(walletSetGasCommand)
	walletCommand.AddCommand(walletSetGasLimitCommand)
	rootCmd.AddCommand(walletCommand)
}
