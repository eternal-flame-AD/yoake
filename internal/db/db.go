package db

import (
	"encoding/json"
	"errors"

	badger "github.com/dgraph-io/badger/v3"
	"github.com/eternal-flame-AD/yoake/config"
)

type DB interface {
	NewTransaction(update bool) DBTxn
}

type DBTxn interface {
	Set(key, value []byte) error
	Delete(key []byte) error
	Get(key []byte) ([]byte, error)
	Commit() error
	Discard()
}

func GetJSON(t DBTxn, key []byte, v interface{}) error {
	if data, err := t.Get(key); err != nil {
		return err
	} else {
		return json.Unmarshal(data, v)
	}
}

func SetJSON(t DBTxn, key []byte, v interface{}) error {
	if data, err := json.Marshal(v); err != nil {
		return err
	} else {
		return t.Set(key, data)
	}
}

func New(conf config.C) (DB, error) {
	if conf.DB.Badger.Dir != "" {
		opts := badger.DefaultOptions(conf.DB.Badger.Dir)
		if db, err := badger.Open(opts); err != nil {
			return nil, err
		} else {
			return &BadgerDB{db}, nil
		}
	}
	return nil, errors.New("no database configured")
}
