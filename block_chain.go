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

const (
	// maxOrphanBlocks is the maximum number of orphan blocks that can be
	// queued.
	maxOrphanBlocks = 100
)

// orphanBlock represents a block that we don't yet have the parent for.  It
// is a normal block plus an expiration time to prevent caching the orphan
// forever.
type orphanBlock struct {
	block      *Block
	expiration time.Time
}

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
	eventBus      *EventBus

	// These fields are related to handling of orphan blocks.  They are
	// protected by a combination of the chain lock and the orphan lock.
	orphanLock   sync.RWMutex
	orphans      map[common.Hash]*orphanBlock
	prevOrphans  map[common.Hash][]*orphanBlock
	oldestOrphan *orphanBlock
}

// WriteStatus status of write
type WriteStatus byte

const (
	NonStatTy WriteStatus = iota
	CanonStatTy
	SideStatTy
)

//NewBlockChain creates a initialised block chain using information available in the database.
//this new blockchain includes a stateTree by which the blockchian can manage the whole state of the chainb.
//such as the account's information of every user.
func NewBlockChain(stateDB, chainDB, extraDB *badger.Storage, eventBus *EventBus) (*BlockChain, error) {
	bc := &BlockChain{
		chainDB:     newChainDB(chainDB),
		stateDB:     stateDB,
		extraDB:     newExtraDB(extraDB),
		eventBus:    eventBus,
		orphans:     make(map[common.Hash]*orphanBlock),
		prevOrphans: make(map[common.Hash][]*orphanBlock),
	}
	// bc.orphansCache = lru.NewCache(2048)
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

func (bc *BlockChain) WriteHead(block *Block) error {
	bc.mu.Lock()

	err := bc.chainDB.WriteHead(block)
	if err != nil {
		return err
	}
	bc.currentBlock = block
	bc.lastBlockHash = block.Hash()
	lastStateRoot := block.StateRoot()
	bc.stateTree = NewStateTree(bc.stateDB, lastStateRoot.Bytes())
	bc.mu.Unlock()

	bc.eventBus.Publish(ChainHeadEvent{block})

	return nil
}

func (bc *BlockChain) GetHead() *Block {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	return bc.chainDB.GetHeadBlock()
}
func (bc *BlockChain) GetBalance(addr common.Address) *big.Int {
	gotBalance := bc.stateTree.GetBalance(addr)
	return gotBalance

}

func (bc *BlockChain) removeNumBlock(block *Block) error {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	if err := bc.chainDB.RemoveNumBlock(block); err != nil {
		return fmt.Errorf("remove NumBlock err: %s", err)
	}

	return nil
}

func (bc *BlockChain) writeNumBlock(block *Block) {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	if err := bc.chainDB.WriteNumBlock(block); err != nil {
		logrus.Errorf("write numBlock err: %s", err)
	}
}

// WriteBlock stores the block inputed to the local database.
func (bc *BlockChain) WriteBlock2DB(block *Block) error {

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

// ProcessBlock is the main workhorse for handling insertion of new blocks into
// the block chain.  It includes functionality such as rejecting duplicate
// blocks, ensuring blocks follow all rules, orphan handling, and insertion into
// the block chain along with best chain selection and reorganization.
//
// whether or not the block is on the main chain and the first indicates
// whether or not the block is an orphan and the second value indicates.
// When no errors occurred during processing, the third return value indicates
// This function is safe for concurrent access.

func (bc *BlockChain) InsertChain(block *Block) (bool, error) {
	bc.chainmu.Lock()
	defer bc.chainmu.Unlock()

	blockHash := block.Hash()
	logrus.Debugf("InsertChain--->Processing height:%d,block %s", block.Height(), blockHash.Hex())

	// The block must not already exist in the main chain or side chains.
	if old := bc.GetBlockByHash(blockHash); old != nil {
		return false, fmt.Errorf("syncer(InsertChain)-->already have block %s", blockHash.Hex())
	}

	// The block must not already exist as an orphan.
	if _, exists := bc.orphans[blockHash]; exists {
		return false, fmt.Errorf("syncer(InsertChain)-->already have block (orphan) %s", blockHash.Hex())
	}

	if err := bc.checkBlockHeaderSanity(block.GetHeader(), blockHash); err != nil {
		return false, err
	}
	var parent *Block

	if parent = bc.GetBlockByHash(block.HashPrevBlock()); parent == nil {
		bc.addOrphanBlock(block)
		return true, nil
	}

	parentHash := block.Header.HashPrevBlock
	logrus.Debugf("syncer(InsertChain)-->block height:%d,hash:%s,parentHash:%s", block.Height(), block.HashHex(), parentHash.Hex())

	status, err := bc.writeBlockWithState(block)
	if err != nil {
		return false, err
	}
	switch status {
	case CanonStatTy:
		logrus.Debugf("syncer(InsertChain)-->CanonStatTy--->Inserted new block-->number:%d hash:%s", block.Height(), block.HashHex())
	case SideStatTy:
		logrus.Debugf("syncer(InsertChain)-->SideStatTy--->Inserted forked block-->number:%d hash:%s", block.Height(), block.HashHex())
	default:
		// This in theory is impossible, but lets be nice to our future selves and leave
		// a log, instead of trying to track down blocks imports that don't emit logs.
		logrus.Warn("syncer(InsertChain)-->Inserted block with unknown status number:%d hash:%s", block.Height(), block.HashHex())
	}

	hash := block.Hash()
	err = bc.processOrphans(&hash)
	if err != nil {
		return false, err
	}

	return false, nil
}

// WriteBlockWithState writes the block and all associated state to the database.
func (bc *BlockChain) WriteBlockWithState(block *Block) (status WriteStatus, err error) {
	bc.chainmu.Lock()
	defer bc.chainmu.Unlock()
	if old := bc.GetBlockByHash(block.Hash()); old != nil {
		logrus.Debugf("hash %s exist!", block.HashHex())
		return NonStatTy, fmt.Errorf("already have block %s", block.HashHex())
	}
	return bc.writeBlockWithState(block)
}

func (bc *BlockChain) writeBlockWithState(block *Block) (status WriteStatus, err error) {

	currentBlock := bc.GetHead()

	preBlock := bc.GetBlockByHash(block.HashPrevBlock())
	if preBlock == nil {
		return NonStatTy, nil
	}
	// ptd := preBlock.Td
	// externTd := new(big.Int).Add(block.Td, ptd)
	// localTd := currentBlock.Td

	preHash := block.HashPrevBlock()
	parent := bc.GetBlockByHash(preHash)
	if parent == nil {
		return NonStatTy, fmt.Errorf("previous block %s is unknown", preHash)
	}

	txs := block.Transactions
	txsRoot := block.TransactionRoot()
	rsRoot := block.ReceiptsRoot()

	header := block.GetHeader()

	targetTxsRoot := CalcTxsRootHash(block.Transactions)
	if !bytes.Equal(targetTxsRoot.Bytes(), txsRoot.Bytes()) {
		return NonStatTy, fmt.Errorf("check transaction root err")
	}
	parentStateRoot := parent.StateRoot()
	stateTree := NewStateTree(bc.stateDB, parentStateRoot.Bytes())
	_, rec, err := bc.ApplyTransactions(stateTree, header, txs)
	if err != nil {
		return NonStatTy, err
	}
	block.Receipts = rec
	AccumulateRewards(stateTree, header)
	stateTree.UpdateAll()
	targetRsRoot := CalcReceiptRootHash(rec)
	if !bytes.Equal(rsRoot.Bytes(), targetRsRoot.Bytes()) {
		return NonStatTy, fmt.Errorf("check receipt root err")
	}
	if err = stateTree.Commit(); err != nil {
		return NonStatTy, err
	}

	logrus.Debugf("WriteBlock2DB--->Inserted new block to DB First-->height:%d hash:%s", block.Height(), block.HashHex())
	if err = bc.WriteBlock2DB(block); err != nil {
		return NonStatTy, err
	}

	reorg := block.Header.WorkSum.Cmp(currentBlock.Header.WorkSum) > 0
	if !reorg && block.Header.WorkSum.Cmp(currentBlock.Header.WorkSum) == 0 {
		// Split same-difficulty blocks by number, then preferentially select
		// the block generated by the local miner as the canonical block.
		if block.Height() < currentBlock.Height() {
			reorg = true
		} else if block.Height() == currentBlock.Height() {
			// target is same and height is same reorg true
			reorg = false
		}
	}

	if reorg {
		// Reorganise the chain if the parent is not the head block
		if block.HashPrevBlock() != currentBlock.Hash() {
			newBlockPreHash := block.HashPrevBlock()
			currentPreHash := currentBlock.HashPrevBlock()

			logrus.Debugf("newblocKHash:%s, newBlockPreHash:%s, currentPreHash:%s, currentBlockHash:%s",
				block.HashHex(), newBlockPreHash.Hex(), currentBlock.HashHex(), currentPreHash.Hex())

			if err := bc.reorg(currentBlock, block); err != nil {
				return NonStatTy, err
			}
		}
		status = CanonStatTy
	} else {
		status = SideStatTy
	}

	if status == CanonStatTy {

		err := bc.WriteHead(block)
		if err != nil {
			return NonStatTy, fmt.Errorf("WriteHead %s err", block.HashHex())
		}

		logrus.Debugf("writeHeadBlock--->Inserted new block to Chain Second-->height:%d,hash:%s", block.Height(), block.HashHex())
	}

	return status, nil
}

func (bc *BlockChain) reorg(oldBlock, newBlock *Block) error {
	logrus.Debugf("start reorg\n")
	var (
		newChain    []*Block
		oldChain    []*Block
		commonBlock *Block
	)
	// Reduce the longer chain to the same number as the shorter one
	if oldBlock.Height() > newBlock.Height() {
		// Old chain is longer, gather all transactions and logs as deleted ones
		for ; oldBlock != nil && oldBlock.Height() != newBlock.Height(); oldBlock = bc.GetBlockByHash(oldBlock.HashPrevBlock()) {
			oldChain = append(oldChain, oldBlock)
			// deletedTxs = append(deletedTxs, oldBlock.Transactions()...)
			// collectLogs(oldBlock.Hash(), true)
		}
	} else {
		// New chain is longer, stash all blocks away for subsequent insertion
		for ; newBlock != nil && newBlock.Height() != oldBlock.Height(); newBlock = bc.GetBlockByHash(newBlock.HashPrevBlock()) {
			newChain = append(newChain, newBlock)
		}
	}
	if oldBlock == nil {
		return fmt.Errorf("invalid old chain")
	}
	if newBlock == nil {
		return fmt.Errorf("invalid new chain")
	}
	// Both sides of the reorg are at the same number, reduce both until the common
	// ancestor is found
	for {
		// If the common ancestor was found, bail out
		if oldBlock.Hash() == newBlock.Hash() {
			commonBlock = oldBlock
			oldChain = append(oldChain, oldBlock)
			newChain = append(newChain, newBlock)
			break
		}
		// Remove an old block as well as stash away a new block
		oldChain = append(oldChain, oldBlock)
		// deletedTxs = append(deletedTxs, oldBlock.Transactions()...)
		// collectLogs(oldBlock.Hash(), true)

		newChain = append(newChain, newBlock)

		// Step back with both chains
		oldBlock = bc.GetBlockByHash(oldBlock.HashPrevBlock())
		if oldBlock == nil {
			return fmt.Errorf("invalid old chain")
		}
		newBlock = bc.GetBlockByHash(newBlock.HashPrevBlock())
		if newBlock == nil {
			return fmt.Errorf("invalid new chain")
		}
	}
	// Ensure the user sees large reorgs
	if len(oldChain) > 0 && len(newChain) > 0 {
		logrus.Debugf("commonBlock.height:%d, commonBlock.hash:%s\nlen(oldChain):%d, oldchainhash:%s\nlen(newChain):%d, newchainhash:%s",
			commonBlock.Height(), commonBlock.HashHex(), len(oldChain), oldChain[0].HashHex(), len(newChain), newChain[0].HashHex())
	}

	for i := len(newChain) - 1; i >= 1; i-- {
		// Insert the block in the canonical way, re-writing history
		bc.WriteHead(newChain[i])
		logrus.Debugf("insert height:%d,hash:%s\n", newChain[i].Height(), newChain[i].HashHex())
	}

	// Delete any canonical number assignments above the new head
	number := bc.CurrentBlock().Height()
	logrus.Errorf("start remove height:%d and so on from old chain", number)
	for i := number + 1; ; i++ {
		block := bc.getBlockByNumber(i)
		if block == nil {
			break
		}

		bc.removeNumBlock(block)
		logrus.Errorf("del height from height and hash error block height:%d,hash:%s", block.Height(), block.HashHex())
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
		Time:    uint64(time.Now().Unix()),
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

// removeOrphanBlock removes the passed orphan block from the orphan pool and
// previous orphan index.
func (bc *BlockChain) removeOrphanBlock(orphan *orphanBlock) {
	// Protect concurrent access.
	bc.orphanLock.Lock()
	defer bc.orphanLock.Unlock()

	// Remove the orphan block from the orphan pool.
	orphanHash := orphan.block.Hash()
	delete(bc.orphans, orphanHash)

	// Remove the reference from the previous orphan index too.  An indexing
	// for loop is intentionally used over a range here as range does not
	// reevaluate the slice on each iteration nor does it adjust the index
	// for the modified slice.
	prevHash := orphan.block.HashPrevBlock()

	orphans := bc.prevOrphans[prevHash]
	for i := 0; i < len(orphans); i++ {
		hash := orphans[i].block.Hash()
		if hash.IsEqual(&orphanHash) {
			copy(orphans[i:], orphans[i+1:])
			orphans[len(orphans)-1] = nil
			orphans = orphans[:len(orphans)-1]
			i--
		}
	}
	bc.prevOrphans[prevHash] = orphans

	// Remove the map entry altogether if there are no longer any orphans
	// which depend on the parent hash.
	if len(bc.prevOrphans[prevHash]) == 0 {
		delete(bc.prevOrphans, prevHash)
	}
}

// addOrphanBlock adds the passed block (which is already determined to be
// an orphan prior calling this function) to the orphan pool.  It lazily cleans
// up any expired blocks so a separate cleanup poller doesn't need to be run.
// It also imposes a maximum limit on the number of outstanding orphan
// blocks and will remove the oldest received orphan block if the limit is
// exceeded.
func (b *BlockChain) addOrphanBlock(block *Block) {
	logrus.Debugf("addOrphanBlock hash:%s", block.HashHex())
	// Remove expired orphan blocks.
	for _, oBlock := range b.orphans {
		if time.Now().After(oBlock.expiration) {
			b.removeOrphanBlock(oBlock)
			continue
		}

		// Update the oldest orphan block pointer so it can be discarded
		// in case the orphan pool fills up.
		if b.oldestOrphan == nil || oBlock.expiration.Before(b.oldestOrphan.expiration) {
			b.oldestOrphan = oBlock
		}
	}

	// Limit orphan blocks to prevent memory exhaustion.
	if len(b.orphans)+1 > maxOrphanBlocks {
		// Remove the oldest orphan to make room for the new one.
		b.removeOrphanBlock(b.oldestOrphan)
		b.oldestOrphan = nil
	}

	// Protect concurrent access.  This is intentionally done here instead
	// of near the top since removeOrphanBlock does its own locking and
	// the range iterator is not invalidated by removing map entries.
	b.orphanLock.Lock()
	defer b.orphanLock.Unlock()

	// Insert the block into the orphan map with an expiration time
	// 1 hour from now.
	expiration := time.Now().Add(time.Hour)
	oBlock := &orphanBlock{
		block:      block,
		expiration: expiration,
	}
	b.orphans[block.Hash()] = oBlock

	// Add to previous hash lookup index for faster dependency lookups.
	prevHash := block.HashPrevBlock()
	b.prevOrphans[prevHash] = append(b.prevOrphans[prevHash], oBlock)
}

// processOrphans determines if there are any orphans which depend on the passed
// block hash (they are no longer orphans if true) and potentially accepts them.
// It repeats the process for the newly accepted blocks (to detect further
// orphans which may no longer be orphans) until there are no more.
//
// The flags do not modify the behavior of this function directly, however they
// are needed to pass along to maybeAcceptBlock.
//
// This function MUST be called with the chain state lock held (for writes).
func (bc *BlockChain) processOrphans(hash *common.Hash) error {
	// Start with processing at least the passed hash.  Leave a little room
	// for additional orphan blocks that need to be processed without
	// needing to grow the array in the common case.
	processHashes := make([]*common.Hash, 0, 10)
	processHashes = append(processHashes, hash)
	for len(processHashes) > 0 {
		// Pop the first hash to process from the slice.
		processHash := processHashes[0]
		processHashes[0] = nil // Prevent GC leak.
		processHashes = processHashes[1:]

		// Look up all orphans that are parented by the block we just
		// accepted.  This will typically only be one, but it could
		// be multiple if multiple blocks are mined and broadcast
		// around the same time.  The one with the most proof of work
		// will eventually win out.  An indexing for loop is
		// intentionally used over a range here as range does not
		// reevaluate the slice on each iteration nor does it adjust the
		// index for the modified slice.
		for i := 0; i < len(bc.prevOrphans[*processHash]); i++ {
			orphan := bc.prevOrphans[*processHash][i]
			if orphan == nil {
				// log.Warnf("Found a nil entry at index %d in the "+
				// 	"orphan dependency list for block %v", i,
				// 	processHash)
				continue
			}

			// Remove the orphan from the orphan pool.
			orphanHash := orphan.block.Hash()
			bc.removeOrphanBlock(orphan)
			i--

			// Potentially accept the block into the block chain.
			_, err := bc.WriteBlockWithState(orphan.block)
			if err != nil {
				return err
			}

			// Add this block to the list of blocks to process so
			// any orphan blocks that depend on this block are
			// handled too.
			processHashes = append(processHashes, &orphanHash)
		}
	}
	return nil
}
