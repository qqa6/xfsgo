package sub

import (
	"fmt"
	"os"
	configs "xblockchain/cmd/config"

	"github.com/spf13/cobra"
)

//
//var (
//	RPCListenAddress = ":9002"
//	ClientAPIAddress = "http://127.0.0.1:9002"
//	P2PListenAddress = ":9001"
//	BlockDbPath = "./data0/blocks"
//	KeyStoragePath = "./data0/keys"
//	P2PBootstraps = []string{}
//	ProtocolVersion = uint32(0)
//	NetworkID = uint32(0)
//)
// var (
// 	RPCListenAddress = ":9004"
// 	ClientAPIAddress = "http://127.0.0.1:9004"
// 	P2PListenAddress = ":9003"
// 	BlockDbPath      = "./data1/blocks"
// 	KeyStoragePath   = "./data1/keys"
// 	P2PBootstraps    = []string{"127.0.0.1:9001"}
// 	ProtocolVersion  = uint32(0)
// 	NetworkID        = uint32(0)
// )

//var (
//	RPCListenAddress = ":9006"
//	ClientAPIAddress = "http://127.0.0.1:9006"
//	P2PListenAddress = ":9005"
//	BlockDbPath = "./data2/blocks"
//	KeyStoragePath = "./data2/keys"
//	P2PBootstraps = []string{"127.0.0.1:9001"}
//	ProtocolVersion = uint32(0)
//	NetworkID = uint32(1)
//)
var (
	cfgFile string
	rootCmd = &cobra.Command{
		Use: "fixcoin",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}
			return nil
		},
	}
	getCommand = &cobra.Command{
		Use: "get <command> [flags]",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func GetConfigSub() configs.ConfigInfo {
	return configs.GetConfig()
}
func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")
	rootCmd.AddCommand(getCommand)
}
