package sub

import (
	"fmt"
	"xblockchain"
	"xblockchain/rpc"

	"github.com/spf13/cobra"
)

var (
	getBlockCommand = &cobra.Command{
		Use:  "block <block_hash>",
		RunE: runGetBlock,
	}
)

type blockReq struct {
	Hash string
}

func runGetBlock(cmd *cobra.Command, args []string) error {
	if len(args) == 0 || len(args) > 1 {
		return cmd.Help()
	}
	hash := args[0]
	blockrq := &blockReq{
		Hash: hash,
	}
	cli := rpc.NewClient(GetConfigSub().Network.ClientAPIAddress)
	var r *xblockchain.Block = nil
	err := cli.CallMethod(1, "Chain.GetBlockByHash", blockrq, &r)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	fmt.Printf("%s\n", r)
	return nil
}

func init() {
	getCommand.AddCommand(getBlockCommand)
}
