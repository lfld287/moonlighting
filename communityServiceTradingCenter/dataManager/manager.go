package dataManager

import (
	"bytes"
	"encoding/gob"
	"errors"
	"github.com/dgraph-io/badger/v3"
	"go.uber.org/zap"
	"moonlighting/common/database/badgerManager"
	"moonlighting/common/logger"
	"regexp"
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
	//Priority      uint64  `json:"updateTimeMs"`
	Key      string            `json:"key"`
	Value    map[string]string `json:"value"`
	Priority uint64            `json:"priority"`
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
	prefix                  string
	dbManager               *badgerManager.Manager
	updateSortKeyListSignal chan int
	sortKeyListLock         sync.RWMutex
	sortKeyList             []string
	stopSignal              chan int
	stopOnce                sync.Once
}

func NewDataManager(l logger.ILogger, prefix string, dbManager *badgerManager.Manager) *Manager {
	return &Manager{
		logger:                  l,
		prefix:                  prefix,
		dbManager:               dbManager,
		updateSortKeyListSignal: make(chan int, 50),
		sortKeyListLock:         sync.RWMutex{},
		sortKeyList:             make([]string, 0),
		stopSignal:              make(chan int),
		stopOnce:                sync.Once{},
	}
}

func (p *Manager) Start() {
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
		key      string
		priority uint64
	}
	kuPairList := make([]kuPair, 0)
	err := p.dbManager.IterateData(func(key []byte, value []byte) {
		kStr := string(key)
		data := deSerializeData(value)
		kuPairList = append(kuPairList, kuPair{
			key:      kStr,
			priority: data.Priority,
		})
	}, []byte(p.prefix))
	if err != nil {
		p.logger.Error("updateSortKeyList failed!", zap.Error(err))
		return
	}
	sort.SliceStable(kuPairList, func(i, j int) bool {
		return kuPairList[i].priority > kuPairList[j].priority
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

func (p *Manager) InsertData(list []Data) (err error) {
	dataList := make([]badgerManager.DataSet, 0)
	for _, data := range list {
		if data.Key == "" {
			return errors.New("contains empty key")
		}
		dataList = append(dataList, badgerManager.DataSet{
			Key:   []byte(p.prefix + data.Key),
			Value: serializeData(data),
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
		keyList[i] = []byte(p.prefix + k[i])
	}
	defer func() {
		if err == nil {
			p.updateSortKeyListSignal <- 1
		}
	}()
	return p.dbManager.DeleteData(keyList)
}

func (p *Manager) QueryData(limit int, page int, matchRules []map[string]string) (res []Data, count int, totalCount int, err error) {
	skip := 0
	if limit > 0 && page > 0 {
		skip = limit * (page - 1)
	}
	count = 0
	totalCount = 0
	valueBuffer := make([]byte, 0)
	res = make([]Data, 0)
	err = p.dbManager.ViewData(func(txn *badger.Txn) error {
		kList := p.getSortKeyList()
		p.logger.Info("", zap.Any("kList", kList))
		for _, key := range kList {
			item, err := txn.Get([]byte(key))
			if err != nil {
				if err == badger.ErrKeyNotFound {
					continue
				} else {
					return err
				}
			}
			if item == nil {
				continue
			}

			valueBuffer, err = item.ValueCopy(valueBuffer)
			if err != nil {
				return err
			}
			data := deSerializeData(valueBuffer)
			alreadyMatch := false

			if matchRules != nil && len(matchRules) > 0 {
				for _, matchRule := range matchRules {

					if alreadyMatch {
						break
					}

					if matchRule == nil || len(matchRule) <= 0 {
						continue
					}

					currentRuleMatch := true

					for k, v := range matchRule {
						fieldVal, ok := data.Value[k]
						if !ok {
							currentRuleMatch = false
							break
						}
						matched, err := regexp.MatchString(v, fieldVal)
						if err != nil {
							return err
						}
						if !matched {
							currentRuleMatch = false
							break
						}
					}

					if currentRuleMatch {
						alreadyMatch = true
					}

				}
			} else {
				alreadyMatch = true
			}

			if alreadyMatch {
				totalCount += 1
				if limit > 0 && page > 0 {
					if totalCount > skip && count < limit {
						count += 1
						res = append(res, data)
					}
				} else {
					count += 1
					res = append(res, data)
				}

			}

		}
		return nil
	})
	if err != nil {
		return nil, 0, 0, err
	}

	return res, count, totalCount, nil
}
