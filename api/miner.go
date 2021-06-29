package api

import (
	"xblockchain"
	"xblockchain/rpc/errors"
)

type MinerAPIHandler struct {
	Miner *xblockchain.Miner
}

func (receiver *MinerAPIHandler) Start(_ EmptyArg, _ *interface{}) error {
	err := receiver.Miner.Run()
	if err != nil {
		return errors.New(-32001, err.Error())
	}
	return nil
}

func (receiver *MinerAPIHandler) Stop(_ EmptyArg, _ *interface{}) error {
	err := receiver.Miner.Stop()
	if err != nil {
		return errors.New(-32001, err.Error())
	}
	return nil
}

func (receiver *MinerAPIHandler) ListTXPendingPool(_ EmptyArg, resp *[]xblockchain.Transaction) error {
	pool := receiver.Miner.TxPendingPool
	txs := make([]xblockchain.Transaction, 0)
	pool.Foreach(func(tx *xblockchain.Transaction) {
		txs = append(txs, *tx)
	})
	*resp = txs
	return nil
}

func (receiver *MinerAPIHandler) GetTXPendingPoolCount(_ EmptyArg, resp *int) error {
	pool := receiver.Miner.TxPendingPool
	c := pool.Count()
	*resp = c
	return nil
}
