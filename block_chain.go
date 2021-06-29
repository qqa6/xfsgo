package xblockchain

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"log"
	"sync"
	"xblockchain/storage/badger"
	"xblockchain/uint256"
)

type BlockChain struct {
	lastBlockHash *uint256.UInt256
	db *blockChainDB
	writeLock sync.Mutex
	eventBus *EventBus
}

type GenesisBlockOpts struct {
	Address string
	Time int64
}

func DefaultGenesisBlockOpts() *GenesisBlockOpts {
	return &GenesisBlockOpts{
		Address: "15MG8A1XBryHWjtoe8W6iohdeqSBZy3DXPMNYfPSUJi6b",
		Time: 1623213337,
	}
}

type blockChainDB struct {
	storage *badger.Storage
	writeLock sync.Mutex
}

func newBlockChainDB(storage *badger.Storage) *blockChainDB {
	return &blockChainDB{
		storage: storage,
	}
}


func (bcdb *blockChainDB) getLastBlockHash() *uint256.UInt256 {
	val,err := bcdb.storage.Get("l")
	if err != nil {
		return uint256.NewUInt256Zero()
	}
	return uint256.NewUInt256BS(val)
}

func (bcdb *blockChainDB) getBlockByHash(hash *uint256.UInt256) (*Block,error) {
	key := fmt.Sprintf("b%s",hash.Hex())
	val,err := bcdb.storage.Get(key)
	if err != nil {
		return nil,err
	}
	return ReadBlockOfBytes(val)
}


func (bcdb *blockChainDB) pushBlock(block *Block) error {
	bcdb.writeLock.Lock()
	defer bcdb.writeLock.Unlock()
	hash := block.Hash
	key := fmt.Sprintf("b%s",hash.Hex())
	sb, err := block.Serialize()
	if err != nil {
		return err
	}
	err = bcdb.storage.Set(key, sb)
	if err != nil {
		return err
	}
	return bcdb.setLastBlockHash(hash)
}

func (bcdb *blockChainDB) setLastBlockHash(hash *uint256.UInt256) error {
	return bcdb.storage.Set("l",hash.ToBytes())
}



func NewBlockChain(opts *GenesisBlockOpts,storage *badger.Storage, bus *EventBus) (*BlockChain,error) {
	db:= newBlockChainDB(storage)
	lastBlockHash := db.getLastBlockHash()
	var genesisHash *uint256.UInt256 = nil
	isGenesisInit := false
	if lastBlockHash.Equals(uint256.NewUInt256Zero()) {
		tx,err := NewCoinBaseTransaction(opts.Address,opts.Address)
		if err != nil {
			return nil, err
		}
		genesis := NewGenesisBlock(tx,opts.Time)
		genesisHash = genesis.Hash
		err = db.pushBlock(genesis)
		if err != nil {
			return nil, err
		}
		isGenesisInit = true
	}
	lastBlockHash = db.getLastBlockHash()
	if isGenesisInit && !lastBlockHash.Equals(genesisHash) {
		return nil, fmt.Errorf("create genesis block faild")
	}
	return &BlockChain {
		lastBlockHash: lastBlockHash,
		db: db,
		eventBus: bus,
	},nil
}


func (blockChain *BlockChain) AddBlock(txs []*Transaction) (*Block,error) {
	blockChain.writeLock.Lock()
	defer blockChain.writeLock.Unlock()
	for _, tx := range txs {
		if !blockChain.VerifyTransaction(tx) {
			log.Panic("ERROR: Invalid transaction")
		}
	}
	lastBlockHash := blockChain.lastBlockHash
	lastBlock,err := blockChain.db.getBlockByHash(lastBlockHash)
	if err != nil {
		return nil,err
	}
	block := NewBlock(lastBlock,txs)
	err = blockChain.db.pushBlock(block)
	if err != nil {
		return nil,err
	}
	blockChain.eventBus.Publish(ChainHeadEvent{block})
	blockChain.lastBlockHash = block.Hash
	return block,nil
}

func (blockChain *BlockChain) GetBlockByHash(hash *uint256.UInt256) (*Block,error) {
	return blockChain.db.getBlockByHash(hash)
}

