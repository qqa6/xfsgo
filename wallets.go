package xblockchain

import (
	"fmt"
	"xblockchain/storage/badger"
)

type Wallets struct {
	db *keystoreDB
	defaultAddr string
}

type keystoreDB struct {
	storage *badger.Storage
}

func newKeystoreDB(storage *badger.Storage) *keystoreDB {
	return &keystoreDB{
		storage: storage,
	}
}
func (keystoreDB *keystoreDB) GetDefaultAddress() (string,error) {
	data, err := keystoreDB.storage.Get("d")
	if err != nil {
		return "", err
	}
	if data == nil {
		return "", fmt.Errorf("not found data")
	}
	return string(data), nil
}

func (keystoreDB *keystoreDB) ListPrivateKeys() ([][]byte,error) {
	var tmp [][]byte = nil
	tmp = make([][]byte, 0)
	err := keystoreDB.storage.PrefixForeach("k", func(k string, v []byte) error {
		tmp = append(tmp, v)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return tmp, nil
}


func (keystoreDB *keystoreDB) GetPrivateKeyByAddress(address string) ([]byte,error) {
	key := fmt.Sprintf("k%s",address)
	return keystoreDB.storage.Get(key)
}

func (keystoreDB *keystoreDB) PutPrivateKey(addr string,privateKey []byte) error {
	key := fmt.Sprintf("k%s",addr)
	return keystoreDB.storage.Set(key, privateKey)
}



func (keystoreDB *keystoreDB) SetDefaultAddress(address string) error {
	return keystoreDB.storage.Set("d", []byte(address))
}


func (keystoreDB *keystoreDB) DelKeyByAddress(address string) error {
	key := fmt.Sprintf("k%s",address)
	privateKey, err := keystoreDB.storage.Get(key)
	if err != nil {
		return err
	}
	if privateKey == nil {
		return fmt.Errorf("not found address")
	}
	return keystoreDB.storage.Del(key)
}

func NewWallets(storage *badger.Storage) *Wallets {
	db := newKeystoreDB(storage)
	defaultAddr, _ := db.GetDefaultAddress()
	return &Wallets{
		db: db,
		defaultAddr: defaultAddr,
	}
}

func (ws *Wallets) NewWallet() (string,error) {
	w := NewRandomWallet()
	return ws.AddWallet(w)
}

func (ws *Wallets) AddWallet(wallet *Wallet) (string,error) {
	addr := wallet.GetAddress()
	privateKey, _ := ws.db.GetPrivateKeyByAddress(addr)
	if privateKey != nil {
		return "", fmt.Errorf("address already exists")
	}
	err := ws.db.PutPrivateKey(addr, wallet.privateKey)
	if err != nil {
		return "", err
	}
	if ws.defaultAddr == "" {
		err := ws.db.SetDefaultAddress(addr)
		if err != nil {
			return "", err
		}
		ws.defaultAddr = addr
	}
	return addr,nil
}


func (ws *Wallets) List() []*Wallet  {
	wallets := make([]*Wallet, 0)
	keys, err := ws.db.ListPrivateKeys()
	if err != nil {
		return wallets
	}
	for i:=0;i<len(keys);i++ {
		key := keys[i]
		if key != nil {
			w := ImportWalletWithKey(key)
			wallets = append(wallets, w)
		}
	}
	return wallets
}

func (ws *Wallets) GetWalletByAddress(address string) (*Wallet,error)  {
	key, err := ws.db.GetPrivateKeyByAddress(address)
	if err != nil {
		return nil, err
	}
	if key == nil {
		return nil, fmt.Errorf("data not found")
	}
	return ImportWalletWithKey(key),nil
}



func (ws *Wallets) SetDefault(address string) error {
	if address == ws.defaultAddr {
		return nil
	}
	err := ws.db.SetDefaultAddress(address)
	if err != nil {
		return err
	}
	err = ws.db.DelKeyByAddress(ws.defaultAddr)
	if err != nil {
		return err
	}
	ws.defaultAddr = address
	return nil
}

func (ws *Wallets) GetDefault() string {
	return ws.defaultAddr
}

func (ws *Wallets) Delete(address string) error {
	if address == ws.defaultAddr {
		err := ws.db.SetDefaultAddress("")
		if err != nil{
			return err
		}
	}
	return ws.db.DelKeyByAddress(address)
}

func (ws *Wallets) Export(address string) (string,error) {
	w,err := ws.GetWalletByAddress(address)
	if err != nil{
		return "",err
	}
	return w.Export(),nil
}

func (ws *Wallets) Import(enc string) (string,error) {
	w,err := ImportWalletWithB64(enc)
	if err != nil{
		return "",err
	}
	return ws.AddWallet(w)
}
