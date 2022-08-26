package dataManager

import (
	"go.uber.org/zap/zapcore"
	"moonlighting/common/database/badgerManager"
	"moonlighting/common/logger/console"
	"testing"
	"time"
)

func TestManager(t *testing.T) {
	l := console.NewConsoleLogger(zapcore.InfoLevel)
	const dbPath = "./ttt2223"
	m := badgerManager.NewBadgerManager(l, dbPath)
	go m.Start()
	defer m.Stop()

	testDataManager := NewDataManager(l, "test.", m)
	go testDataManager.Start()
	defer testDataManager.Stop()

	<-time.After(2 * time.Second)

	err := testDataManager.InsertData([]Data{
		{
			Key: "key1",
			Value: map[string]string{
				"eee1": "fff1",
			},
			Priority: 0,
		},
		{
			Key: "key2",
			Value: map[string]string{
				"eee2": "fff1",
			},
			Priority: 56,
		},
		{
			Key: "key3",
			Value: map[string]string{
				"eee3": "fff1",
			},
			Priority: 22,
		},
		{
			Key: "key4",
			Value: map[string]string{
				"eee4": "fff1",
			},
			Priority: 42,
		},
	})

	if err != nil {
		t.Fatal(err)
	}

	res, count, totalCount, err := testDataManager.QueryData(2, 1, map[string]string{})
	t.Log(res, count, totalCount, err)
}
