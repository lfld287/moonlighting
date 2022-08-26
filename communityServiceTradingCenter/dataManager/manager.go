package dataManager

import (
	"bytes"
	"encoding/gob"
	"github.com/dgraph-io/badger/v3"
	"go.uber.org/zap"
	"moonlighting/common/database/badgerManager"
	"moonlighting/common/logger"
	"sort"
	"sync"
)

type Data struct {
	//SerialNo          string  `json:"serialNo"`
	//Theme             string  `json:"theme"`
	//ProjectCycle      string  `json:"projectCycle"`
	//Qualification     string  `json:"qualification"`
	//PublishTimeMs     uint64  `json:"publishTimeMs"`
	//Content           string  `json:"content"`
	//Amount            float64 `json:"amount"`
	//PurchaseLink      string  `json:"purchaseLink"`
	//SupportHotline    string  `json:"supportHotline"`
	//DeclareDeadlineMs uint64  `json:"declareDeadlineMs"`
	//InsertTimeMs      uint64  `json:"updateTimeMs"`
	Value        map[string]string `json:"value"`
	InsertTimeMs uint64            `json:"insertTimeMs"`
}

func init() {
	gob.Register(Data{})
}

func serializeData(data Data) []byte {
	tmp := bytes.NewBuffer(nil)
	encoder := gob.NewEncoder(tmp)
	err := encoder.Encode(data)
	if err != nil {
		panic(err)
	}
	return tmp.Bytes()
}

func deSerializeData(buffer []byte) Data {
	tmp := bytes.NewBuffer(buffer)
	encoder := gob.NewDecoder(tmp)
	var data Data
	err := encoder.Decode(&data)
	if err != nil {
		panic(err)
	}
	return data
}

type Manager struct {
	logger                  logger.ILogger
	dbPath                  string
	dbManager               *badgerManager.Manager
	updateSortKeyListSignal chan int
	sortKeyListLock         sync.RWMutex
	sortKeyList             []string
	stopSignal              chan int
	stopOnce                sync.Once
}

func (p *Manager) Start() {
	go p.dbManager.Start()
	go p.loopMain()
	//init sorted key list
	p.updateSortKeyListSignal <- 1
	<-p.stopSignal
}

func (p *Manager) Stop() {
	p.stopOnce.Do(func() {
		select {
		case <-p.stopSignal:
			return
		default:

		}
		p.dbManager.Stop()
		close(p.stopSignal)
	})
}

func (p *Manager) loopMain() {
	for true {
		select {
		case <-p.stopSignal:
			{
				return
			}
		case <-p.updateSortKeyListSignal:
			{
				//clean channel before update
				waitForChanEmpty := func() {
					for true {
						select {
						case <-p.updateSortKeyListSignal:
							{
								continue
							}
						default:
							return
						}
					}
				}
				waitForChanEmpty()
				p.updateSortKeyList()
			}
		}
	}
}

func (p *Manager) updateSortKeyList() {
	type kuPair struct {
		key          string
		insertTimeMs uint64
	}
	kuPairList := make([]kuPair, 0)
	err := p.dbManager.IterateData(func(key []byte, value []byte) {
		kStr := string(key)
		data := deSerializeData(value)
		kuPairList = append(kuPairList, kuPair{
			key:          kStr,
			insertTimeMs: data.InsertTimeMs,
		})
	})
	if err != nil {
		p.logger.Error("updateSortKeyList failed!", zap.Error(err))
		return
	}
	sort.SliceStable(kuPairList, func(i, j int) bool {
		return kuPairList[i].insertTimeMs > kuPairList[j].insertTimeMs
	})

	tmp := make([]string, 0)
	for i := 0; i < len(kuPairList); i++ {
		tmp = append(tmp, kuPairList[i].key)
	}
	p.sortKeyListLock.Lock()
	defer p.sortKeyListLock.Unlock()
	p.sortKeyList = tmp
}

func (p *Manager) getSortKeyList() []string {
	p.sortKeyListLock.RLock()
	defer p.sortKeyListLock.RUnlock()
	sklCopy := p.sortKeyList
	return sklCopy
}

func (p *Manager) InsertData(m map[string]Data) (err error) {
	dataList := make([]badgerManager.DataSet, 0)
	for k, v := range m {
		dataList = append(dataList, badgerManager.DataSet{
			Key:   []byte(k),
			Value: serializeData(v),
		})
	}
	defer func() {
		if err == nil {
			p.updateSortKeyListSignal <- 1
		}
	}()
	return p.dbManager.InsertData(dataList)
}

func (p *Manager) DeleteData(k []string) (err error) {
	keyList := make([][]byte, len(k))
	for i := 0; i < len(keyList); i++ {
		keyList[i] = []byte(k[i])
	}
	defer func() {
		if err == nil {
			p.updateSortKeyListSignal <- 1
		}
	}()
	return p.dbManager.DeleteData(keyList)
}

func (p *Manager) QueryData(limit int, page int, matchRule map[string]string) (map[string]Data, int, error) {
	skip := 0
	if limit > 0 && page > 0 {
		skip = limit * (page - 1)
	}
	count := 0
	totalCount := 0
	err := p.dbManager.ViewData(func(txn *badger.Txn) error {
		for _, key := range p.getSortKeyList() {
			item, err := txn.Get([]byte(key))
			if err != nil {
				return err
			}
			if item == nil {
				continue
			}

			if matchRule != nil && len(matchRule) > 0 {
				err = item.Value(func(val []byte) error {
					data := deSerializeData(val)
					for k, v := range matchRule {
						fieldVal, _ := data.Value[k]
						regexp.
					}
					return nil
				})
				if err != nil {
					return err
				}
			}

		}
		return nil
	})
	if err != nil {
		return nil, 0, err
	}
}
