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
	getStateCommand = &cobra.Command{
		Use:   "state",
		Short: "get state info",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	getAccountCommand = &cobra.Command{
		Use:   "getaccount --roothash [hash] <address>",
		Short: "Specifies the hash value of the world state tree root",
		RunE:  GetAccount,
	}
)

func GetAccount(cmd *cobra.Command, args []string) error {
	var RootHash, Address string
	if len(args) < 1 {
		return cmd.Help()
	}
	Address = args[0]
	if len(args) > 1 {
		RootHash = args[0]
		Address = args[1]
	}

	config, err := parseClientConfig(cfgFile)
	if err != nil {
		return err
	}

	cli := xfsgo.NewClient(config.rpcClientApiHost)
	result := make(map[string]interface{}, 1)
	req := &getAccountArgs{
		RootHash: RootHash,
		Address:  Address,
	}
	err = cli.CallMethod(1, "State.GetStateObj", &req, &result)
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

func init() {
	rootCmd.AddCommand(getStateCommand)
	getStateCommand.AddCommand(getAccountCommand)
}