func (blockChain *BlockChain) GetBlockById(id uint64) (*Block,error) {

	lastblock,err := blockChain.db.getBlockByHash(blockChain.lastBlockHash)

	if err != nil {
		return nil, err
	}
	if lastblock.Height == id {
		return lastblock,nil
	}
	if id > lastblock.Height {
		return nil, fmt.Errorf("id overflow head block")
	}
	iterator := blockChain.Iterator()
	if !iterator.HasNext() {
		return nil,fmt.Errorf("id overflow head block")
	}
	for iterator.HasNext() {
		block,err := iterator.Next()
		if err != nil {
			return nil, err
		}
		if block != nil && block.Height == id {
			return block,nil
		}
	}
	return nil,fmt.Errorf("not found")
}

func (blockChain *BlockChain) GetHeadBlock() (*Block,error) {
	headHash := blockChain.GetLastBlockHash()
	return blockChain.GetBlockByHash(headHash)
}

func (blockChain *BlockChain) GetHead() (uint64,error) {
	block, err := blockChain.GetHeadBlock()
	if err != nil {
		return 0, err
	}
	return block.Height, nil
}

func (blockChain *BlockChain) GetLastBlockHash() *uint256.UInt256 {
	return blockChain.db.getLastBlockHash()
}

type BlockchainIterator struct {
	currentHash *uint256.UInt256
	db *blockChainDB
}

func (blockChain *BlockChain) Iterator() *BlockchainIterator {
	return &BlockchainIterator{
		currentHash: blockChain.lastBlockHash,
		db: blockChain.db,
	}
}

func (iterator *BlockchainIterator) Top() (*Block,error) {
	lastHash := iterator.db.getLastBlockHash()
	block,err := iterator.db.getBlockByHash(lastHash)
	if err != nil {
		return nil,err
	}
	iterator.currentHash = lastHash
	return block,nil
}
func (iterator *BlockchainIterator) Next() (*Block,error) {
	block,err := iterator.db.getBlockByHash(iterator.currentHash)
	if err != nil {
		return nil,err
	}
	iterator.currentHash = block.HashPrevBlock
	return block,nil
}

func (iterator *BlockchainIterator) HasNext() bool {
	return !iterator.currentHash.Equals(uint256.NewUInt256Zero())
}

func (blockChain *BlockChain) FindUnspentTransactions(pubKeyHash []byte) ([]*Transaction,error) {
	unspentTXs := make([]*Transaction, 0)
	iterator := blockChain.Iterator()
	spentTXOs := make(map[string][]int)


	if !iterator.HasNext() {
		return unspentTXs,nil
	}
	for iterator.HasNext() {
		block,err := iterator.Next()
		if err != nil {
			return nil, err
		}
		txs := block.Transactions
		for i:=0; i<len(txs); i++ {
			tx := txs[i]
			txIDHex := tx.ID.HexstrFull()
			outputs := tx.Outputs
			OutputStep:
			for j:=0; j<len(outputs); j++ {
				output := outputs[j]
				if spentTXOs[txIDHex] != nil {
					for k:=0; k< len(spentTXOs[txIDHex]); k++ {
						spentOutIndex := spentTXOs[txIDHex][k]
						if j == spentOutIndex {
							continue OutputStep
						}
					}
				}
				if output.IsLockedWithKey(pubKeyHash){
					unspentTXs = append(unspentTXs,tx)
				}
			}
			if !tx.IsCoinBase() {
				inputs := tx.Inputs
				for j:=0; j<len(inputs); j++ {
					input := inputs[j]
					if input.UsesKey(pubKeyHash) {
						inTxID := input.TxID.Hex()
						spentTXOs[inTxID] = append(spentTXOs[inTxID], input.Output )
					}
				}
			}
		}
	}
	return unspentTXs,nil
}

func (blockChain *BlockChain) FindUTXO(pubKeyHash []byte) ([]*TransactionOutput,error) {
	var UTXOs []*TransactionOutput = nil
	unspentTxs,err := blockChain.FindUnspentTransactions(pubKeyHash)
	if err != nil {
		return nil, err
	}
	for i:=0; i<len(unspentTxs); i++ {
		tx := unspentTxs[i]
		outs := tx.Outputs
		for j := 0; j < len(tx.Outputs); j++ {
			output := outs[j]
			if output.IsLockedWithKey(pubKeyHash) {
				UTXOs = append(UTXOs, output)
			}
		}
	}
	return UTXOs,err
}
func (blockChain *BlockChain) GetBlockHashes(from int, count int) []uint256.UInt256 {
	if from < 0 || count <= 0 {
		return nil
	}
	tmp := make([]uint256.UInt256, 0)
	for i := from; i < from + count; i++ {
		block, err := blockChain.GetBlockById(uint64(i))
		if err != nil {
			break
		}
		tmp = append(tmp, *block.Hash)
	}
	return tmp
}


