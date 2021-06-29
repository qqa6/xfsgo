package xblockchain

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"github.com/btcsuite/btcutil/base58"
	"log"
	"xblockchain/uint256"
	ecdsa2 "xblockchain/util/crypto/ecdsa"
)

type Transaction struct {
	ID *uint256.UInt256 `json:"id"`
	Inputs []*TransactionInput `json:"inputs"`
	Outputs []*TransactionOutput `json:"outputs"`
}

type TransactionInput struct {
	TxID *uint256.UInt256 `json:"id"`
	Output int `json:"output"`
	ScriptSig []byte `json:"script_sig"`

}
type TransactionOutput struct {
	Value uint64 `json:"value"`
	ScriptPubKey []byte `json:"script_pub_key"`
	PubKeyHash []byte `json:"pub_key_hash"`
}

func NewTransactionOutput(value uint64, address string) (*TransactionOutput,error) {
	txout := &TransactionOutput{Value: value}
	err := txout.Lock(address)
	if err != nil{
		return nil,err
	}
	return txout,nil
}

func (input *TransactionInput) UsesKey(pubKeyHash []byte) bool {
	scriptPayload := fmt.Sprintf(`%s;up();pkh();push("%s");eq()`,input.ScriptSig, pubKeyHash)
	scriptbin, err := BuildScriptBin(scriptPayload)
	if err != nil {
		return false
	}
	vm := NewScriptVM()
	err = vm.ExecScriptBin(scriptbin)
	if err != nil {
		return false
	}
	if vm.RunningStackIsEmpty() {
		return false
	}
	return true
}

func (output *TransactionOutput) Lock(address string) error {
	pkhbs := ParsePubKeyHashByAddress(address)
	pkhEnc := base58.Encode(pkhbs)
	script := fmt.Sprintf(`up();pkh();push("%s");eq();checkSig()`,pkhEnc)
	scriptbin, err := BuildScriptBin(script)
	if err != nil {
		return err
	}
	output.ScriptPubKey = scriptbin
	output.PubKeyHash = pkhbs
	return nil
}

func (output *TransactionOutput) IsLockedWithKey(pubKeyHash []byte) bool  {
	return bytes.Compare(pubKeyHash,output.PubKeyHash) == 0
}

func NewCoinBaseTransaction(to, data string) (*Transaction,error) {
	if data == "" {
		randData := make([]byte, 20)
		_, err := rand.Read(randData)
		if err != nil {
			log.Panic(err)
		}
		data = fmt.Sprintf("%x", randData)
	}
	txin := &TransactionInput{
		TxID: uint256.NewUInt256Zero(),
		Output: -1,
	}
	txout,err := NewTransactionOutput(10,to)
	if err != nil {
		return nil, err
	}
	tx := &Transaction{
		ID: uint256.NewUInt256Zero(),
		Inputs: []*TransactionInput{txin},
		Outputs: []*TransactionOutput{txout},
	}
	tx.ID = uint256.NewUInt256BS(tx.Hash())
	return tx,nil
}

func NewUTXOTransaction(from, to string, amount uint64, bc *BlockChain, privateKey []byte) (*Transaction,error) {
	var inputs []*TransactionInput = nil
	var outputs []*TransactionOutput = nil
	pubKeyHashFrom := ParsePubKeyHashByAddress(from)
	acc, unspentOutputs,err := bc.FindSpendableOutputs(pubKeyHashFrom, amount)
	if err != nil {
		return nil, err
	}
	if acc < amount {
		return nil, fmt.Errorf("not enough funds")
	}
	for txIDHex, outs := range unspentOutputs {
		txID := uint256.NewUInt256(txIDHex)
		for out := range outs {
			input := &TransactionInput{
				TxID: txID,
				Output: out,
			}
			inputs = append(inputs, input)
		}
	}
	output, err := NewTransactionOutput(amount, to)
	if err != nil {
		return nil, err
	}
	outputs = append(outputs, output)
	if acc > amount {
		output, err = NewTransactionOutput(acc - amount, from)
		if err != nil {
			return nil, err
		}
		outputs = append(outputs, output)
	}
	tx := &Transaction{
		ID: nil,
		Inputs: inputs,
		Outputs: outputs,
	}
	tx.SetId()
	err = bc.SignTransaction(tx,privateKey)
	return tx,err
}
func (tx *Transaction) SetId() {
	encoded := bytes.Buffer{}
	encoder := gob.NewEncoder(&encoded)
	err := encoder.Encode(tx)
	if err != nil {}
	hash := sha256.Sum256(encoded.Bytes())
	tx.ID = uint256.NewUInt256BS(hash[:])
}

