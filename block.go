package xblockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"
	"xblockchain/uint256"
)

type Block struct {
	Height uint64 `json:"height"`
	Timestamp int64 `json:"timestamp"`
	HashPrevBlock *uint256.UInt256 `json:"hash_prev_block"`
	Nonce *uint256.UInt256 `json:"nonce"`
	Hash *uint256.UInt256 `json:"hash"`
	Transactions []*Transaction `json:"transactions"`
}

func NewGenesisBlock(tx *Transaction, t int64) *Block {
	return NewBlockFromTime(nil,[]*Transaction{tx},t)
}

func NewBlockFromTime(prevBlock *Block, txs []*Transaction, t int64) *Block {
	height := uint64(0)
	prevBlockHash := uint256.NewUInt256Zero()
	if prevBlock != nil {
		height = prevBlock.Height+1
		prevBlockHash = prevBlock.Hash
	}
	block := &Block{
		Height: height,
		Timestamp: t,
		HashPrevBlock: prevBlockHash,
		Transactions: txs,
	}
	if block.Height == 0 {
		fmt.Printf("create genesis block...\n")
	}
	pow := NewProofOfWork(block)
	nonce, hash  := pow.Run()
	block.Nonce = nonce
	block.Hash = hash
	fmt.Printf("mining block success height: %d, hash: %s, txcount: %d\n", block.Height,hash.Hex(),len(txs))
	return block
}



func NewBlock(prevBlock *Block, txs []*Transaction) *Block {
	return NewBlockFromTime(prevBlock, txs, time.Now().Unix())
}

func (block *Block) Serialize() ([]byte,error) {
	jsonbs, err := json.Marshal(block)
	if err != nil {
		return nil,err
	}
	return jsonbs,nil
}

func (block *Block) HashBytesTransactions() []byte {
	var txBytes [][]byte
	for i := 0; i < len(block.Transactions); i++ {
		tx := block.Transactions[i]
		txBytes = append(txBytes, tx.ID.ToBytes())
	}
	txHash := sha256.Sum256(bytes.Join(txBytes, []byte{}))
	return txHash[:]
}

func ReadBlockOfBytes(bs []byte) (*Block,error) {
	block := &Block{}
	err := json.Unmarshal(bs,block)
	if err != nil {
		return nil,err
	}
	return block,nil
}
func (block *Block) String() string {
	jsonByte,err := json.Marshal(block)
	if err != nil {
		return ""
	}
	return string(jsonByte)
}

func (block *Block) GetMiner() string {
	txs := block.Transactions
	for _, tx := range txs {
		if tx.IsCoinBase() {
			txout := tx.Outputs[0]
			return PubKeyHash2Address(txout.PubKeyHash)
		}
	}
	return ""
}



