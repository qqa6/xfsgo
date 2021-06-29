package badger

import (
	"github.com/dgraph-io/badger/v3"
	"io/ioutil"
	"log"
)

type Storage struct {
	db *badger.DB
}

type loggingLevel int

const (
	DEBUG loggingLevel = iota
	INFO
	WARNING
	ERROR
)
type defaultLog struct {
	*log.Logger
	level loggingLevel
}
func (l *defaultLog) Errorf(f string, v ...interface{}) {
	if l.level <= ERROR {
		l.Printf("ERROR: "+f, v...)
	}
}

func (l *defaultLog) Warningf(f string, v ...interface{}) {
	if l.level <= WARNING {
		l.Printf("WARNING: "+f, v...)
	}
}

func (l *defaultLog) Infof(f string, v ...interface{}) {
	if l.level <= INFO {
		l.Printf("INFO: "+f, v...)
	}
}

func (l *defaultLog) Debugf(f string, v ...interface{}) {
	if l.level <= DEBUG {
		l.Printf("DEBUG: "+f, v...)
	}
}



func defaultLogger(level loggingLevel) *defaultLog {
	return &defaultLog{
		Logger: log.New(ioutil.Discard, "badger ", log.LstdFlags),
		level: level,
	}
}

func New(pathname string) *Storage {
	storage := &Storage{}
	opts := badger.DefaultOptions(pathname)
	opts.Logger = defaultLogger(WARNING)
	var err error = nil
	storage.db, err = badger.Open(opts)
	if err != nil {
		panic(err)
	}
	return storage
}

func (storage *Storage) Set(key string, val []byte) error {
	return storage.db.Update(func(txn *badger.Txn) error {
		err := txn.Set([]byte(key),val)
		return err
	})
}

func (storage *Storage) Get(key string) (val []byte, err error) {
	err = storage.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		val,err = item.ValueCopy(val)
		if err != nil {
			return err
		}
		return nil
	})
	return
}


func (storage *Storage) Del(key string) error {
	return storage.db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(key))
	})
}

func (storage *Storage) Close() error {
	return storage.db.Close()
}


func (storage *Storage) Foreach(fn func(k string,v []byte) error ) error {
	return storage.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			key := item.Key()
			val, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}
			err = fn(string(key), val)
			if err != nil {
				return err
			}
		}
		return nil
	})
}


func (storage *Storage) PrefixForeach(prefix string,fn func(k string,v []byte) error ) error {
	return storage.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Seek([]byte(prefix)); it.ValidForPrefix([]byte(prefix)); it.Next() {
			item := it.Item()
			key := item.Key()
			val, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}
			err = fn(string(key), val)
			if err != nil {
				return err
			}
		}
		return nil
	})
}