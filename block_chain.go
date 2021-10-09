// Copyright 2018 The xfsgo Authors
// This file is part of the xfsgo library.
//
// The xfsgo library is free software: you can redistribute it and/or modify
// it under the terms of the MIT Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The xfsgo library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// MIT Lesser General Public License for more details.
//
// You should have received a copy of the MIT Lesser General Public License
// along with the xfsgo library. If not, see <https://mit-license.org/>.

package xfsgo

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"
	"xfsgo/common"
	"xfsgo/common/rawencode"
	"xfsgo/lru"
	"xfsgo/storage/badger"

	"github.com/sirupsen/logrus"
)

var zeroBigN = new(big.Int).SetInt64(0)

const (
	// blocks can be created per second(in seconds)
	targetTimePerBlock = int64(time.Minute * 10 / time.Second)
	// adjustment factor
	adjustmentFactor = int64(4)
	// target value you want blocks to be created per second（in seconds）
	targetTimespan = int64(time.Hour * 24 * 14 / time.Second)
)

// BlockChain represents the canonical chain given a database with a genesis
// block. The Blockchain manages chain inserts, saves, transfers state.
// The BlockChain also helps in returning blocks required from any chain included
// in the database as well as blocks that represents the canonical chain.

type BlockChain struct {
	stateDB       *badger.Storage
	chainDB       *chainDB
	extraDB       *extraDB
	genesisBlock  *Block
	currentBlock  *Block
	lastBlockHash common.Hash
	stateTree     *StateTree
	mu            sync.RWMutex
	chainmu       sync.RWMutex
	orphansCache  *lru.Cache
	eventBus      *EventBus
}

//NewBlockChain creates a initialised block chain using information available in the database.
//this new blockchain includes a stateTree by which the blockchian can manage the whole state of the chainb.
//such as the account's information of every user.
func NewBlockChain(stateDB, chainDB, extraDB *badger.Storage, eventBus *EventBus) (*BlockChain, error) {
	bc := &BlockChain{
		chainDB:  newChainDB(chainDB),
		stateDB:  stateDB,
		extraDB:  newExtraDB(extraDB),
		eventBus: eventBus,
	}
	bc.orphansCache = lru.NewCache(2048)
	bc.genesisBlock = bc.GetBlockByNumber(0)
	if bc.genesisBlock == nil {
		return nil, errors.New("no genesis block")
	}
	if err := bc.setLastState(); err != nil {
		return nil, err
	}
	stateRootHash := bc.currentBlock.StateRoot()
	bc.stateTree = NewStateTree(stateDB, stateRootHash.Bytes())
	return bc, nil
}

func (bc *BlockChain) GetNonce(addr common.Address) uint64 {
	return bc.stateTree.GetNonce(addr)
}

func (bc *BlockChain) GetBlockByNumber(num uint64) *Block {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	return bc.getBlockByNumber(num)
}

func (bc *BlockChain) getBlockByNumber(num uint64) *Block {
	return bc.chainDB.GetBlockByNumber(num)
}

func (bc *BlockChain) GetBlockByHash(hash common.Hash) *Block {
	return bc.chainDB.GetBlockByHash(hash)
}

func (bc *BlockChain) GetReceiptByHash(hash common.Hash) *Receipt {
	return bc.extraDB.GetReceipt(hash)
}

func (bc *BlockChain) GetBlocksFromHash(hash common.Hash, n int) []*Block {
	var blocks = make([]*Block, n)
	for i := 0; i < n; i++ {
		block := bc.GetBlockByHash(hash)
		if block == nil {
			break
		}
		blocks = append(blocks, block)
		hash = block.HashPrevBlock()
	}
	return blocks
}
func (bc *BlockChain) GenesisBlock() *Block {
	return bc.genesisBlock
}

func (bc *BlockChain) CurrentBlock() *Block {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	return bc.currentBlock
}

func (bc *BlockChain) LastBlockHash() common.Hash {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	return bc.lastBlockHash
}

func (bc *BlockChain) setLastState() error {
	if block := bc.chainDB.GetHeadBlock(); block != nil {
		bc.currentBlock = block
		bc.lastBlockHash = block.Hash()
	}
	return nil
}

func (bc *BlockChain) GetHead() *Block {
	return bc.chainDB.GetHeadBlock()
}
func (bc *BlockChain) GetBalance(addr common.Address) *big.Int {
	gotBalance := bc.stateTree.GetBalance(addr)
	return gotBalance

}

