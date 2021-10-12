package sub

import (
	"encoding/json"
	"fmt"
	"xfsgo"

	"github.com/spf13/cobra"
)

var (
	netCommand = &cobra.Command{
		Use:   "net",
		Short: "Network related operations",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	getPeersCommand = &cobra.Command{
		Use:   "peers",
		Short: "View established peer-to-peer links",
		RunE:  getPeers,
	}
	addPeerCommand = &cobra.Command{
		Use:   "addpeer <url>",
		Short: "Add peer-to-peer link",
		RunE:  addPeer,
	}
	delPeerCommand = &cobra.Command{
		Use:   "delpeer <node_id>",
		Short: "Add peer-to-peer link",
		RunE:  delPeer,
	}
	getNodeIdCommand = &cobra.Command{
		Use:   "getid",
		Short: "View the ID of the current node",
		RunE:  getNodeId,
	}
)

func getPeers(cmd *cobra.Command, args []string) error {
	config, err := parseClientConfig(cfgFile)
	if err != nil {
		return err
	}
	res := make([]string, 0)
	cli := xfsgo.NewClient(config.rpcClientApiHost, config.rpcClientApiTimeOut)
	err = cli.CallMethod(1, "Net.GetPeers", nil, &res)
	if err != nil {
		return err
	}
	if len(res) > 1 {
		bs, err := json.Marshal(res)
		if err != nil {
			return err
		}
		fmt.Println(string(bs))
	}
	return nil
}

func addPeer(cmd *cobra.Command, args []string) error {
	// fmt.Printf("niho")
	if len(args) < 1 {
		return cmd.Help()
	}
	config, err := parseClientConfig(cfgFile)
	if err != nil {
		return err
	}
	var res string
	cli := xfsgo.NewClient(config.rpcClientApiHost, config.rpcClientApiTimeOut)
	req := &AddPeerArgs{
		Url: args[0],
	}
	err = cli.CallMethod(1, "Net.AddPeer", &req, &res)
	if err != nil {
		return err
	}
	return nil
}

func delPeer(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return cmd.Help()
	}
	config, err := parseClientConfig(cfgFile)
	if err != nil {
		return err
	}
	var res string
	cli := xfsgo.NewClient(config.rpcClientApiHost, config.rpcClientApiTimeOut)
	req := &DelPeerArgs{
		Id: args[0],
	}
	err = cli.CallMethod(1, "Net.DelPeer", &req, &res)
	if err != nil {
		return err
	}
	return nil
}

func getNodeId(cmd *cobra.Command, args []string) error {
	config, err := parseClientConfig(cfgFile)
	if err != nil {
		return err
	}
	var res string
	cli := xfsgo.NewClient(config.rpcClientApiHost, config.rpcClientApiTimeOut)
	err = cli.CallMethod(1, "Net.GetNodeId", nil, &res)
	if err != nil {
		return err
	}
	fmt.Printf("%v\n", res)
	return nil
}

func init() {
	rootCmd.AddCommand(netCommand)
	netCommand.AddCommand(getPeersCommand)
	netCommand.AddCommand(addPeerCommand)
	netCommand.AddCommand(delPeerCommand)
	netCommand.AddCommand(getNodeIdCommand)
}
