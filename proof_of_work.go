package xblockchain

import (
	"bytes"
	"crypto/sha256"
	"strconv"
	"xblockchain/uint256"
)

const targetBits = 22

type ProofOfWork struct {
	Block *Block
	Target *uint256.UInt256
}

func NewProofOfWork(block *Block) *ProofOfWork {
	target := uint256.NewUInt256One()
	target = target.Lsh(256-targetBits)
	return &ProofOfWork{
		Block: block,
		Target: target,
	}
}
func (proofOfWork ProofOfWork) PrepareData(nonce *uint256.UInt256) []byte{
	block := proofOfWork.Block
	timestamp := []byte(strconv.FormatInt(block.Timestamp, 10))
	targetBytes := []byte(strconv.FormatInt(targetBits, 10))
	data := bytes.Join([][]byte{
		block.HashPrevBlock.ToBytes(),
		block.HashBytesTransactions(),
		timestamp,
		targetBytes,
		nonce.ToBytes(),
	},[]byte{})
	return data
}

func (proofOfWork ProofOfWork) Run() (*uint256.UInt256, *uint256.UInt256){
	nonce := uint256.NewUInt256Zero()
	hashint := uint256.NewUInt256Zero()
	for nonce.Lt(uint256.NewUInt256Max()) {
		data := proofOfWork.PrepareData(nonce)
		hash := sha256.Sum256(data)
		hashint = uint256.NewUInt256BS(hash[:])
		if hashint.Lt(proofOfWork.Target) {
			break
		}else {
			nonce = nonce.Add(uint256.NewUInt256One())
		}
	}
	return nonce,hashint
}

func (proofOfWork ProofOfWork) Validate() bool {
	nonce := proofOfWork.Block.Nonce
	data := proofOfWork.PrepareData(nonce)
	hash := sha256.Sum256(data)
	hashint := uint256.NewUInt256BS(hash[:])
	return hashint.Lt(proofOfWork.Target)
}