// WriteBlock stores the block inputed to the local database.
func (bc *BlockChain) WriteBlock(block *Block) error {
	cb := bc.currentBlock
	if block.Height() > cb.Height() {
		bc.mu.Lock()
		if err := bc.chainDB.WriteHead(block); err != nil {
			logrus.Errorf("write head err: %s", err)
		}
		bc.currentBlock = block
		bc.lastBlockHash = block.Hash()
		lastStateRoot := block.StateRoot()
		bc.stateTree = NewStateTree(bc.stateDB, lastStateRoot.Bytes())
		bc.mu.Unlock()
		bc.eventBus.Publish(ChainHeadEvent{block})
	}
	if err := bc.extraDB.WriteBlockTransaction(block); err != nil {
		return err
	}
	if err := bc.extraDB.WriteReceipts(block.Receipts); err != nil {
		return err
	}
	if err := bc.extraDB.WriteBlockReceipts(block); err != nil {
		return err
	}
	return bc.chainDB.WriteBlock(block)
}

func (bc *BlockChain) WriteBlockTransaction(block *Block) error {
	return bc.extraDB.WriteBlockTransaction(block)
}

func (bc *BlockChain) WriteBlockReceipts(block *Block) error {
	return bc.extraDB.WriteBlockReceipts(block)
}

func (bc *BlockChain) GetTransaction(Hash common.Hash) *Transaction {
	return bc.extraDB.GetTransactionByHash(Hash)
}

// calculate rewards for packing the block by miners
func calcBlockSubsidy(currentHeight uint64) *big.Int {
	// reduce the reward by half
	nSubsidy := uint64(50) >> uint(currentHeight/210000)
	//logrus.Debugf("nSubsidy: %d", nSubsidy)
	return common.BaseCoin2Atto(float64(nSubsidy))
}

// AccumulateRewards calculates the rewards and add it to the miner's account.
func AccumulateRewards(stateTree *StateTree, header *BlockHeader) {
	subsidy := calcBlockSubsidy(header.Height)
	//logrus.Debugf("current height of the blockchain %d, reward: %d", header.Height, subsidy)
	stateTree.AddBalance(header.Coinbase, subsidy)
}

// InsertChain executes the actual chain insertion.
func (bc *BlockChain) InsertChain(block *Block) error {
	bc.chainmu.Lock()
	defer bc.chainmu.Unlock()
	blockHash := block.Hash()
	txs := block.Transactions
	header := block.GetHeader()
	txsRoot := block.TransactionRoot()
	rsRoot := block.ReceiptsRoot()
	//logrus.Infof("Processing block %v", blockHash)
	if old := bc.GetBlockByHash(blockHash); old != nil {
		return fmt.Errorf("already have block %v", blockHash)
	}
	if _, has := bc.orphansCache.Get(blockHash); has {
		return fmt.Errorf("already have block (orphan) %v", blockHash)
	}
	if err := bc.checkBlockHeaderSanity(header, blockHash); err != nil {
		return err
	}
	var parent *Block
	if parent = bc.GetBlockByHash(block.HashPrevBlock()); parent == nil {
		encoded, _ := rawencode.Encode(block)
		bc.orphansCache.Put(blockHash, encoded)
		return nil
	}
	targetTxsRoot := CalcTxsRootHash(block.Transactions)
	if !bytes.Equal(targetTxsRoot.Bytes(), txsRoot.Bytes()) {
		return fmt.Errorf("check transaction root err")
	}
	parentStateRoot := parent.StateRoot()
	stateTree := NewStateTree(bc.stateDB, parentStateRoot.Bytes())
	_, rec, err := bc.ApplyTransactions(stateTree, header, txs)
	if err != nil {
		return err
	}
	block.Receipts = rec
	AccumulateRewards(stateTree, header)
	stateTree.UpdateAll()
	targetRsRoot := CalcReceiptRootHash(rec)
	if !bytes.Equal(rsRoot.Bytes(), targetRsRoot.Bytes()) {
		return fmt.Errorf("check receipt root err")
	}
	if err = stateTree.Commit(); err != nil {
		return err
	}
	if err = bc.WriteBlock(block); err != nil {
		return err
	}

	return nil
}

func (bc *BlockChain) ApplyTransactions(stateTree *StateTree, header *BlockHeader, txs []*Transaction) (*big.Int, []*Receipt, error) {
	receipts := make([]*Receipt, 0)
	totalUsedGas := big.NewInt(0)
	for _, tx := range txs {
		rec, err := bc.applyTransaction(stateTree, header, tx)
		if err != nil {
			logrus.Errorf("wrong to execute the transactions: %s err:%v", tx.Hash(), err.Error())
			return nil, nil, err
		}
		logrus.Infof("excute the transactions successfully: %s, receipt: %d", tx.Hash(), rec.Hash())
		if rec != nil {
			totalUsedGas.Add(big.NewInt(0), rec.GasUsed)
			receipts = append(receipts, rec)
		}
	}
	return totalUsedGas, receipts, nil
}

