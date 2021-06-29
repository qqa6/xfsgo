package sub

import (
	"fmt"
	"xblockchain/rpc"

	"github.com/spf13/cobra"
)

var (
	walletCommand = &cobra.Command{
		Use: "wallet",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	walletListCommand = &cobra.Command{
		Use: "list",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWalletList()
		},
	}
	walletNewCommand = &cobra.Command{
		Use: "new",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWalletNew()
		},
	}
	walletDelCommand = &cobra.Command{
		Use:  "del <address>",
		RunE: runWalletDel,
	}
	walletExportCommand = &cobra.Command{
		Use:  "export <address>",
		RunE: runWalletExport,
	}
	walletImportCommand = &cobra.Command{
		Use:  "import <private_key>",
		RunE: runWalletImport,
	}
)

type walletReq struct {
	Address string
}
type walletImportReq struct {
	PrivateKey string
}

type walletResp struct {
	Address string `json:"address"`
}

func runWalletNew() error {
	cli := rpc.NewClient(GetConfigSub().Network.ClientAPIAddress)
	var addr *string = nil
	err := cli.CallMethod(1, "Wallet.New", nil, &addr)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	fmt.Println(*addr)
	return nil
}

func runWalletDel(cmd *cobra.Command, args []string) error {
	if len(args) == 0 || len(args) > 1 {
		return cmd.Help()
	}
	addr := args[0]
	addrq := &walletReq{
		Address: addr,
	}
	cli := rpc.NewClient(GetConfigSub().Network.ClientAPIAddress)
	var r *interface{} = nil
	err := cli.CallMethod(1, "Wallet.Del", addrq, &r)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	return nil
}

func runWalletExport(cmd *cobra.Command, args []string) error {
	if len(args) == 0 || len(args) > 1 {
		return cmd.Help()
	}
	addr := args[0]
	addrq := &walletReq{
		Address: addr,
	}
	cli := rpc.NewClient(GetConfigSub().Network.ClientAPIAddress)
	var r *string = nil
	err := cli.CallMethod(1, "Wallet.ExportByAddress", addrq, &r)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	fmt.Printf("%s\n", *r)
	return nil
}

func runWalletImport(cmd *cobra.Command, args []string) error {
	if len(args) == 0 || len(args) > 1 {
		return cmd.Help()
	}
	addr := args[0]
	importrq := &walletImportReq{
		PrivateKey: addr,
	}
	cli := rpc.NewClient(GetConfigSub().Network.ClientAPIAddress)
	var r *string = nil
	err := cli.CallMethod(1, "Wallet.ImportByPrivateKey", importrq, &r)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	fmt.Printf("%s\n", *r)
	return nil
}

func runWalletList() error {
	var defAddr *string = nil
	cli := rpc.NewClient(GetConfigSub().Network.ClientAPIAddress)
	err := cli.CallMethod(1, "Wallet.GetDefaultAddress", nil, &defAddr)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	wallets := make([]*walletResp, 0)
	err = cli.CallMethod(1, "Wallet.List", nil, &wallets)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	fmt.Printf("Address                                       Default\n")
	for _, w := range wallets {
		fmt.Printf("%-45s ", w.Address)
		if w.Address == *defAddr {
			fmt.Print("x ")
		}
		fmt.Println()
	}
	return nil
}

func init() {
	walletCommand.AddCommand(walletListCommand)
	walletCommand.AddCommand(walletNewCommand)
	walletCommand.AddCommand(walletDelCommand)
	walletCommand.AddCommand(walletImportCommand)
	walletCommand.AddCommand(walletExportCommand)
	rootCmd.AddCommand(walletCommand)
}
