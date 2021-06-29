package sub

import (
	"fmt"
	"xblockchain"
	"xblockchain/rpc"

	"github.com/spf13/cobra"
)

var (
	getMEMPoolCommand = &cobra.Command{
		Use:  "mempool",
		RunE: runGetMEMPool,
	}
)

func runGetMEMPool(_ *cobra.Command, _ []string) error {

	cli := rpc.NewClient(GetConfigSub().Network.ClientAPIAddress)
	var r *[]xblockchain.Transaction = nil
	err := cli.CallMethod(1, "Miner.ListTXPendingPool", nil, &r)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	if r == nil {
		fmt.Println("not found")
		return err
	}
	txs := *r
	fmt.Printf("No  ID        \n")
	for i := 0; i < len(txs); i++ {
		tx := txs[i]
		txIdStr := tx.ID.HexstrFull()
		fmt.Printf("%-3d ", i)
		fmt.Printf("%-3s...%-3s ", txIdStr[:3], txIdStr[len(txIdStr)-3:])
		fmt.Println()
	}
	return nil
}

func init() {
	getCommand.AddCommand(getMEMPoolCommand)
}
