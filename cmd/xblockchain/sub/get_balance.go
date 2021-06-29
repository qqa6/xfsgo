package sub

import (
	"fmt"
	"xblockchain/rpc"

	"github.com/spf13/cobra"
)

var (
	getBalanceCommand = &cobra.Command{
		Use:  "balance <address>",
		RunE: runGetBalance,
	}
)

type addressReq struct {
	Address string
}

func runGetBalance(cmd *cobra.Command, args []string) error {
	if len(args) == 0 || len(args) > 1 {
		return cmd.Help()
	}
	addr := args[0]
	addrq := &addressReq{
		Address: addr,
	}
	cli := rpc.NewClient(GetConfigSub().Network.ClientAPIAddress)
	var r *uint64 = nil
	err := cli.CallMethod(1, "Chain.GetBalance", addrq, &r)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	fmt.Printf("%d\n", *r)
	return nil
}

func init() {
	getCommand.AddCommand(getBalanceCommand)
}
