package api

import (
	"math/big"
	"testing"
	"time"
	"xfsgo"
	"xfsgo/assert"
	"xfsgo/common"
)
type BlockHeader struct {
	Height        uint64         `json:"height"`
	Version       uint32         `json:"version"`
	HashPrevBlock common.Hash    `json:"hash_prev_block"`
	Timestamp     uint64         `json:"timestamp"`
	Coinbase      common.Address `json:"coinbase"`
	// merkle tree root hash
	StateRoot        common.Hash `json:"state_root"`
	TransactionsRoot common.Hash `json:"transactions_root"`
	ReceiptsRoot     common.Hash `json:"receipts_root"`
	GasLimit         *big.Int    `json:"gas_limit"`
	GasUsed          *big.Int    `json:"gas_used"`
	// pow consensus.
	Bits  uint32 `json:"bits"`
	Nonce uint64 `json:"nonce"`
}

func makeBlock(height uint64) *xfsgo.Block {
	timestamp := time.Now().Unix()
	header := &xfsgo.BlockHeader{
		Height: height,
		Version: 0,
		HashPrevBlock: common.Bytes2Hash([]byte{0xff,0xff,0xff}),
		Timestamp: uint64(timestamp),
		Coinbase: common.Bytes2Address([]byte{0xff,0xff,0xff}),
		StateRoot: common.Bytes2Hash([]byte{0xff,0xff,0xff}),
		TransactionsRoot: common.Bytes2Hash([]byte{0xff,0xff,0xff}),
		ReceiptsRoot: common.Bytes2Hash([]byte{0xff,0xff,0xff}),
		GasLimit: new(big.Int).SetInt64(0xff),
		GasUsed: new(big.Int).SetInt64(0xff),
		Bits: 655397,
		Nonce: 0,
	}
	return xfsgo.NewBlock(header, nil,nil)
}


func Test_coverBlock2Resp(t *testing.T) {
	blk := makeBlock(456)
	var resp *BlockResp
	err := coverBlock2Resp(blk, &resp)
	assert.Error(t, err)
	assert.Equal(t,resp.Height, blk.Header.Height)
	assert.Equal(t, resp.Version, blk.Header.Version)
	assert.Equal(t, resp.HashPrevBlock, blk.Header.HashPrevBlock)
	assert.Equal(t, resp.Timestamp, blk.Header.Timestamp)
	assert.Equal(t, resp.Coinbase, blk.Header.Coinbase)
	assert.Equal(t, resp.StateRoot, blk.Header.StateRoot)
	assert.Equal(t, resp.TransactionsRoot, blk.Header.TransactionsRoot)
	assert.Equal(t, resp.ReceiptsRoot, blk.Header.ReceiptsRoot)
	assert.Equal(t, resp.GasLimit, blk.Header.GasLimit)
	assert.Equal(t, resp.GasUsed, blk.Header.GasUsed)
	assert.Equal(t, resp.Bits, blk.Header.Bits)
	assert.Equal(t, resp.Nonce, blk.Header.Nonce)
	assert.Equal(t, resp.Hash, blk.Hash())
	assert.Equal(t, resp.Transactions, blk.Transactions)
	assert.Equal(t, resp.Receipts, blk.Receipts)
}