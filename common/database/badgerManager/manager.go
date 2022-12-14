package badgerManager

import (
	"errors"
	"github.com/dgraph-io/badger/v3"
	"go.uber.org/zap"
	"moonlighting/common/logger"
	"sync"
)

type DataSet struct {
	Key   []byte
	Value []byte
}

type IterationFunc func(key []byte, value []byte)

type Manager struct {
	logger     logger.ILogger
	dbPath     string
	internalDB *badger.DB
	stopSignal chan int
	stopOnce   sync.Once
	initOnce   sync.Once
}

func NewBadgerManager(l logger.ILogger, dbPath string) *Manager {
	return &Manager{
		logger:     l,
		dbPath:     dbPath,
		internalDB: nil,
		stopSignal: make(chan int),
		stopOnce:   sync.Once{},
		initOnce:   sync.Once{},
	}
}

func (p *Manager) Start() {
	_ = p.checkDB()
	<-p.stopSignal
}

func (p *Manager) Stop() {
	p.stopOnce.Do(func() {
		select {
		case <-p.stopSignal:
			return
		default:

		}
		close(p.stopSignal)
	})
}

func (p *Manager) checkDB() error {
	if p.internalDB == nil || p.internalDB.IsClosed() {
		return p.openDataBase()
	} else {
		return nil
	}
}

func (p *Manager) openDataBase() error {
	var err = errors.New("open twice")
	p.initOnce.Do(func() {
		p.internalDB, err = badger.Open(badger.DefaultOptions(p.dbPath))
	})
	return err
}

func (p *Manager) InsertData(dList []DataSet) error {
	err := p.checkDB()
	if err != nil {
		return err
	}
	err = p.internalDB.Update(func(txn *badger.Txn) error {
		for _, d := range dList {
			err = txn.Set(d.Key, d.Value)
			if err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

func (p *Manager) LoadData(keyList [][]byte) ([]DataSet, error) {
	err := p.checkDB()
	if err != nil {
		return nil, err
	}
	var res = make([]DataSet, 0)
	err = p.internalDB.View(func(txn *badger.Txn) error {
		for _, key := range keyList {
			item, err := txn.Get(key)
			if err != nil {
				return err
			}
			if item == nil {
				return errors.New("item nil")
			}
			key := item.KeyCopy(nil)
			value, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}
			res = append(res, DataSet{
				Key:   key,
				Value: value,
			})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (p *Manager) DeleteData(keyList [][]byte) error {
	err := p.checkDB()
	if err != nil {
		return err
	}
	err = p.internalDB.Update(func(txn *badger.Txn) error {
		for _, key := range keyList {
			err = txn.Delete(key)
			if err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

func (p *Manager) ViewData(raw func(txn *badger.Txn) error) error {
	err := p.checkDB()
	if err != nil {
		return err
	}
	err = p.internalDB.View(raw)
	return err
}

func (p *Manager) UpdateData(raw func(txn *badger.Txn) error) error {
	err := p.checkDB()
	if err != nil {
		return err
	}
	err = p.internalDB.Update(raw)
	return err
}

/*
It's byte-wise lexicographical sorting. We can probably clarify that in the README.

https://golang.org/pkg/bytes/#Compare
*/

func (p *Manager) IterateData(loadFunc IterationFunc, prefix []byte) error {
	err := p.checkDB()
	if err != nil {
		return err
	}
	err = p.internalDB.View(func(txn *badger.Txn) error {
		opt := badger.DefaultIteratorOptions
		if prefix != nil && len(prefix) > 0 {
			opt.Prefix = prefix
		}
		iter := txn.NewIterator(opt)
		if iter == nil {
			return errors.New("create iter failed")
		}
		defer iter.Close()

		var item *badger.Item = nil
		keyBuf := make([]byte, 0)
		valueBuf := make([]byte, 0)
		for iter.Rewind(); iter.Valid(); iter.Next() {
			item = iter.Item()
			if item == nil {
				continue
			}
			keyBuf = item.KeyCopy(keyBuf)
			valueBuf, err = item.ValueCopy(valueBuf)
			if err != nil {
				p.logger.Error("iteration error", zap.ByteString("key", keyBuf), zap.Error(err))
				continue
			}
			loadFunc(keyBuf, valueBuf)
		}
		return nil
	})
	return err
}
