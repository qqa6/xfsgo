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
	"math"
	"xfsgo"

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
	getStateObjCommand = &cobra.Command{
		Use:   "getstateobj <roothash> <address>",
		Short: "get state object  <roothash>  <address>",
		RunE:  getStateObj,
	}
)

func getStateObj(cmd *cobra.Command, args []string) error {
	if len(args) != 2 {
		return cmd.Help()
	}

	config, err := parseClientConfig(cfgFile)
	if err != nil {
		return err
	}

	cli := xfsgo.NewClient(config.rpcClientApiHost)
	balance := make(map[string]interface{}, 1)
	req := &getStateObjArgs{
		RootHash: args[0],
		Address:  args[1],
	}
	err = cli.CallMethod(1, "State.GetStateObj", &req, &balance)
	if err != nil {
		fmt.Println(err)
		return err
	}

	var t uint64
	if balance["balance"] != nil {
		t = uint64(balance["balance"].(float64) * math.Pow10(0))
	}

	var nonce uint64
	if balance["nonce"].(float64) != float64(0) {
		nonce = uint64(balance["nonce"].(float64) * math.Pow10(0))
	}

	fmt.Print("address                                 balance          nonce")
	fmt.Println()
	fmt.Printf("%-40v", balance["address"])
	fmt.Printf("%-17v", t)
	fmt.Printf("%-d\n", nonce)
	return nil

}

func init() {
	rootCmd.AddCommand(getStateCommand)
	getStateCommand.AddCommand(getStateObjCommand)
}
