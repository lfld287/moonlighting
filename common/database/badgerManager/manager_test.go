package badgerManager

import (
	"go.uber.org/zap/zapcore"
	"moonlighting/common/logger/console"
	"sync"
	"testing"
	"time"
)

var testDataSet = []DataSet{
	{
		Key:   []byte("k1"),
		Value: []byte("v1"),
	},
	{
		Key:   []byte("k2"),
		Value: []byte("v2"),
	},
	{
		Key:   []byte("k3"),
		Value: []byte("v3"),
	},
	{
		Key:   []byte("k4"),
		Value: []byte("v4"),
	},
	{
		Key:   []byte("k5"),
		Value: []byte("v5"),
	},
}

const dbPath = "./testDb"

func TestBadgerManager(t *testing.T) {
	l := console.NewConsoleLogger(zapcore.InfoLevel)
	m := NewBadgerManager(l, dbPath)
	var err error
	go m.Start()
	defer m.Stop()

	err = m.InsertData(testDataSet)
	if err != nil {
		t.Fatal("insert data failed", err)
	}

	err = m.IterateData(func(key []byte, value []byte) {
		t.Logf("key:%s value:%s \n", key, value)
	}, nil)
	if err != nil {
		t.Fatal("iterate data failed", err)
	}

	err = m.DeleteData([][]byte{[]byte("k1"), []byte("k2")})
	if err != nil {
		t.Fatal("iterate data failed", err)
	}

	err = m.IterateData(func(key []byte, value []byte) {
		t.Logf("key:%s value:%s \n", key, value)
	}, nil)
	if err != nil {
		t.Fatal("iterate data failed", err)
	}

	wg := sync.WaitGroup{}

	wg.Add(2)

	go func() {
		defer wg.Done()
		err = m.IterateData(func(key []byte, value []byte) {
			t.Logf("parallel key:%s value:%s \n", key, value)
		}, nil)
		if err != nil {
			t.Error("iterate data failed", err)
			return
		}
	}()

	go func() {
		defer wg.Done()

		err = m.IterateData(func(key []byte, value []byte) {
			t.Logf("parallel key:%s value:%s \n", key, value)
		}, nil)
		if err != nil {
			t.Error("iterate data failed", err)
			return
		}
	}()

	wg.Wait()

	<-time.After(10 * time.Second)

}
