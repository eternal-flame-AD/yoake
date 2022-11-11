package db

import "github.com/dgraph-io/badger/v3"

type BadgerDB struct {
	db *badger.DB
}

func (b *BadgerDB) NewTransaction(readonly bool) DBTxn {
	return &BadgerDBTxn{
		txn: b.db.NewTransaction(readonly),
	}
}

type BadgerDBTxn struct {
	txn *badger.Txn
}

func (t *BadgerDBTxn) Set(key []byte, value []byte) error {
	return t.txn.Set(key, value)
}

func (t *BadgerDBTxn) Delete(key []byte) error {
	return t.txn.Delete(key)
}

func (t *BadgerDBTxn) Get(key []byte) ([]byte, error) {
	item, err := t.txn.Get(key)
	if err != nil {
		return nil, err
	}
	return item.ValueCopy(nil)
}

func (t *BadgerDBTxn) Commit() error {
	return t.txn.Commit()
}

func (t *BadgerDBTxn) Discard() {
	t.txn.Discard()
}
