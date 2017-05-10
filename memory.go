package mmbot

import (
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
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

func (m *Memory) Get(plugin Plugin, key string) (string, error) {
	ns_key := fmt.Sprintf("%s:%s", m.namespace(plugin), key)
	if val, err := m.db.Get([]byte(ns_key), nil); err != nil {
		return "", err
	} else {
		return string(val), nil
	}
}

func (m *Memory) Put(plugin Plugin, key string, val string) error {
	ns_key := fmt.Sprintf("%s:%s", m.namespace(plugin), key)
	if err := m.db.Put([]byte(ns_key), []byte(val), nil); err != nil {
		return err
	} else {
		return nil
	}
}

func (m *Memory) Del(plugin Plugin, key string) (string, error) {
	var val string
	if tmp, err := m.Get(plugin, key); err != nil {
		return "", err
	} else {
		val = tmp
	}

	ns_key := fmt.Sprintf("%s:%s", m.namespace(plugin), key)
	if err := m.db.Delete([]byte(ns_key), nil); err != nil {
		return "", err
	} else {
		return val, nil
	}
}

func (m *Memory) List(plugin Plugin) (map[string]string, error) {
	list := map[string]string{}

	ns_prefix := fmt.Sprintf("%s:", m.namespace(plugin))
	iter := m.db.NewIterator(util.BytesPrefix([]byte(ns_prefix)), nil)
	for iter.Next() {
		ns_key := iter.Key()
		key := ns_key[len(ns_prefix):]
		val := iter.Value()
		list[string(key)] = string(val)
	}

	iter.Release()
	if err := iter.Error(); err != nil {
		return nil, err
	} else {
		return list, nil
	}
}

func (m *Memory) namespace(plugin Plugin) string {
	namespace := reflect.TypeOf(plugin).String()
	return strings.ToUpper(namespace)
}
