package api

import (
	"strconv"
	"xblockchain"
	"xblockchain/rpc/errors"
	"xblockchain/uint256"
)


var (
	notFoundBlockErr = errors.New(-32001, "Not found block")
)

type ChainAPIHandler struct {
	BlockChain *xblockchain.BlockChain
}


type GetBlockByIdArgs struct {
	Id string
}

type GetBlockByHashArgs struct {
	Hash string
}

type GetBalanceOfAddressArgs struct {
	Address string
}

type GetTransactionArgs struct {
	Id string
}

type ChainInfoResp struct {
	LastBlockHeight uint64 `json:"last_block_height"`
	LastBlockHash *uint256.UInt256 `json:"last_block_hash"`
	LastBlockTime uint64 `json:"last_block_time"`
}




func (receiver *ChainAPIHandler) GetBlockById(args GetBlockByIdArgs, block *xblockchain.Block) error {
	id, err := strconv.Atoi(args.Id)
	if err != nil {
		return errors.New(-32000,err.Error())
	}
	b,err := receiver.BlockChain.GetBlockById(uint64(id))
	if err != nil {
		return errors.New(-32001, err.Error())
	}
	*block = *b
	return nil
}



func (receiver *ChainAPIHandler) GetInfo(args EmptyArg, info *ChainInfoResp) error {
	hash := receiver.BlockChain.GetLastBlockHash()
	block,err := receiver.BlockChain.GetBlockByHash(hash)
	if err != nil {
		return errors.New(-32001, err.Error())
	}
	if block == nil {
		return notFoundBlockErr
	}
	blockHeight := block.Height
	blockTime := block.Timestamp
	*info = *&ChainInfoResp{
		LastBlockHash: hash,
		LastBlockHeight: blockHeight,
		LastBlockTime: uint64(blockTime),
	}
	return nil
}

func (receiver *ChainAPIHandler) LastBlockId(args EmptyArg, blockId *uint64) error {
	hash := receiver.BlockChain.GetLastBlockHash()
	block,err := receiver.BlockChain.GetBlockByHash(hash)
	if err != nil {
		return errors.New(-32001, err.Error())
	}
	n := block.Height
	*blockId = n
	return nil
}

func (receiver *ChainAPIHandler) LastBlockHash(args EmptyArg, blockId *string) error {
	hash := receiver.BlockChain.GetLastBlockHash()
	if hash.Equals(uint256.NewUInt256Zero()) {
		return errors.New(-32001, "not found")
	}
	*blockId = hash.Hex()
	return nil
}

func (receiver *ChainAPIHandler) GetBlockByHash(args GetBlockByHashArgs, block *xblockchain.Block) error {
	hash := uint256.NewUInt256(args.Hash)
	data,err := receiver.BlockChain.GetBlockByHash(hash)
	if err != nil {
		return errors.New(-32001, err.Error())
	}
	*block = *data
	return nil
}


func (receiver *ChainAPIHandler) GetBalance(args GetBalanceOfAddressArgs, resp *uint64) error {
	balance,err := receiver.BlockChain.GetBalanceOfAddress(args.Address)
	if err != nil {
		return errors.New(-32001, err.Error())
	}
	*resp =  balance
	return nil
}

func (receiver *ChainAPIHandler) GetTransaction(args GetTransactionArgs, resp *xblockchain.Transaction) error {
	txId := args.Id
	txIdUint := uint256.NewUInt256(txId)
	tx,err := receiver.BlockChain.FindTransaction(txIdUint)
	if err != nil {
		return errors.New(-32001, err.Error())
	}
	*resp =  *tx
	return nil
}