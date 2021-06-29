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
	getTxCommand = &cobra.Command{
		Use:  "tx <id>",
		RunE: runGetTx,
	}
)

func runGetTx(cmd *cobra.Command, args []string) error {
	if len(args) == 0 || len(args) > 1 {
		return cmd.Help()
	}
	idStr := args[0]
	cli := rpc.NewClient(GetConfigSub().Network.ClientAPIAddress)
	reqArgs := &api.GetTransactionArgs{
		Id: idStr,
	}
	var r *xblockchain.Transaction = nil
	err := cli.CallMethod(1, "Chain.GetTransaction", reqArgs, &r)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	tx := *r
	jsonbyte, err := json.Marshal(tx)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	fmt.Printf("%s\n", string(jsonbyte))
	return nil
}

func init() {
	getCommand.AddCommand(getTxCommand)
}
