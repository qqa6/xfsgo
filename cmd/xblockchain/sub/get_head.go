package sub

import (
	"fmt"
	"xblockchain/rpc"

	"github.com/spf13/cobra"
)

var (
	getHeadCommand = &cobra.Command{
		Use:  "head",
		RunE: runGetHead,
	}
)

func runGetHead(cmd *cobra.Command, args []string) error {
	cli := rpc.NewClient(GetConfigSub().Network.ClientAPIAddress)
	var r *string = nil
	err := cli.CallMethod(1, "Chain.LastBlockHash", nil, &r)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	if r == nil {
		fmt.Println("not found")
		return err
	}
	fmt.Printf("%s\n", *r)
	return nil
}

func init() {
	getCommand.AddCommand(getHeadCommand)
}
