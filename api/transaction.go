package api

import (
	"strconv"
	"xblockchain"
	"xblockchain/rpc/errors"
)

type TXAPIHandler struct {
	Wallets *xblockchain.Wallets
	BlockChain *xblockchain.BlockChain
	TxPendingPool *xblockchain.TxPool
}


type SendTransactionArg struct {
	From string
	To string
	Value string
}



func (receiver *TXAPIHandler) SendTransaction(args SendTransactionArg, resp *xblockchain.Transaction) error {
	wallets := receiver.Wallets
	fromAddr := args.From
	if fromAddr == "" {
		fromAddr = wallets.GetDefault()
	}
	fromWallet, err := wallets.GetWalletByAddress(fromAddr)
	if err != nil {
		return errors.New(-32001, err.Error())
	}
	toAddr := args.To
	if toAddr == "" {
		return errors.New(-32001, "to address cannot be blank")
	}

	toValueStr := args.Value
	if toValueStr == "" {
		return errors.New(-32001, "value cannot be blank")
	}

	toValue,err := strconv.ParseUint(toValueStr,10,64)
	if err != nil {
		return errors.New(-32001, err.Error())
	}
	bc := receiver.BlockChain
	tx,err := xblockchain.NewUTXOTransaction(fromAddr,toAddr,toValue,bc,fromWallet.GetPrivateKey())
	if err != nil {
		return errors.New(-32001, err.Error())
	}
	pool := receiver.TxPendingPool
	if err = pool.AddTx(tx); err != nil {
		return errors.New(-32001, err.Error())
	}
	*resp = *tx
	return nil
}

func (receiver *TXAPIHandler) SendRawTransaction(_ EmptyArg, _ *interface{}) error {
	//err := receiver.Miner.Run()
	//if err != nil {
	//	return errors.New(-32001, err.Error())
	//}
	return nil
}