package repository_test

import (
	"sync"

	"github.com/DATA-DOG/go-txdb"
)

var once sync.Once

func RegisterTxDB() {
	once.Do(func() {
		txdb.Register("txdb", "mysql", "root:SENTINELA!@/test_bgw7")
	})
}
