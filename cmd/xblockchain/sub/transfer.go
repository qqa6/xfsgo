package sub

import (
	"encoding/json"
	"fmt"
	"xblockchain"
	"xblockchain/api"
	"xblockchain/rpc"

	"github.com/spf13/cobra"
)

var (
	transferCommand = &cobra.Command{
		Use:  "transfer [from] <to> <value>",
		RunE: runTransfer,
	}
)

func runTransfer(cmd *cobra.Command, args []string) error {
	if len(args) == 0 || len(args) < 2 || len(args) > 3 {
		return cmd.Help()
	}
	argc := len(args)
	fromAddr := ""
	toAddr := ""
	value := ""
	if argc == 2 {
		fromAddr = ""
		toAddr = args[0]
		value = args[1]
	} else {
		fromAddr = args[0]
		toAddr = args[1]
		value = args[2]
	}
	arg := &api.SendTransactionArg{
		From:  fromAddr,
		To:    toAddr,
		Value: value,
	}
	cli := rpc.NewClient(GetConfigSub().Network.ClientAPIAddress)
	var r *xblockchain.Transaction = nil
	err := cli.CallMethod(1, "Transaction.SendTransaction", arg, &r)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	jsonbyte, err := json.Marshal(*r)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	fmt.Printf("%s\n", string(jsonbyte))
	return nil
}

func init() {
	rootCmd.AddCommand(transferCommand)
}
