package sub

import (
	"fmt"
	"xblockchain/rpc"

	"github.com/spf13/cobra"
)

var (
	minerCommand = &cobra.Command{
		Use: "miner",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}
			return nil
		},
	}
	minerStartCommand = &cobra.Command{
		Use:  "start",
		RunE: runMinerStart,
	}
	minerStopCommand = &cobra.Command{
		Use:  "stop",
		RunE: runMinerStop,
	}
)

func runMinerStart(_ *cobra.Command, _ []string) error {
	cli := rpc.NewClient(GetConfigSub().Network.ClientAPIAddress)
	var r *interface{} = nil
	err := cli.CallMethod(1, "Miner.Start", nil, &r)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	fmt.Println("miner running...")
	return nil
}

func runMinerStop(_ *cobra.Command, _ []string) error {
	cli := rpc.NewClient(GetConfigSub().Network.ClientAPIAddress)
	var r *interface{} = nil
	err := cli.CallMethod(1, "Miner.Stop", nil, &r)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	fmt.Println("miner stop...")
	return nil
}

func init() {
	minerCommand.AddCommand(minerStartCommand)
	minerCommand.AddCommand(minerStopCommand)
	rootCmd.AddCommand(minerCommand)
}