func (blockChain *BlockChain) InsertBatchBlock(blocks []*Block) (int, error){
	index := 0
	var err error = nil
	for i, block := range blocks {
		index = i
		if err = blockChain.InsertBlock(block); err != nil{
			break
		}
	}
	return index, err
}


func (blockChain *BlockChain) InsertBlock(block *Block) error {
	blockChain.writeLock.Lock()
	defer blockChain.writeLock.Unlock()
	logrus.Infof("block insert to db by hash: %s", block.Hash.Hex())
	if block == nil {
		return errors.New("empty block")
	}
	if old, _ := blockChain.GetBlockByHash(block.Hash); old != nil {
		logrus.Infof("block is exists hash: %s", block.Hash.Hex())
		return nil
	}
	pow := NewProofOfWork(block)
	if !pow.Validate() {
		return errors.New("pow validate err")
	}
	txs := block.Transactions
	for _, tx := range txs {
		if !blockChain.VerifyTransaction(tx) {
			return errors.New("verify transaction err")
		}
	}
	if err := blockChain.db.pushBlock(block); err != nil {
		logrus.Warnf("push block to db err: %s", err)
		return err
	}
	logrus.Infof("save db by hash: %s success", block.Hash.Hex())
	blockChain.eventBus.Publish(ChainHeadEvent{block})
	blockChain.lastBlockHash = block.Hash
	return nil
}

func (blockChain *BlockChain) FindTransaction(id *uint256.UInt256) (*Transaction,error) {
	iterator := blockChain.Iterator()
	if !iterator.HasNext() {
		return nil, fmt.Errorf("not found")
	}
	for iterator.HasNext() {
		block,err := iterator.Next()
		if err != nil {
			return nil, err
		}
		txs := block.Transactions
		for _, tx := range txs {
			if tx.ID.Equals(id) {
				return tx,err
			}
		}
	}
	return nil, fmt.Errorf("not found")
}

func (blockChain *BlockChain) FindSpendableOutputs(pubKeyHash []byte, amount uint64) (uint64, map[string][]uint64,error) {
	unspentOutputs := make(map[string][]uint64)
	txs,err := blockChain.FindUnspentTransactions(pubKeyHash)
	if err != nil {
		return 0, nil, err
	}
	accumulated := uint64(0)
	loopTag:
	for i:=0; i<len(txs); i++ {
		tx := txs[i]
		outputs := tx.Outputs
		for j := 0; j<len(outputs); j++ {
			out := outputs[j]
			if out.IsLockedWithKey(pubKeyHash) && accumulated < amount {

				accumulated += out.Value
				unspentOutputs[tx.ID.Hex()] = append(unspentOutputs[tx.ID.Hex()], out.Value)
				if accumulated >= amount {
					break loopTag
				}
			}
		}
	}
	return accumulated, unspentOutputs, nil
}

func (blockChain *BlockChain) SignTransaction(tx *Transaction, key []byte) error {
	quoteTxs := make(map[string]*Transaction)
	for _, vin := range tx.Inputs {
		quoteTx,err := blockChain.FindTransaction(vin.TxID)
		if err != nil {
			return err
		}
		quoteTxs[quoteTx.ID.Hex()] = quoteTx
	}
	return tx.Sign(key, quoteTxs)
}

func (blockChain *BlockChain) VerifyTransaction(tx *Transaction) bool  {
	if tx.IsCoinBase() {
		return true
	}
	quoteTxs := make(map[string]*Transaction)
	for _, vin := range tx.Inputs {
		quoteTx,err := blockChain.FindTransaction(vin.TxID)
		if err != nil {
			return false
		}
		if quoteTx == nil {
			return false
		}
		quoteTxs[quoteTx.ID.Hex()] = quoteTx
	}
	err := tx.Verify(quoteTxs)
	if err != nil {
		fmt.Println(err)
		return false
	}
	return true
}

func (blockChain *BlockChain) GetBalanceOfAddress(address string) (uint64,error) {
	pubKeyHash := ParsePubKeyHashByAddress(address)
	utxos,err := blockChain.FindUTXO(pubKeyHash)
	if err != nil {
		return 0, err
	}
	balance := uint64(0)
	for _, out := range utxos {
		balance += out.Value
	}
	return balance,nil
}


