package mmbot

import (
	"fmt"
	"os"
	"reflect"

	"github.com/syndtr/goleveldb/leveldb"
)

const (
	LEVELDB_PATH = "mmbot.ldb"
)

type Memory struct {
	db *leveldb.DB
}

func NewMemory() (*Memory, error) {
	storage := os.Getenv("LEVELDB_PATH")
	if storage == "" {
		storage = LEVELDB_PATH
	}

	db, err := leveldb.OpenFile(storage, nil)
	if err != nil {
		return nil, err
	}

	return &Memory{db}, nil
}

func (m *Memory) Get(plugin BotPlugin, key string) (string, error) {
	pluginType := reflect.TypeOf(plugin)

	ns_key := fmt.Sprintf("%s:%s:%s", pluginType.PkgPath, pluginType.Name, key)
	if val, err := m.db.Get([]byte(ns_key), nil); err != nil {
		return "", err
	} else {
		return string(val), nil
	}
}

func (m *Memory) Put(plugin BotPlugin, key string, val string) error {
	pluginType := reflect.TypeOf(plugin)

	ns_key := fmt.Sprintf("%s:%s:%s", pluginType.PkgPath, pluginType.Name, key)
	if err := m.db.Put([]byte(ns_key), []byte(val), nil); err != nil {
		return err
	} else {
		return nil
	}
}
