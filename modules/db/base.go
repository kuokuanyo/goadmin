package db

import (
	"database/sql"
	"sync"
)

// Base is a common Connection.
type Base struct {
	DbList map[string]*sql.DB
	Once   sync.Once
}

// Close implements the method Connection.Close.
// 關閉資料庫連線
func (db *Base) Close() []error {
	errs := make([]error, 0)
	for _, d := range db.DbList {
		errs = append(errs, d.Close())
	}
	return errs
}

// GetDB implements the method Connection.GetDB.
// 藉由參數key取得Base.DbList[key]
func (db *Base) GetDB(key string) *sql.DB {
	return db.DbList[key]
}
