package botkit

import (
	"fmt"
	"reflect"

	"github.com/syndtr/goleveldb/leveldb"
)

const LEVELDB_PATH = "mmbot.ldb"

type Brain struct {
	db *leveldb.DB
}

func NewBrain() (*Brain, error) {
	db, err := leveldb.OpenFile(LEVELDB_PATH, nil)
	if err != nil {
		return nil, err
	}

	b := new(Brain)
	b.db = db

	return b, nil
}

func (b *Brain) Get(adaptor MMBotInterface, key string) (string, error) {
	t := reflect.TypeOf(adaptor)
	k := fmt.Sprintf("%s:%s:%s", t.PkgPath, t.Name, key)

	if val, err := b.db.Get([]byte(k), nil); err != nil {
		return "", err
	} else {
		return string(val), nil
	}
}

func (b *Brain) Put(adaptor MMBotInterface, key string, val string) error {
	t := reflect.TypeOf(adaptor)
	k := fmt.Sprintf("%s:%s:%s", t.PkgPath, t.Name, key)

	if err := b.db.Put([]byte(k), []byte(val), nil); err != nil {
		return err
	} else {
		return nil
	}
}