func (bc *BlockChain) checkBlockHeaderSanity(header *BlockHeader, blockHash common.Hash) error {
	target := BitsUnzip(header.Bits)
	if target.Sign() <= 0 {
		return fmt.Errorf("bits must be a non-negative integer")
	}
	max := BitsUnzip(bc.genesisBlock.Bits())
	//target difficuty should be less than the minimum difficuty based on the genesisBlock
	if target.Cmp(max) > 0 {
		return fmt.Errorf("pow check err")
	}
	current := new(big.Int).SetBytes(blockHash[:])
	// the current hash can not be larger than the target hash value
	if current.Cmp(target) > 0 {
		return fmt.Errorf("pow check err")
	}
	return nil
}

func (bc *BlockChain) checkTransactionSanity(tx *Transaction) error {
	if !tx.VerifySignature() {
		return fmt.Errorf("VerifySignature err")
	}
	return nil
}

// IntrinsicGas computes the 'intrisic gas' for a message
// with the given data.
func (bc *BlockChain) IntrinsicGas(_ []byte) *big.Int {
	igas := new(big.Int).Set(common.TxGas)
	return igas
}

func (bc *BlockChain) UseGas(gas, amount *big.Int) (*big.Int, error) {
	if gas.Cmp(amount) < 0 {
		return nil, errors.New("out of gas")
	}

	return gas.Sub(gas, amount), nil
}

func (bc *BlockChain) applyTransaction(stateTree *StateTree, header *BlockHeader, tx *Transaction) (*Receipt, error) {
	if err := bc.checkTransactionSanity(tx); err != nil {
		return nil, err
	}
	cb := stateTree.GetStateObj(header.Coinbase)
	// Pre-pay gas / Buy gas of the coinbase account
	cb.SetGasLimit(header.GasLimit)
	mgas := tx.GasLimit

	txGasCount := new(big.Int).Mul(mgas, tx.GasPrice)

	sender, err := tx.FromAddr()
	if err != nil {
		return nil, err
	}
	senderObj := stateTree.GetOrNewStateObj(sender) // get from Transaction state object
	if senderObj.GetBalance().Cmp(txGasCount) < 0 {
		return nil, fmt.Errorf("insufficient for gas (%x). Req %v, has %v", sender.Bytes()[:4], txGasCount, senderObj.GetBalance())
	}

	if err = cb.SubGas(mgas, tx.GasPrice); err != nil {
		return nil, err
	}

	gas := new(big.Int).Set(txGasCount) // user max gas
	initialGas := new(big.Int)
	initialGas.Set(mgas)
	senderObj.SubBalance(txGasCount) // Traders pre order gas

	// IntrinsicGas
	basicGas := bc.IntrinsicGas([]byte{}) // Basic service charge
	if mgas.Cmp(basicGas) < 0 {
		return nil, errors.New("out of gas")
	}

	surplus, err := bc.UseGas(gas, basicGas) // Actual gas consumption
	if err != nil {
		return nil, err
	}

	senderObj.AddBalance(surplus)
	cb.AddBalance(basicGas)

	stateTree.AddNonce(sender, 1)

	if err = bc.callTransfer(stateTree, sender, tx.To, tx.Value); err != nil {
		return nil, err
	}

	stateTree.UpdateAll()

	receipt := &Receipt{
		TxHash:  tx.Hash(),
		Version: tx.Version,
		Status:  uint32(1),
		GasUsed: basicGas,
	}
	return receipt, nil
}

func (bc *BlockChain) callTransfer(st *StateTree, from, to common.Address, amount *big.Int) error {
	return bc.transfer(st, from, to, amount)
}

func (bc *BlockChain) transfer(st *StateTree, from, to common.Address, amount *big.Int) error {
	fromObj := st.GetOrNewStateObj(from)
	toObj := st.GetOrNewStateObj(to)

	if fromObj.balance.Cmp(amount) < 0 {
		return errors.New("from balance is not enough")
	}

	fromObj.SubBalance(amount)
	toObj.AddBalance(amount)
	return nil
}

func (bc *BlockChain) GetBlockHashes(from uint64, count uint64) []common.Hash {
	head := bc.currentBlock.Height()
	if from+count > head {
		count = head
	}
	hashes := make([]common.Hash, 0)
	for h := uint64(0); from+h <= count; h++ {
		block := bc.GetBlockByNumber(from + h)
		hashes = append(hashes, block.Hash())
	}
	return hashes
}

