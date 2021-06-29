package api

import (
	"xblockchain"
	error2 "xblockchain/rpc/errors"
)


type WalletsHandler struct {
	Wallets *xblockchain.Wallets
}

type GetWalletByAddressArg struct {
	Address string
}

type WalletOutResp struct {
	Address string `json:"address"`
}

type WalletDelArgs struct {
	Address string
}

type WalletImportArg struct {
	PrivateKey string
}

func (receiver *WalletsHandler) New(args EmptyArg, resp *string) error {
	addr, err := receiver.Wallets.NewWallet()
	if err != nil {
		return error2.New(-32001, err.Error())
	}
	*resp = addr
	return nil
}

func (receiver *WalletsHandler) Del(args WalletDelArgs,resp *interface{}) error {
	err := receiver.Wallets.Delete(args.Address)
	if err != nil {
		return error2.New(-32001, err.Error())
	}
	return nil
}

func (receiver *WalletsHandler) List(args EmptyArg, resp *[]*WalletOutResp) error {
	list := receiver.Wallets.List()
	out := make([]*WalletOutResp,0)
	for _, wallet := range list {
		out = append(out, &WalletOutResp{
			Address: wallet.GetAddress(),
		})
	}
	*resp = out
	return nil
}


func (receiver *WalletsHandler) GetDefaultAddress(args EmptyArg, resp *string) error {
	address := receiver.Wallets.GetDefault()
	*resp = address
	return nil
}


func (receiver *WalletsHandler) GetWalletByAddress(args GetWalletByAddressArg, resp *xblockchain.Wallet) error {
	w, err:= receiver.Wallets.GetWalletByAddress(args.Address)
	if err != nil {
		return error2.New(-32001, err.Error())
	}
	resp = w
	return nil
}

func (receiver *WalletsHandler) ExportByAddress(args GetWalletByAddressArg, resp *string) error {
	w, err:= receiver.Wallets.Export(args.Address)
	if err != nil {
		return error2.New(-32001, err.Error())
	}
	*resp = w
	return nil
}

func (receiver *WalletsHandler) ImportByPrivateKey(args WalletImportArg, resp *string) error {
	addr, err:= receiver.Wallets.Import(args.PrivateKey)
	if err != nil {
		return error2.New(-32001, err.Error())
	}
	*resp = addr
	return nil
}