func (tx *Transaction) Hash() []byte {
	encoded := bytes.Buffer{}
	encoder := gob.NewEncoder(&encoded)
	err := encoder.Encode(tx)
	if err != nil {}
	hash := sha256.Sum256(encoded.Bytes())
	return hash[:]
}

func (tx *Transaction) IsCoinBase() bool {
	return len(tx.Inputs) == 1 && tx.Inputs[0].TxID.IsZero() && tx.Inputs[0].Output == -1
}


func (tx *Transaction) Sign(key []byte, quoteTxs map[string]*Transaction) error {
	if tx.IsCoinBase() {
		return nil
	}
	privateKey,err := x509.ParseECPrivateKey(key)
	if err != nil { return err }
	txCopy := tx.TrimmedCopy()
	for inID, vin := range txCopy.Inputs {
		quoteTx := quoteTxs[vin.TxID.Hex()]
		vout := quoteTx.Outputs[vin.Output]
		pubKey := ecdsa2.ParsePubKeyWithPrivateKey(key)
		pubKeyHash := PubKeyHash(pubKey)
		if !vout.IsLockedWithKey(pubKeyHash) {
			return fmt.Errorf("sign transaction err, output cannot be referenced")
		}
		pubKeyEnc := base58.Encode(pubKey)
		data := fmt.Sprintf("%v\n",txCopy)
		signature, err := privateKey.Sign(rand.Reader,[]byte(data),nil)
		if err != nil { return err }
		signatureEnc := base58.Encode(signature)
		scriptPayload := fmt.Sprintf(`push("%s");push("%s")`,signatureEnc, pubKeyEnc)
		scriptbin, err := BuildScriptBin(scriptPayload)
		if err != nil {
			return err
		}
		tx.Inputs[inID].ScriptSig = scriptbin
	}
	return nil
}

func (tx *Transaction) Verify(quoteTxs map[string]*Transaction) error {
	if tx.IsCoinBase() {
		return nil
	}
	txCopy := tx.TrimmedCopy()
	for vid, vin := range txCopy.Inputs {
		quoteTx := quoteTxs[vin.TxID.Hex()]
		vout := quoteTx.Outputs[vin.Output]
		vc := tx.Inputs[vid]
		data := fmt.Sprintf("%v\n",txCopy)

		scriptSigDec,err := ParseScriptBin2Str(vc.ScriptSig)
		if err != nil {
			return err
		}
		scriptPubKeyDec,err := ParseScriptBin2Str(vout.ScriptPubKey)
		if err != nil {
			return err
		}
		scriptPayload := fmt.Sprintf("%s;%s",scriptSigDec,scriptPubKeyDec)
		scriptPayloadBin,err := BuildScriptBin(scriptPayload)
		if err != nil {
			return err
		}
		vm := NewScriptVM()
		err = vm.ExecScriptBinCheckSign(scriptPayloadBin, func(pubkey []byte, sign []byte) bool {
			pkds := base58.Decode(string(pubkey))
			pkGen, err := x509.ParsePKIXPublicKey(pkds)
			if err != nil {
				return false
			}
			pk := pkGen.(*ecdsa.PublicKey)
			signDec := base58.Decode(string(sign))
			if !ecdsa.VerifyASN1(pk, []byte(data), signDec) {
				return  false
			}
			return true
		})
		if err != nil {
			return err
		}
		if vm.RunningStackIsEmpty() {
			return fmt.Errorf("tx dont verify")
		}
		vmstat := vm.PopRunningStack()
		if vmstat != "1" {
			return fmt.Errorf("tx dont verify")
		}
	}
	return nil
}


func (tx *Transaction) TrimmedCopy() *Transaction {
	var inputs []*TransactionInput
	var outputs []*TransactionOutput

	for _, vin := range tx.Inputs {
		inputs = append(inputs,
			&TransactionInput{
			TxID: vin.TxID,
			Output: vin.Output,
			ScriptSig: nil,
		})
	}

	for _, vout := range tx.Outputs {
		outputs = append(outputs,
			&TransactionOutput{
			Value: vout.Value,
			ScriptPubKey: vout.ScriptPubKey,
			PubKeyHash: vout.PubKeyHash,
		})
	}

	txCopy := &Transaction{tx.ID, inputs, outputs}

	return txCopy
}

func (tx *Transaction) String() string {
	jsonByte,err := json.Marshal(tx)
	if err != nil {
		return ""
	}
	return string(jsonByte)
}
