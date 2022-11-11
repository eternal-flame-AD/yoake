package db

import (
	"errors"

	"github.com/dgraph-io/badger/v3"
)

func IsNotFound(err error) bool {
	if errors.Is(err, badger.ErrKeyNotFound) {
		return true
	}
	return false
}