func (bc *BlockChain) GetBlockHashesFromHash(hash common.Hash, max uint64) (chain []common.Hash) {
	block := bc.GetBlockByHash(hash)
	if block == nil {
		return
	}
	// XXX Could be optimised by using a different database which only holds hashes (i.e., linked list)
	for i := uint64(0); i < max; i++ {
		block = bc.GetBlockByHash(block.HashPrevBlock())
		if block == nil {
			break
		}
		chain = append(chain, block.Hash())
	}

	return
}
func (bc *BlockChain) GetBlocks(from uint64, count uint64) []*Block {
	head := bc.currentBlock.Height()
	if from+count > head {
		count = head
	}
	hashes := make([]*Block, 0)
	for h := uint64(0); from+h <= count; h++ {
		block := bc.GetBlockByNumber(from + h)
		if block == nil {
			break
		}
		hashes = append(hashes, block)
	}
	return hashes
}

// FindAncestor tries to locate the common ancestor block of the local chain and
// a remote peers blockchain. In the general case when our node was in sync and
// on the correct chain, checking the top N blocks should already get us a match.
// In the rare scenario when we ended up on a long soft fork (i.e. none of the
// head blocks match), we do a binary search to find the common ancestor.

func (bc *BlockChain) FindAncestor(current *Block, height uint64) *Block {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	return bc.findAncestor(current, height)
}

func (bc *BlockChain) findAncestor(current *Block, height uint64) *Block {
	if current == nil {
		return nil
	}
	if height < 0 || height > current.Height() {
		return nil
	}
	n := current
	for ; n != nil && n.Height() != height; n = bc.GetBlockByHash(n.HashPrevBlock()) {
		// Intentionally left blank
	}
	return n
}

// calcNextRequiredDifficulty calculates the Next Difficulty which will be compared to the
// hash value computed by previou block's hash,transactions hash,timestamp and nonce.
// IF the hash value calculated less than the diffucylty. you have the right to pack a new block
// and brodcast it to the blockchain. the hash value also will be stores in the block header by ziped to 32bit.
func (bc *BlockChain) calcNextRequiredDifficulty(lastBlock *Block) (uint32, error) {
	if lastBlock == nil {
		return 0, nil
	}
	lastHeight := lastBlock.Height()
	blocksPerRetarget := uint64(targetTimespan / targetTimePerBlock)
	// if the height of the next block is not an integral multiple of the target，no changes.
	if (lastHeight+1)%blocksPerRetarget != 0 {
		return lastBlock.Bits(), nil
	}
	// otherwise, find the recent target value.
	first := bc.findAncestor(lastBlock, blocksPerRetarget-1)
	if first == nil {
		return 0, fmt.Errorf("not found ancestor block by height: %d\n", blocksPerRetarget-1)
	}
	firstTime := first.Timestamp()
	lastTime := lastBlock.Timestamp()

	minRetargetTimespan := targetTimespan / adjustmentFactor
	maxRetargetTimespan := targetTimespan * adjustmentFactor

	//calculate the time difference
	actualTimespan := int64(lastTime - firstTime)
	adjustedTimespan := actualTimespan

	// set the maxmun and minimun value of time difference.
	if actualTimespan < minRetargetTimespan {
		adjustedTimespan = minRetargetTimespan
	} else if actualTimespan > maxRetargetTimespan {
		adjustedTimespan = maxRetargetTimespan
	}

	// algorithm of difficulty adjuestment.
	// currentDifficulty * (adjustedTimespan / targetTimespan)
	oldTarget := BitsUnzip(lastBlock.Bits())
	newTarget := new(big.Int).Mul(oldTarget, big.NewInt(adjustedTimespan))
	//targetTimeSpan := targetTimespan / time.Second
	newTarget.Div(newTarget, big.NewInt(targetTimespan))
	powLimit := BitsUnzip(bc.genesisBlock.Bits())
	// if the new target difficulty is larger than the limitation of genesis block. it need be reset
	if newTarget.Cmp(powLimit) > 0 {
		newTarget.Set(powLimit)
	}
	newTargetBits := BigByZip(newTarget)
	return newTargetBits, nil
}

//calculate the next difficuty for hash value of next block.
func (bc *BlockChain) CalcNextRequiredDifficulty() (uint32, error) {
	bc.mu.Lock()
	difficulty, err := bc.calcNextRequiredDifficulty(bc.currentBlock)
	bc.mu.Unlock()
	return difficulty, err
}

func (bc *BlockChain) CurrentStateTree() *StateTree {
	return bc.stateTree
}
